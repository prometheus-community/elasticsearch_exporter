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
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
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
	logger log.Logger
	client *http.Client
	url    *url.URL

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter

	dataStreamMetrics []*dataStreamMetric
}

// NewDataStream defines DataStream Prometheus metrics
func NewDataStream(logger log.Logger, client *http.Client, url *url.URL) *DataStream {
	return &DataStream{
		logger: logger,
		client: client,
		url:    url,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "data_stream_stats", "up"),
			Help: "Was the last scrape of the ElasticSearch Data Stream stats endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "data_stream_stats", "total_scrapes"),
			Help: "Current total ElasticSearch Data STream scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "data_stream_stats", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
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

	ch <- ds.up.Desc()
	ch <- ds.totalScrapes.Desc()
	ch <- ds.jsonParseFailures.Desc()
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
			_ = level.Warn(ds.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return dsr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ds.jsonParseFailures.Inc()
		return dsr, err
	}

	if err := json.Unmarshal(bts, &dsr); err != nil {
		ds.jsonParseFailures.Inc()
		return dsr, err
	}

	return dsr, nil
}

// Collect gets DataStream metric values
func (ds *DataStream) Collect(ch chan<- prometheus.Metric) {
	ds.totalScrapes.Inc()
	defer func() {
		ch <- ds.up
		ch <- ds.totalScrapes
		ch <- ds.jsonParseFailures
	}()

	dataStreamStatsResp, err := ds.fetchAndDecodeDataStreamStats()
	if err != nil {
		ds.up.Set(0)
		_ = level.Warn(ds.logger).Log(
			"msg", "failed to fetch and decode data stream stats",
			"err", err,
		)
		return
	}

	ds.up.Set(1)

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
