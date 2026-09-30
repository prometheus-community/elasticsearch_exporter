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
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/elasticsearch_exporter/pkg/clusterinfo"
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

// fetchClusterNameOnce performs a single request to the root endpoint to obtain the cluster name.
func fetchClusterNameOnce(s *Shards) string {
	if s.lastClusterInfo != nil && s.lastClusterInfo.ClusterName != "unknown_cluster" {
		return s.lastClusterInfo.ClusterName
	}
	u := *s.url
	u.Path = path.Join(u.Path, "/")
	resp, err := s.client.Get(u.String())
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var root struct {
				ClusterName string `json:"cluster_name"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&root); err == nil && root.ClusterName != "" {
				s.lastClusterInfo = &clusterinfo.Response{ClusterName: root.ClusterName}
				return root.ClusterName
			}
		}
	}
	return "unknown_cluster"
}

// NewShards defines Shards Prometheus metrics
func NewShards(logger *slog.Logger, client *http.Client, url *url.URL) *Shards {
	var shardPtr *Shards
	nodeLabels := labels{
		keys: func(...string) []string {
			return []string{"node", "cluster"}
		},
		values: func(lastClusterinfo *clusterinfo.Response, base ...string) []string {
			if lastClusterinfo != nil {
				return append(base, lastClusterinfo.ClusterName)
			}
			if shardPtr != nil {
				return append(base, fetchClusterNameOnce(shardPtr))
			}
			return append(base, "unknown_cluster")
		},
	}

	shards := &Shards{
		// will assign later

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
			},
		},

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

	shardPtr = shards
	return shards
}

// Describe Shards
func (s *Shards) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.jsonParseFailures.Desc()

	for _, metric := range s.nodeShardMetrics {
		ch <- metric.Desc
	}
}

// streamShards decodes a /_cat/shards JSON array one element at a time,
// invoking emit for each, so the full shard list is never materialized.
func streamShards(r io.Reader, emit func(ShardResponse)) error {
	dec := json.NewDecoder(r)

	if _, err := dec.Token(); err != nil { // opening '[' of the array
		return err
	}
	for dec.More() {
		var shard ShardResponse
		if err := dec.Decode(&shard); err != nil {
			return err
		}
		emit(shard)
	}
	return nil
}

// streamAndAggregateShards GETs /_cat/shards and counts STARTED shards per
// node while decoding the response one shard at a time. This keeps peak
// memory proportional to a single shard entry instead of materializing the
// entire (potentially many-thousand-element) shard list.
func (s *Shards) streamAndAggregateShards(ctx context.Context, nodeShards map[string]float64) error {
	u := *s.url
	u.Path = path.Join(u.Path, "/_cat/shards")
	q := u.Query()
	q.Set("format", "json")
	u.RawQuery = q.Encode()

	return fetchURL(ctx, s.client, s.logger, u.String(), func(r io.Reader) error {
		if err := streamShards(r, func(shard ShardResponse) {
			if shard.State == "STARTED" {
				nodeShards[shard.Node]++
			}
		}); err != nil {
			s.jsonParseFailures.Inc()
			return err
		}
		return nil
	})
}

// Collect number of shards on each node
func (s *Shards) Collect(ch chan<- prometheus.Metric) {
	defer func() {
		ch <- s.jsonParseFailures
	}()

	nodeShards := make(map[string]float64)
	if err := s.streamAndAggregateShards(context.TODO(), nodeShards); err != nil {
		s.logger.Warn(
			"failed to fetch and decode node shards stats",
			"err", err,
		)
		return
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
