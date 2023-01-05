// Copyright (c) 2023 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package cmd

import (
	"context"
	"fmt"
	"log"
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
			log.Fatal("Could not create temporary directory")
			return
		}
		defer os.RemoveAll(tmpDir)

		wsInfo, err := client.Info.WorkspaceInfo(ctx, &api.WorkspaceInfoRequest{})
		if err != nil {
			utils.LogError(ctx, err, "Could not fetch the workspace info", client)
			return
		}

		ctx = context.Background()
		gitpodConfig, err := util.ParseGitpodConfig(wsInfo.CheckoutLocation)
		if err != nil {
			log.Fatal("Could not parse gitpod config")
			return
		}

		if gitpodConfig == nil {
			fmt.Println("To test the image build, you need to configure your project with a .gitpod.yml file")
			fmt.Println("")
			fmt.Println("For a quick start, try running:\n$ gp init -i")
			fmt.Println("")
			fmt.Println("Alternatively, check out the following docs for getting started configuring your project")
			fmt.Println("https://www.gitpod.io/docs/configure#configure-gitpod")
			return
		}

		var baseimage string
		switch img := gitpodConfig.Image.(type) {
		case nil:
			baseimage = ""
		case string:
			baseimage = "FROM " + img
		case map[interface{}]interface{}:
			dockerfilePath := filepath.Join(wsInfo.CheckoutLocation, img["file"].(string))

			if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
				fmt.Println("Your .gitpod.yml points to a Dockerfile that doesn't exist: " + dockerfilePath)
				return
			}
			dockerfile, err := os.ReadFile(dockerfilePath)
			if err != nil {
				log.Fatal("Could not read the Dockerfile")
				return
			}
			if string(dockerfile) == "" {
				fmt.Println("Your Gitpod's Dockerfile is empty")
				fmt.Println("")
				fmt.Println("To learn how to customize your workspace, check out the following docs:")
				fmt.Println("https://www.gitpod.io/docs/configure/workspaces/workspace-image#use-a-custom-dockerfile")
				fmt.Println("")
				fmt.Println("Once you configure your Dockerfile, re-run this command to validate your changes")
				return
			}
			baseimage = "\n" + string(dockerfile) + "\n"
		default:
			fmt.Println("Check your .gitpod.yml and make sure the image property is configured correctly")
			return
		}

		if baseimage == "" {
			fmt.Println("Your project is not using any custom Docker image.")                                        // todo: cleanup
			fmt.Println("Check out the following docs, to know how to get started")                                  // todo: cleanup
			fmt.Println("")                                                                                          // todo: cleanup
			fmt.Println("https://www.gitpod.io/docs/configure/workspaces/workspace-image#use-a-public-docker-image") // todo: cleanup
			return
		}

		err = os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte(baseimage), 0644)
		if err != nil {
			fmt.Println("Could not write the temporary Dockerfile")
			log.Fatal(err)
			return
		}

		tag := "temp-build-" + time.Now().Format("20060102150405")

		dockerCmd := exec.Command("docker", "build", "-t", tag, "--progress=tty", ".")
		dockerCmd.Dir = tmpDir

		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		go func() {
			<-ctx.Done()
			if proc := dockerCmd.Process; proc != nil {
				_ = proc.Kill()
			}
		}()

		// TODO: duration
		err = dockerCmd.Run()
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Println("Image Build Failed")
			log.Fatal(err)
			return
		} else if err != nil {
			fmt.Println("Docker error")
			log.Fatal(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
