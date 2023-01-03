// Copyright (c) 2023 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package rollout

import (
	"context"
	"testing"
	"time"
)

type MockAnalyzer struct {
}

func (m *MockAnalyzer) MoveForward(ctx context.Context, clusterName string) (bool, error) {
	return true, nil
}

func TestSuccessfulRollout(t *testing.T) {
	rolloutJob := New("ws-1", "ws-2", 1*time.Second, 1*time.Second, 25, &MockAnalyzer{})
	rolloutJob.Start()
}

type FailureMockAnalyzer struct {
}

func (m *FailureMockAnalyzer) MoveForward(ctx context.Context, clusterName string) (bool, error) {
	return false, nil
}

func TestFailureRollout(t *testing.T) {
	rolloutJob := New("ws-1", "ws-2", 1*time.Second, 1*time.Second, 25, &FailureMockAnalyzer{})
	rolloutJob.Start()
}
