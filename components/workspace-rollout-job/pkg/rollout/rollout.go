// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package rollout

import (
	"context"
	"time"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/gitpod-io/gitpod/workspace-rollout-job/pkg/analysis"
	"github.com/gitpod-io/gitpod/workspace-rollout-job/pkg/wsbridge"
)

type RollOutJob struct {
	oldCluster           string
	newCluster           string
	currentScore         int32
	analyzer             analysis.Analyzer
	RolloutAction        wsbridge.RolloutAction
	rolloutStep          int32
	analysisWaitDuration time.Duration

	ticker *time.Ticker
	revert chan bool
	done   chan bool
}

func New(oldCluster, newCluster string, rolloutWaitDuration, analysisWaitDuration time.Duration, step int32, analyzer analysis.Analyzer, rolloutAction wsbridge.RolloutAction) *RollOutJob {
	return &RollOutJob{
		oldCluster:           oldCluster,
		newCluster:           newCluster,
		currentScore:         0,
		rolloutStep:          step,
		analyzer:             analyzer,
		RolloutAction:        rolloutAction,
		done:                 make(chan bool),
		revert:               make(chan bool),
		analysisWaitDuration: analysisWaitDuration,
		// move forward every waitDuration
		ticker: time.NewTicker(rolloutWaitDuration),
	}
}

// Start runs the job synchronously
func (r *RollOutJob) Start(ctx context.Context) {
	// keep checking the analyzer asynchronously to see if there is a
	// problem with the new cluster
	go func() {
		for {
			// check every analysisWaitDuration
			time.Sleep(r.analysisWaitDuration)
			moveForward, err := r.analyzer.MoveForward(context.Background(), r.newCluster)
			if err != nil {
				log.Error("Failed to retrieve new cluster error count: ", err)
			}
			// Analyzer says no, stop the rollout
			if !moveForward {
				close(r.revert)
			}
		}
	}()

	func() {
		for {
			select {
			case <-r.ticker.C:
				if r.currentScore == 100 {
					r.Stop()
					return
				}
				r.currentScore += r.rolloutStep
				if err := r.RolloutAction.UpdateScore(ctx, r.newCluster, r.currentScore); err != nil {
					log.Error("Failed to update new cluster score: ", err)
				}
				if err := r.RolloutAction.UpdateScore(ctx, r.oldCluster, 100-r.currentScore); err != nil {
					log.Error("Failed to update old cluster score: ", err)
				}

				log.Infof("Updated scores as %s:%d, %s:%d", r.newCluster, r.currentScore, r.oldCluster, 100-r.currentScore)
			case <-r.revert:
				log.Info("Reverting the rollout")
				if err := r.RolloutAction.UpdateScore(ctx, r.newCluster, 0); err != nil {
					log.Error("Failed to update new cluster score: ", err)
				}

				if err := r.RolloutAction.UpdateScore(ctx, r.oldCluster, 100); err != nil {
					log.Error("Failed to update old cluster score: ", err)
				}

			case <-r.done:
				return
			}
		}
	}()
}

func (r *RollOutJob) Stop() {
	close(r.done)
	r.ticker.Stop()
}
