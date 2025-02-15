// Copyright 2023 The Prometheus Authors
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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ilmStatuses = []string{"STOPPED", "RUNNING", "STOPPING"}
)

type ilmStatusMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(ilm *IlmStatusResponse, status string) float64
	Labels func(status string) []string
}

// IlmStatusCollector information struct
type IlmStatusCollector struct {
	logger *slog.Logger
	client *http.Client
	url    *url.URL

	metric ilmStatusMetric
}

type IlmStatusResponse struct {
	OperationMode string `json:"operation_mode"`
}

// NewIlmStatus defines Indices IndexIlms Prometheus metrics
func NewIlmStatus(logger *slog.Logger, client *http.Client, url *url.URL) *IlmStatusCollector {
	subsystem := "ilm"

	return &IlmStatusCollector{
		logger: logger,
		client: client,
		url:    url,

		metric: ilmStatusMetric{
			Type: prometheus.GaugeValue,
			Desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem, "status"),
				"Current status of ILM. Status can be STOPPED, RUNNING, STOPPING.",
				[]string{"operation_mode"}, nil,
			),
			Value: func(ilm *IlmStatusResponse, status string) float64 {
				if ilm.OperationMode == status {
					return 1
				}
				return 0
			},
		},
	}
}

// Describe add Snapshots metrics descriptions
func (im *IlmStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- im.metric.Desc
}

func (im *IlmStatusCollector) fetchAndDecodeIlm() (*IlmStatusResponse, error) {
	u := *im.url
	u.Path = path.Join(im.url.Path, "/_ilm/status")

	res, err := im.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		im.logger.Warn("failed to read response body", "err", err)
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		im.logger.Warn("failed to close response body", "err", err)
		return nil, err
	}

	var imr IlmStatusResponse
	if err := json.Unmarshal(body, &imr); err != nil {
		return nil, err
	}

	return &imr, nil
}

// Collect gets all indices Ilms metric values
func (im *IlmStatusCollector) Collect(ch chan<- prometheus.Metric) {
	indicesIlmsResponse, err := im.fetchAndDecodeIlm()
	if err != nil {
		im.logger.Warn(
			"failed to fetch and decode cluster ilm status",
			"err", err,
		)
		return
	}

	for _, status := range ilmStatuses {
		ch <- prometheus.MustNewConstMetric(
			im.metric.Desc,
			im.metric.Type,
			im.metric.Value(indicesIlmsResponse, status),
			status,
		)
	}

}
