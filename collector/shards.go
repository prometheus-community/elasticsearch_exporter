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
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/elasticsearch_exporter/pkg/clusterinfo"
)

var (
	shardsTotalLabels = []string{"node", "cluster"}
	shardsStateLabels = []string{"node", "cluster", "index", "shard"}

	shardsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "node_shards", "total"),
		"Total shards per node",
		shardsTotalLabels, nil,
	)
	jsonParseFailures = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "node_shards", "json_parse_failures"),
		"Number of errors while parsing JSON.",
		nil, nil,
	)
	shardsState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "node_shards", "state"),
		"Shard state allocated per node by index and shard (0=unassigned, 10=primary started, 11=primary initializing, 12=primary relocating, 20=replica started, 21=replica initializing, 22=replica relocating).",
		shardsStateLabels, nil,
	)
)

// ShardResponse has shard's node and index info
type ShardResponse struct {
	Index  string  `json:"index"`
	Shard  string  `json:"shard"`
	State  string  `json:"state"`
	Node   *string `json:"node,omitempty"`
	Prirep string  `json:"prirep"`
}

// Shards information struct
type Shards struct {
	logger          *slog.Logger
	client          *http.Client
	url             *url.URL
	clusterInfoCh   chan *clusterinfo.Response
	lastClusterInfo *clusterinfo.Response

	jsonParseFailures float64
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

// getClusterName performs a single request to the root endpoint to obtain the cluster name.
func (s *Shards) getClusterName() string {
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
	shards := &Shards{
		logger: logger,
		client: client,
		url:    url,

		clusterInfoCh: make(chan *clusterinfo.Response),
		lastClusterInfo: &clusterinfo.Response{
			ClusterName: "unknown_cluster",
		},
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
	ch <- jsonParseFailures
	ch <- shardsTotal
	ch <- shardsState
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
		s.jsonParseFailures++
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
		ch <- prometheus.MustNewConstMetric(
			jsonParseFailures,
			prometheus.CounterValue,
			s.jsonParseFailures,
		)
	}()

	sr, err := s.fetchAndDecodeShards()
	if err != nil {
		s.logger.Warn(
			"failed to fetch and decode shards",
			"err", err,
		)
		return
	}

	clusterName := s.getClusterName()

	nodeShards := make(map[string]float64)

	for _, shard := range sr {
		node := "-"
		if shard.Node != nil {
			node = *shard.Node
		}
		if shard.State == "STARTED" {
			nodeShards[node]++
		}

		ch <- prometheus.MustNewConstMetric(
			shardsState,
			prometheus.GaugeValue,
			s.encodeState(shard),
			node,
			clusterName,
			shard.Index,
			shard.Shard,
		)
	}

	for node, shards := range nodeShards {
		ch <- prometheus.MustNewConstMetric(
			shardsTotal,
			prometheus.GaugeValue,
			shards,
			node,
			clusterName,
		)
	}
}

func (s *Shards) encodeState(shard ShardResponse) float64 {
	if shard.Node == nil || shard.State == "UNASSIGNED" {
		return 0
	}

	var state float64
	switch shard.Prirep {
	case "p":
		state = 10
	case "r":
		state = 20
	default:
		s.logger.Warn("unknown shard type", "type", shard.Prirep)
		return 0
	}

	switch shard.State {
	case "STARTED":
		return state
	case "INITIALIZING":
		state += 1
	case "RELOCATING":
		state += 2
	default:
		s.logger.Warn("unknown shard state", "state", shard.State)
		return 0
	}
	return state
}
