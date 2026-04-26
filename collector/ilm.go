// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ilmStatusOptions = []string{"STOPPED", "RUNNING", "STOPPING"}

	ilmIndexStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ilm_index", "status"),
		"Status of ILM policy for index",
		[]string{"index", "phase", "action", "step"}, nil)

	ilmStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ilm", "status"),
		"Current status of ILM. Status can be STOPPED, RUNNING, STOPPING.",
		[]string{"operation_mode"}, nil,
	)
)

func init() {
	registerCollector("ilm", defaultDisabled, NewILM)
}

type ILM struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

func NewILM(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &ILM{
		logger: logger,
		hc:     hc,
		u:      u,
	}, nil
}

type IlmResponse struct {
	Indices map[string]IlmIndexResponse `json:"indices"`
}

type IlmIndexResponse struct {
	Index          string  `json:"index"`
	Managed        bool    `json:"managed"`
	Phase          string  `json:"phase"`
	Action         string  `json:"action"`
	Step           string  `json:"step"`
	StepTimeMillis float64 `json:"step_time_millis"`
}

type IlmStatusResponse struct {
	OperationMode string `json:"operation_mode"`
}

func (i *ILM) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	var ir IlmResponse

	indexURL := i.u.ResolveReference(&url.URL{Path: "/_all/_ilm/explain"})

	if err := getAndDecodeURL(ctx, i.hc, i.logger, indexURL.String(), &ir); err != nil {
		return fmt.Errorf("failed to load ILM index explain: %w", err)
	}

	var isr IlmStatusResponse

	indexStatusURL := i.u.ResolveReference(&url.URL{Path: "/_ilm/status"})

	if err := getAndDecodeURL(ctx, i.hc, i.logger, indexStatusURL.String(), &isr); err != nil {
		return fmt.Errorf("failed to load ILM status: %w", err)
	}

	for name, ilm := range ir.Indices {
		ch <- prometheus.MustNewConstMetric(
			ilmIndexStatus,
			prometheus.GaugeValue,
			bool2Float(ilm.Managed),
			name, ilm.Phase, ilm.Action, ilm.Step,
		)
	}

	for _, status := range ilmStatusOptions {
		statusActive := false
		if isr.OperationMode == status {
			statusActive = true
		}

		ch <- prometheus.MustNewConstMetric(
			ilmStatus,
			prometheus.GaugeValue,
			bool2Float(statusActive),
			status,
		)
	}

	return nil
}
