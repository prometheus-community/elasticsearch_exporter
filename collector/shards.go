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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus/client_golang/prometheus"
)

var nodeShardsTotal = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, "node_shards", "total"),
	"Total shards per node",
	[]string{"node", "cluster"}, nil,
)

func init() {
	registerCollector("shards", defaultDisabled, NewShards)
}

// ShardResponse has shard's node and index info
type ShardResponse struct {
	Index string `json:"index"`
	Shard string `json:"shard"`
	State string `json:"state"`
	Node  string `json:"node"`
}

// Shards information struct
type Shards struct {
	logger *slog.Logger
	client *http.Client
	url    *url.URL

	jsonParseFailures prometheus.Counter
}

// NewShards defines Shards Prometheus metrics
func NewShards(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &Shards{
		logger: logger,
		client: hc,
		url:    u,

		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "node_shards", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
	}, nil
}

// streamShards decodes a /_cat/shards JSON array one element at a time,
// invoking emit for each, so the full shard list is never materialized.
func streamShards(r io.Reader, emit func(ShardResponse)) error {
	dec := json.NewDecoder(r)

	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim('[') {
		return fmt.Errorf("expected '[' opening the shards array, got %v", tok)
	}
	for dec.More() {
		var shard ShardResponse
		if err := dec.Decode(&shard); err != nil {
			return err
		}
		emit(shard)
	}
	// More also reports false at EOF or on a read error, so a body cut off
	// right after a complete shard is only caught by consuming the closing ']'.
	_, err = dec.Token()
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
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

// Update collects the number of shards on each node
func (s *Shards) Update(ctx context.Context, uc UpdateContext, ch chan<- prometheus.Metric) error {
	defer func() {
		ch <- s.jsonParseFailures
	}()

	nodeShards := make(map[string]float64)
	if err := s.streamAndAggregateShards(ctx, nodeShards); err != nil {
		return fmt.Errorf("failed to fetch and decode node shards stats: %w", err)
	}

	clusterName := "unknown_cluster"
	if info, err := uc.GetClusterInfo(ctx); err == nil && info.ClusterName != "" {
		clusterName = info.ClusterName
	}

	for node, shards := range nodeShards {
		ch <- prometheus.MustNewConstMetric(
			nodeShardsTotal,
			prometheus.GaugeValue,
			shards,
			node, clusterName,
		)
	}

	return nil
}
