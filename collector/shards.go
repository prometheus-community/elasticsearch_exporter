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
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus-community/elasticsearch_exporter/pkg/clusterinfo"

	"github.com/prometheus/client_golang/prometheus"
)

// ShardResponse has shard's node and index info
type ShardResponse struct {
	Index string `json:"index"`
	Shard string `json:"shard"`
	State string `json:"state"`
	Node  string `json:"node"`
}

// Shards information struct
type Shards struct {
	logger          *slog.Logger
	client          *http.Client
	url             *url.URL
	clusterInfoCh   chan *clusterinfo.Response
	lastClusterInfo *clusterinfo.Response

	nodeShardMetrics  []*nodeShardMetric
	jsonParseFailures prometheus.Counter
}

// ClusterLabelUpdates returns a pointer to a channel to receive cluster info updates. It implements the
// (not exported) clusterinfo.consumer interface
func (s *Shards) ClusterLabelUpdates() *chan *clusterinfo.Response {
	return &s.clusterInfoCh
}

// String implements the stringer interface. It is part of the clusterinfo.consumer interface
func (s *Shards) String() string {
	return namespace + "shards"
}

type nodeShardMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(shards float64) float64
	Labels labels
}

// NewShards defines Shards Prometheus metrics
func NewShards(logger *slog.Logger, client *http.Client, url *url.URL) *Shards {

	nodeLabels := labels{
		keys: func(...string) []string {
			return []string{"node", "cluster"}
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

	shards := &Shards{
		logger: logger,
		client: client,
		url:    url,

		clusterInfoCh: make(chan *clusterinfo.Response),
		lastClusterInfo: &clusterinfo.Response{
			ClusterName: "unknown_cluster",
		},

		nodeShardMetrics: []*nodeShardMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "node_shards", "total"),
					"Total shards per node",
					nodeLabels.keys(), nil,
				),
				Value: func(shards float64) float64 {
					return shards
				},
				Labels: nodeLabels,
			}},

		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "node_shards", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
	}

	// start go routine to fetch clusterinfo updates and save them to lastClusterinfo
	go func() {
		logger.Debug("starting cluster info receive loop")
		for ci := range shards.clusterInfoCh {
			if ci != nil {
				logger.Debug("received cluster info update", "cluster", ci.ClusterName)
				shards.lastClusterInfo = ci
			}
		}
		logger.Debug("exiting cluster info receive loop")
	}()

	return shards
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
			s.logger.Warn(
				"failed to close http.Client",
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

// Collect number of shards on each node
func (s *Shards) Collect(ch chan<- prometheus.Metric) {

	defer func() {
		ch <- s.jsonParseFailures
	}()

	sr, err := s.fetchAndDecodeShards()
	if err != nil {
		s.logger.Warn(
			"failed to fetch and decode node shards stats",
			"err", err,
		)
		return
	}

	nodeShards := make(map[string]float64)

	for _, shard := range sr {
		if shard.State == "STARTED" {
			nodeShards[shard.Node]++
		}
	}

	for node, shards := range nodeShards {
		for _, metric := range s.nodeShardMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(shards),
				metric.Labels.values(s.lastClusterInfo, node)...,
			)
		}
	}
}
