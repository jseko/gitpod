// Copyright (c) 2023 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package analysis

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

const (
	// Rate of increase in error count
	errorMetric = "rate(gitpod_ws_manager_workspace_starts_failure_total{cluster=%s})"
)

type PrometheusAnalyzer struct {
	prometheusURL string
	startTime     time.Time
}

func NewPrometheusAnalyzer(promURL string) *PrometheusAnalyzer {
	return &PrometheusAnalyzer{
		prometheusURL: promURL,
		startTime:     time.Now(),
	}
}

func (pa *PrometheusAnalyzer) MoveForward(ctx context.Context, clusterName string) (bool, error) {
	client, err := api.NewClient(api.Config{
		Address: pa.prometheusURL,
	})
	if err != nil {
		return false, err
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, warnings, err := v1api.QueryRange(ctx, fmt.Sprintf(errorMetric, clusterName), v1.Range{
		Start: pa.startTime,
		End:   time.Now(),
	})
	if err != nil {
		return false, err
	}
	if len(warnings) > 0 {
		log.Warnf("Warnings: %v\n", warnings)
	}

	val := float64(result.(*model.Scalar).Value)
	if math.IsNaN(val) {
		return false, errors.New("query result value is not-a-number")
	}

	// Return true if the error rate is 0
	if int64(val) == 0 {
		return true, nil
	}

	return false, nil

}
