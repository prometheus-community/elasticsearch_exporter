// Copyright 2022 The Prometheus Authors
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

type dataStreamMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(dataStreamStats DataStreamStatsDataStream) float64
	Labels func(dataStreamStats DataStreamStatsDataStream) []string
}

var (
	defaultDataStreamLabels      = []string{"data_stream"}
	defaultDataStreamLabelValues = func(dataStreamStats DataStreamStatsDataStream) []string {
		return []string{dataStreamStats.DataStream}
	}
)

// DataStream Information Struct
type DataStream struct {
	logger *slog.Logger
	client *http.Client
	url    *url.URL

	dataStreamMetrics []*dataStreamMetric
}

// NewDataStream defines DataStream Prometheus metrics
func NewDataStream(logger *slog.Logger, client *http.Client, url *url.URL) *DataStream {
	return &DataStream{
		logger: logger,
		client: client,
		url:    url,

		dataStreamMetrics: []*dataStreamMetric{
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "data_stream", "backing_indices_total"),
					"Number of backing indices",
					defaultDataStreamLabels, nil,
				),
				Value: func(dataStreamStats DataStreamStatsDataStream) float64 {
					return float64(dataStreamStats.BackingIndices)
				},
				Labels: defaultDataStreamLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "data_stream", "store_size_bytes"),
					"Store size of data stream",
					defaultDataStreamLabels, nil,
				),
				Value: func(dataStreamStats DataStreamStatsDataStream) float64 {
					return float64(dataStreamStats.StoreSizeBytes)
				},
				Labels: defaultDataStreamLabelValues,
			},
		},
	}
}

// Describe adds DataStream metrics descriptions
func (ds *DataStream) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range ds.dataStreamMetrics {
		ch <- metric.Desc
	}
}

func (ds *DataStream) fetchAndDecodeDataStreamStats() (DataStreamStatsResponse, error) {
	var dsr DataStreamStatsResponse

	u := *ds.url
	u.Path = path.Join(u.Path, "/_data_stream/*/_stats")
	res, err := ds.client.Get(u.String())
	if err != nil {
		return dsr, fmt.Errorf("failed to get data stream stats health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			ds.logger.Warn(
				"failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return dsr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := io.ReadAll(res.Body)
	if err != nil {
		return dsr, err
	}

	if err := json.Unmarshal(bts, &dsr); err != nil {
		return dsr, err
	}

	return dsr, nil
}

// Collect gets DataStream metric values
func (ds *DataStream) Collect(ch chan<- prometheus.Metric) {

	dataStreamStatsResp, err := ds.fetchAndDecodeDataStreamStats()
	if err != nil {
		ds.logger.Warn(
			"failed to fetch and decode data stream stats",
			"err", err,
		)
		return
	}

	for _, metric := range ds.dataStreamMetrics {
		for _, dataStream := range dataStreamStatsResp.DataStreamStats {
			fmt.Printf("Metric: %+v", dataStream)
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(dataStream),
				metric.Labels(dataStream)...,
			)
		}
	}
}
