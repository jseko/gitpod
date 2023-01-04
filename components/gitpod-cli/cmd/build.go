// Copyright (c) 2023 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gitpod-io/gitpod/common-go/util"
	"github.com/gitpod-io/gitpod/gitpod-cli/pkg/supervisor"
	"github.com/gitpod-io/gitpod/gitpod-cli/pkg/utils"
	"github.com/gitpod-io/gitpod/supervisor/api"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:    "build",
	Short:  "Builds the workspace image (useful to debug a workspace custom image)",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err := supervisor.New(ctx)
		if err != nil {
			utils.LogError(ctx, err, "Could not get workspace info required to build", client)
			return
		}
		defer client.Close()

		tmpDir, err := os.MkdirTemp("", "gp-build-*")
		if err != nil {
			utils.LogError(ctx, err, "Could not create temporary directory", client)
			return
		}
		defer os.RemoveAll(tmpDir)

		wsInfo, err := client.Info.WorkspaceInfo(ctx, &api.WorkspaceInfoRequest{})
		if err != nil {
			utils.LogError(ctx, err, "Could not fetch the workspace info", client)
			return
		}

		gitpodConfig, err := util.ParseGitpodConfig(wsInfo.CheckoutLocation)

		var baseimage string
		switch img := gitpodConfig.Image.(type) {
		case nil:
			baseimage = "FROM gitpod/workspace-full:latest"
		case string:
			baseimage = "FROM " + img
		case map[interface{}]interface{}:
			dockerfilePath := img["file"].(string)
			dockerfile, err := ioutil.ReadFile(filepath.Join(wsInfo.CheckoutLocation, dockerfilePath))
			if err != nil {
				utils.LogError(ctx, err, "Could not read the Dockerfile", client)
				return
			}
			baseimage = "\n" + string(dockerfile) + "\n"
		default:
			utils.LogError(ctx, err, "unsupported image: "+img.(string), client)
			return
		}

		// fmt.Println("baseimage: " + baseimage)

		tag := "temp-build-" + time.Now().Format("20060102150405")

		dockerCmd := exec.Command("docker", "build", "-t", tag, ".")
		dockerCmd.Dir = tmpDir

		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		err = ioutil.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte(baseimage), 0644)
		if err != nil {
			utils.LogError(ctx, err, "Could not write the temporary Dockerfile", client)
			return
		}

		go func() {
			<-ctx.Done()
			if proc := dockerCmd.Process; proc != nil {
				_ = proc.Kill()
			}
		}()

		err = dockerCmd.Run()
		if _, ok := err.(*exec.ExitError); ok {
			utils.LogError(ctx, err, "Workspace image build failed", client)
			return
		} else if err != nil {
			utils.LogError(ctx, err, "Docker error", client)
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
