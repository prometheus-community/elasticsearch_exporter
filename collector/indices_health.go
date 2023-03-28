// Copyright 2021 The Prometheus Authors
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
	"github.com/prometheus-community/elasticsearch_exporter/pkg/clusterinfo"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	indexColors = []string{"green", "yellow", "red"}
)

type indicesHealthLabels struct {
	keys   func(...string) []string
	values func(*clusterinfo.Response, ...string) []string
}

type indexHealthMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(indexHealth indexHealthResponse, color string) float64
	Labels indicesHealthLabels
}

// IndiceHealth type defines the collector struct
type IndicesHealth struct {
	logger          log.Logger
	client          *http.Client
	url             *url.URL
	clusterInfoCh   chan *clusterinfo.Response
	lastClusterInfo *clusterinfo.Response

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter

	indexesHealthMetrics []*indexHealthMetric
}

// NewIndicesHealth defines IndicesHealth metrics
func NewIndicesHealth(logger log.Logger, client *http.Client, url *url.URL) *IndicesHealth {
	subsystem := "indices_health"

	indexLabels := indicesHealthLabels{
		keys: func(...string) []string {
			return []string{"index", "color", "cluster"}
		},
		values: func(lastClusterinfo *clusterinfo.Response, s ...string) []string {
			if lastClusterinfo != nil {
				return append(s, lastClusterinfo.ClusterName)
			}
			// this shouldn't happen, as the clusterinfo Retriever has a blocking
			// Run method. It blocks until the first clusterinfo call has succeeded
			return append(s, "unknown_cluster")
		},
	}

	indicesHealth := &IndicesHealth{
		logger:        logger,
		client:        client,
		url:           url,
		clusterInfoCh: make(chan *clusterinfo.Response),
		lastClusterInfo: &clusterinfo.Response{
			ClusterName: "unknown_cluster",
		},

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "up"),
			Help: "Was the last scrape of the Elasticsearch cat indices endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "total_scrapes"),
			Help: "Current total Elasticsearch cat indices scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),

		indexesHealthMetrics: []*indexHealthMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "health"),
					"Whether all primary and replica index shards are allocated.",
					indexLabels.keys(), nil,
				),
				Value: func(indexHealth indexHealthResponse, color string) float64 {
					if indexHealth.Health == color {
						return 1
					}
					return 0
				},
				Labels: indexLabels,
			},
		},
	}

	// start go routine to fetch clusterinfo updates and save them to lastClusterinfo
	go func() {
		_ = level.Debug(logger).Log("msg", "starting cluster info receive loop")
		for ci := range indicesHealth.clusterInfoCh {
			if ci != nil {
				_ = level.Debug(logger).Log("msg", "received cluster info update", "cluster", ci.ClusterName)
				indicesHealth.lastClusterInfo = ci
			}
		}
		_ = level.Debug(logger).Log("msg", "exiting cluster info receive loop")
	}()

	return indicesHealth
}

// Describe add IndicesHealth metrics descriptions
func (ih *IndicesHealth) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range ih.indexesHealthMetrics {
		ch <- metric.Desc
	}
	ch <- ih.up.Desc()
	ch <- ih.totalScrapes.Desc()
	ch <- ih.jsonParseFailures.Desc()
}

// ClusterLabelUpdates returns a pointer to a channel to receive cluster info updates. It implements the
// (not exported) clusterinfo.consumer interface
func (ih *IndicesHealth) ClusterLabelUpdates() *chan *clusterinfo.Response {
	return &ih.clusterInfoCh
}

// String implements the stringer interface. It is part of the clusterinfo.consumer interface
func (ih *IndicesHealth) String() string {
	return namespace + "indiceshealth"
}

func (ih *IndicesHealth) queryURL(u *url.URL) ([]byte, error) {
	res, err := ih.client.Get(u.String())
	if err != nil {
		return []byte{}, fmt.Errorf("failed to get resource from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(ih.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}

	return bts, nil
}

func (ih *IndicesHealth) fetchAndDecodeIndicesHealth() (CatIndicesResponse, error) {
	var isr CatIndicesResponse

	u := *ih.url
	u.Path = path.Join(u.Path, "/_cat/indices")
	u.RawQuery = "format=json&h=health,index"

	bts, err := ih.queryURL(&u)
	if err != nil {
		return isr, err
	}

	if err := json.Unmarshal(bts, &isr); err != nil {
		ih.jsonParseFailures.Inc()
		return isr, err
	}

	return isr, nil
}

// Collect gets indices health metric values
func (ih *IndicesHealth) Collect(ch chan<- prometheus.Metric) {
	ih.totalScrapes.Inc()
	defer func() {
		ch <- ih.up
		ch <- ih.totalScrapes
		ch <- ih.jsonParseFailures
	}()

	catIndicesResponse, err := ih.fetchAndDecodeIndicesHealth()
	if err != nil {
		ih.up.Set(0)
		_ = level.Warn(ih.logger).Log(
			"msg", "failed to fetch and decode cat indices",
			"err", err,
		)
		return
	}
	ih.up.Set(1)

	for _, metric := range ih.indexesHealthMetrics {
		for _, indexHealth := range catIndicesResponse {
			for _, color := range indexColors {
				ch <- prometheus.MustNewConstMetric(
					metric.Desc,
					metric.Type,
					metric.Value(indexHealth, color),
					metric.Labels.values(ih.lastClusterInfo, indexHealth.Index, color)...,
				)
			}
		}
	}
}
