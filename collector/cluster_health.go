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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	colors = []string{"green", "yellow", "red"}
)

// ClusterHealth type defines the collector struct
type ClusterHealth struct {
	logger log.Logger
	client *http.Client
	url    *url.URL

	activePrimaryShards     *prometheus.Desc
	activeShards            *prometheus.Desc
	relocatingShards        *prometheus.Desc
	unassignedShards        *prometheus.Desc
	delayedUnassignedShards *prometheus.Desc
	initializingShards      *prometheus.Desc
	dataNodes               *prometheus.Desc
	nodes                   *prometheus.Desc
	inFlightFetch           *prometheus.Desc
	pendingTasks            *prometheus.Desc
	taskMaxWaitQueue        *prometheus.Desc
	status                  *prometheus.Desc
}

func (c *ClusterHealth) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	chr, err := c.fetch(ctx)
	if err != nil {
		return err
	}
	clusterName := chr.ClusterName

	ch <- prometheus.MustNewConstMetric(c.activePrimaryShards, prometheus.GaugeValue, float64(chr.ActivePrimaryShards), clusterName)
	ch <- prometheus.MustNewConstMetric(c.activeShards, prometheus.GaugeValue, float64(chr.ActiveShards), clusterName)
	ch <- prometheus.MustNewConstMetric(c.relocatingShards, prometheus.GaugeValue, float64(chr.RelocatingShards), clusterName)
	ch <- prometheus.MustNewConstMetric(c.unassignedShards, prometheus.GaugeValue, float64(chr.UnassignedShards), clusterName)
	ch <- prometheus.MustNewConstMetric(c.delayedUnassignedShards, prometheus.GaugeValue, float64(chr.DelayedUnassignedShards), clusterName)
	ch <- prometheus.MustNewConstMetric(c.initializingShards, prometheus.GaugeValue, float64(chr.InitializingShards), clusterName)
	ch <- prometheus.MustNewConstMetric(c.dataNodes, prometheus.GaugeValue, float64(chr.NumberOfDataNodes), clusterName)
	ch <- prometheus.MustNewConstMetric(c.nodes, prometheus.GaugeValue, float64(chr.NumberOfNodes), clusterName)
	ch <- prometheus.MustNewConstMetric(c.inFlightFetch, prometheus.GaugeValue, float64(chr.NumberOfInFlightFetch), clusterName)
	ch <- prometheus.MustNewConstMetric(c.pendingTasks, prometheus.GaugeValue, float64(chr.NumberOfPendingTasks), clusterName)
	ch <- prometheus.MustNewConstMetric(c.taskMaxWaitQueue, prometheus.GaugeValue, float64(chr.TaskMaxWaitingInQueueMillis/1000), clusterName)

	for _, color := range colors {
		if chr.Status == color {
			ch <- prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, 1, clusterName, color)
		} else {
			ch <- prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, 0, clusterName, color)
		}
	}

	return nil
}

// NewClusterHealth returns a new Collector exposing ClusterHealth stats.
func NewClusterHealth(logger log.Logger, client *http.Client, url *url.URL) *ClusterHealth {
	subsystem := "cluster_health"

	return &ClusterHealth{
		logger: logger,
		client: client,
		url:    url,

		activePrimaryShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "active_primary_shards"),
			"Number of active primary shards in the cluster. This is an aggregate of all indices.",
			[]string{clusterLabel},
			nil,
		),
		activeShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "active_shards"),
			"Number of shards in the cluster that are active including replica shards.",
			[]string{clusterLabel},
			nil,
		),
		relocatingShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "relocating_shards"),
			"Number of shards in the cluster that are relocating.",
			[]string{clusterLabel},
			nil,
		),
		unassignedShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "unassigned_shards"),
			"Number of shards in the cluster that are unassigned.",
			[]string{clusterLabel},
			nil,
		),
		delayedUnassignedShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "delayed_unassigned_shards"),
			"Shards delayed to reduce the reallocation overhead",
			[]string{clusterLabel},
			nil,
		),
		initializingShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "initializing_shards"),
			"Number of shards that are initializing.",
			[]string{clusterLabel},
			nil,
		),
		dataNodes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "data_nodes"),
			"Number of data nodes in the cluster.",
			[]string{clusterLabel},
			nil,
		),
		inFlightFetch: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "in_flight_fetch"),
			"Number of in-flight fetch requests.",
			[]string{clusterLabel},
			nil,
		),
		nodes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "nodes"),
			"Number of nodes in the cluster.",
			[]string{clusterLabel},
			nil,
		),
		pendingTasks: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pending_tasks"),
			"Number of pending tasks in the cluster.",
			[]string{clusterLabel},
			nil,
		),
		taskMaxWaitQueue: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "task_max_wait_queue_seconds"),
			"Maximum time of tasks waiting in the queue.",
			[]string{clusterLabel},
			nil,
		),
		status: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "status"),
			"Status of the cluster.",
			[]string{clusterLabel, "color"},
			nil,
		),
	}
}

// Describe set Prometheus metrics descriptions.
func (c *ClusterHealth) Describe(ch chan<- *prometheus.Desc) {
}

func (c *ClusterHealth) fetch(ctx context.Context) (clusterHealthResponse, error) {
	var chr clusterHealthResponse
	u := c.url.ResolveReference(&url.URL{Path: "_cluster/health"})
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return chr, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return chr, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return chr, fmt.Errorf("%s returned status %s", u.String(), resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return chr, fmt.Errorf("failed to read body: %v", err)
	}

	err = json.Unmarshal(body, &chr)
	if err != nil {
		return chr, fmt.Errorf("failed to unmarshal json: %v", err)
	}

	return chr, nil
}

func (c *ClusterHealth) fetchAndDecodeClusterHealth() (clusterHealthResponse, error) {
	var chr clusterHealthResponse

	u := *c.url
	u.Path = path.Join(u.Path, "/_cluster/health")
	res, err := c.client.Get(u.String())
	if err != nil {
		return chr, fmt.Errorf("failed to get cluster health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(c.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return chr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return chr, err
	}

	if err := json.Unmarshal(bts, &chr); err != nil {
		return chr, err
	}

	return chr, nil
}

// Collect collects ClusterHealth metrics.
func (c *ClusterHealth) Collect(ch chan<- prometheus.Metric) {

}
