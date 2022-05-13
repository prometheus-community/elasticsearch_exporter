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
	"net/http"
	"net/url"
	"path"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultNodeShardLabels = []string{"node"}

	defaultNodeShardLabelValues = func(node string) []string {
		return []string{
			node,
		}
	}
)

// ShardResponse has shard's node and index info
type ShardResponse struct {
	Index string `json:"index"`
	Shard string `json:"shard"`
	Node  string `json:"node"`
}

// Shards information struct
type Shards struct {
	logger log.Logger
	client *http.Client
	url    *url.URL

	nodeShardMetrics  []*nodeShardMetric
	jsonParseFailures prometheus.Counter
}

type nodeShardMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(shards float64) float64
	Labels func(node string) []string
}

// NewShards defines Shards Prometheus metrics
func NewShards(logger log.Logger, client *http.Client, url *url.URL) *Shards {
	return &Shards{
		logger: logger,
		client: client,
		url:    url,

		nodeShardMetrics: []*nodeShardMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "node_shards", "total"),
					"Total shards per node",
					defaultNodeShardLabels, nil,
				),
				Value: func(shards float64) float64 {
					return shards
				},
				Labels: defaultNodeShardLabelValues,
			}},

		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "node_shards", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
	}
}

// Describe Shards
func (s *Shards) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.jsonParseFailures.Desc()

	for _, metric := range s.nodeShardMetrics {
		ch <- metric.Desc
	}
}

func (s *Shards) getAndParseURL(u *url.URL) ([]ShardResponse, error) {
	res, err := s.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(s.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}
	var sfr []ShardResponse
	if err := json.NewDecoder(res.Body).Decode(&sfr); err != nil {
		s.jsonParseFailures.Inc()
		return nil, err
	}
	return sfr, nil
}

func (s *Shards) fetchAndDecodeShards() ([]ShardResponse, error) {

	u := *s.url
	u.Path = path.Join(u.Path, "/_cat/shards")
	q := u.Query()
	q.Set("format", "json")
	u.RawQuery = q.Encode()
	sfr, err := s.getAndParseURL(&u)
	if err != nil {
		return sfr, err
	}
	return sfr, err
}

// Collect number of shards on each nodes
func (s *Shards) Collect(ch chan<- prometheus.Metric) {

	defer func() {
		ch <- s.jsonParseFailures
	}()

	sr, err := s.fetchAndDecodeShards()
	if err != nil {
		_ = level.Warn(s.logger).Log(
			"msg", "failed to fetch and decode node shards stats",
			"err", err,
		)
		return
	}

	nodeShards := make(map[string]float64)

	for _, shard := range sr {
		if val, ok := nodeShards[shard.Node]; ok {
			nodeShards[shard.Node] = val + 1
		} else {
			nodeShards[shard.Node] = 1
		}
	}

	for node, shards := range nodeShards {
		for _, metric := range s.nodeShardMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(shards),
				metric.Labels(node)...,
			)
		}
	}
}
