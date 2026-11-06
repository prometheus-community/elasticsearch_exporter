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
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	colors                     = []string{"green", "yellow", "red"}
	defaultClusterHealthLabels = []string{"cluster"}

	clusterHealthActivePrimaryShards = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "active_primary_shards"),
		"The number of primary shards in your cluster. This is an aggregate total across all indices.",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthActiveShards = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "active_shards"),
		"Aggregate total of all shards across all indices, which includes replica shards.",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthDelayedUnassignedShards = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "delayed_unassigned_shards"),
		"Shards delayed to reduce reallocation overhead",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthInitializingShards = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "initializing_shards"),
		"Count of shards that are being freshly created.",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthNumberOfDataNodes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "number_of_data_nodes"),
		"Number of data nodes in the cluster.",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthNumberOfInFlightFetch = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "number_of_in_flight_fetch"),
		"The number of ongoing shard info requests.",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthTaskMaxWaitingInQueueMillis = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "task_max_waiting_in_queue_millis"),
		"Tasks max time waiting in queue.",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthNumberOfNodes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "number_of_nodes"),
		"Number of nodes in the cluster.",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthNumberOfPendingTasks = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "number_of_pending_tasks"),
		"Cluster level changes which have not yet been executed",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthRelocatingShards = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "relocating_shards"),
		"The number of shards that are currently moving from one node to another node.",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthUnassignedShards = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "unassigned_shards"),
		"The number of shards that exist in the cluster state, but cannot be found in the cluster itself.",
		defaultClusterHealthLabels, nil,
	)
	clusterHealthStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_health", "status"),
		"Whether all primary and replica shards are allocated.",
		[]string{"cluster", "color"}, nil,
	)
)

func init() {
	registerCollector("cluster-health", defaultEnabled, NewClusterHealth)
}

// ClusterHealth type defines the collector struct
type ClusterHealth struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

type clusterHealthResponse struct {
	ClusterName                 string  `json:"cluster_name"`
	Status                      string  `json:"status"`
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}

// NewClusterHealth returns a new Collector exposing ClusterHealth stats.
func NewClusterHealth(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &ClusterHealth{
		logger: logger,
		hc:     hc,
		u:      u,
	}, nil
}

// Update collects ClusterHealth metrics.
func (c *ClusterHealth) Update(ctx context.Context, _ UpdateContext, ch chan<- prometheus.Metric) error {
	var chr clusterHealthResponse
	u := *c.u
	u.Path = path.Join(u.Path, "/_cluster/health")

	if err := getAndDecodeURL(ctx, c.hc, c.logger, u.String(), &chr); err != nil {
		return err
	}

	gauges := []struct {
		desc  *prometheus.Desc
		value int
	}{
		{clusterHealthActivePrimaryShards, chr.ActivePrimaryShards},
		{clusterHealthActiveShards, chr.ActiveShards},
		{clusterHealthDelayedUnassignedShards, chr.DelayedUnassignedShards},
		{clusterHealthInitializingShards, chr.InitializingShards},
		{clusterHealthNumberOfDataNodes, chr.NumberOfDataNodes},
		{clusterHealthNumberOfInFlightFetch, chr.NumberOfInFlightFetch},
		{clusterHealthTaskMaxWaitingInQueueMillis, chr.TaskMaxWaitingInQueueMillis},
		{clusterHealthNumberOfNodes, chr.NumberOfNodes},
		{clusterHealthNumberOfPendingTasks, chr.NumberOfPendingTasks},
		{clusterHealthRelocatingShards, chr.RelocatingShards},
		{clusterHealthUnassignedShards, chr.UnassignedShards},
	}
	for _, g := range gauges {
		ch <- prometheus.MustNewConstMetric(g.desc, prometheus.GaugeValue, float64(g.value), chr.ClusterName)
	}

	for _, color := range colors {
		ch <- prometheus.MustNewConstMetric(
			clusterHealthStatus,
			prometheus.GaugeValue,
			bool2Float(chr.Status == color),
			chr.ClusterName, color,
		)
	}

	return nil
}
