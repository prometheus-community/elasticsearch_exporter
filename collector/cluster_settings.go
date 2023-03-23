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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-kit/log"
	"github.com/imdario/mergo"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("clustersettings", defaultDisabled, NewClusterSettings)
}

type ClusterSettingsCollector struct {
	logger log.Logger
	u      *url.URL
	hc     *http.Client
}

func NewClusterSettings(logger log.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &ClusterSettingsCollector{
		logger: logger,
		u:      u,
		hc:     hc,
	}, nil
}

var clusterSettingsDesc = map[string]*prometheus.Desc{
	"shardAllocationEnabled": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "clustersettings_stats", "shard_allocation_enabled"),
		"Current mode of cluster wide shard routing allocation settings.",
		nil, nil,
	),

	"maxShardsPerNode": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "clustersettings_stats", "max_shards_per_node"),
		"Current maximum number of shards per node setting.",
		nil, nil,
	),
}

// clusterSettingsResponse is a representation of a Elasticsearch Cluster Settings
type clusterSettingsResponse struct {
	Defaults   clusterSettingsSection `json:"defaults"`
	Persistent clusterSettingsSection `json:"persistent"`
	Transient  clusterSettingsSection `json:"transient"`
}

// clusterSettingsSection is a representation of a Elasticsearch Cluster Settings
type clusterSettingsSection struct {
	Cluster clusterSettingsCluster `json:"cluster"`
}

// clusterSettingsCluster is a representation of a Elasticsearch clusterSettingsCluster Settings
type clusterSettingsCluster struct {
	Routing clusterSettingsRouting `json:"routing"`
	// This can be either a JSON object (which does not contain the value we are interested in) or a string
	MaxShardsPerNode interface{} `json:"max_shards_per_node"`
}

// clusterSettingsRouting is a representation of a Elasticsearch Cluster shard routing configuration
type clusterSettingsRouting struct {
	Allocation clusterSettingsAllocation `json:"allocation"`
}

// clusterSettingsAllocation is a representation of a Elasticsearch Cluster shard routing allocation settings
type clusterSettingsAllocation struct {
	Enabled string `json:"enable"`
}

func (c *ClusterSettingsCollector) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	u := c.u.ResolveReference(&url.URL{Path: "_cluster/settings"})
	q := u.Query()
	q.Set("include_defaults", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var data clusterSettingsResponse
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	// Merge all settings into one struct
	merged := data.Defaults

	err = mergo.Merge(&merged, data.Persistent, mergo.WithOverride)
	if err != nil {
		return err
	}
	err = mergo.Merge(&merged, data.Transient, mergo.WithOverride)
	if err != nil {
		return err
	}

	// Max shards per node
	if maxShardsPerNodeString, ok := merged.Cluster.MaxShardsPerNode.(string); ok {
		maxShardsPerNode, err := strconv.ParseInt(maxShardsPerNodeString, 10, 64)
		if err == nil {
			ch <- prometheus.MustNewConstMetric(
				clusterSettingsDesc["maxShardsPerNode"],
				prometheus.GaugeValue,
				float64(maxShardsPerNode),
			)
		}
	}

	// Shard allocation enabled
	shardAllocationMap := map[string]int{
		"all":           0,
		"primaries":     1,
		"new_primaries": 2,
		"none":          3,
	}

	ch <- prometheus.MustNewConstMetric(
		clusterSettingsDesc["shardAllocationEnabled"],
		prometheus.GaugeValue,
		float64(shardAllocationMap[merged.Cluster.Routing.Allocation.Enabled]),
	)

	return nil
}
