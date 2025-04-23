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
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("clustersettings", defaultDisabled, NewClusterSettings)
}

type ClusterSettingsCollector struct {
	logger *slog.Logger
	u      *url.URL
	hc     *http.Client
}

func NewClusterSettings(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
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

	"thresholdEnabled": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "clustersettings_allocation", "threshold_enabled"),
		"Is disk allocation decider enabled.",
		nil, nil,
	),

	"floodStageRatio": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "clustersettings_allocation_watermark", "flood_stage_ratio"),
		"Flood stage watermark as a ratio.",
		nil, nil,
	),

	"highRatio": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "clustersettings_allocation_watermark", "high_ratio"),
		"High watermark for disk usage as a ratio.",
		nil, nil,
	),

	"lowRatio": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "clustersettings_allocation_watermark", "low_ratio"),
		"Low watermark for disk usage as a ratio.",
		nil, nil,
	),

	"floodStageBytes": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "clustersettings_allocation_watermark", "flood_stage_bytes"),
		"Flood stage watermark as in bytes.",
		nil, nil,
	),

	"highBytes": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "clustersettings_allocation_watermark", "high_bytes"),
		"High watermark for disk usage in bytes.",
		nil, nil,
	),

	"lowBytes": prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "clustersettings_allocation_watermark", "low_bytes"),
		"Low watermark for disk usage in bytes.",
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
	Enabled string              `json:"enable"`
	Disk    clusterSettingsDisk `json:"disk"`
}

// clusterSettingsDisk is a representation of a Elasticsearch Cluster shard routing disk allocation settings
type clusterSettingsDisk struct {
	ThresholdEnabled string                   `json:"threshold_enabled"`
	Watermark        clusterSettingsWatermark `json:"watermark"`
}

// clusterSettingsWatermark is representation of Elasticsearch Cluster shard routing disk allocation watermark settings
type clusterSettingsWatermark struct {
	FloodStage string `json:"flood_stage"`
	High       string `json:"high"`
	Low        string `json:"low"`
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

	// Threshold enabled
	thresholdMap := map[string]int{
		"false": 0,
		"true":  1,
	}

	ch <- prometheus.MustNewConstMetric(
		clusterSettingsDesc["thresholdEnabled"],
		prometheus.GaugeValue,
		float64(thresholdMap[merged.Cluster.Routing.Allocation.Disk.ThresholdEnabled]),
	)

	// Watermark bytes or ratio metrics
	if strings.HasSuffix(merged.Cluster.Routing.Allocation.Disk.Watermark.High, "b") {
		flooodStageBytes, err := getValueInBytes(merged.Cluster.Routing.Allocation.Disk.Watermark.FloodStage)
		if err != nil {
			c.logger.Error("failed to parse flood_stage bytes", "err", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				clusterSettingsDesc["floodStageBytes"],
				prometheus.GaugeValue,
				flooodStageBytes,
			)
		}

		highBytes, err := getValueInBytes(merged.Cluster.Routing.Allocation.Disk.Watermark.High)
		if err != nil {
			c.logger.Error("failed to parse high bytes", "err", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				clusterSettingsDesc["highBytes"],
				prometheus.GaugeValue,
				highBytes,
			)
		}

		lowBytes, err := getValueInBytes(merged.Cluster.Routing.Allocation.Disk.Watermark.Low)
		if err != nil {
			c.logger.Error("failed to parse low bytes", "err", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				clusterSettingsDesc["lowBytes"],
				prometheus.GaugeValue,
				lowBytes,
			)
		}

		return nil
	}

	// Watermark ratio metrics
	floodRatio, err := getValueAsRatio(merged.Cluster.Routing.Allocation.Disk.Watermark.FloodStage)
	if err != nil {
		c.logger.Error("failed to parse flood_stage ratio", "err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(
			clusterSettingsDesc["floodStageRatio"],
			prometheus.GaugeValue,
			floodRatio,
		)
	}

	highRatio, err := getValueAsRatio(merged.Cluster.Routing.Allocation.Disk.Watermark.High)
	if err != nil {
		c.logger.Error("failed to parse high ratio", "err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(
			clusterSettingsDesc["highRatio"],
			prometheus.GaugeValue,
			highRatio,
		)
	}

	lowRatio, err := getValueAsRatio(merged.Cluster.Routing.Allocation.Disk.Watermark.Low)
	if err != nil {
		c.logger.Error("failed to parse low ratio", "err", err)
	} else {
		ch <- prometheus.MustNewConstMetric(
			clusterSettingsDesc["lowRatio"],
			prometheus.GaugeValue,
			lowRatio,
		)
	}

	return nil
}

func getValueInBytes(value string) (float64, error) {
	type UnitValue struct {
		unit string
		val  float64
	}

	unitValues := []UnitValue{
		{"pb", 1024 * 1024 * 1024 * 1024 * 1024},
		{"tb", 1024 * 1024 * 1024 * 1024},
		{"gb", 1024 * 1024 * 1024},
		{"mb", 1024 * 1024},
		{"kb", 1024},
		{"b", 1},
	}

	for _, uv := range unitValues {
		if strings.HasSuffix(value, uv.unit) {
			numberStr := strings.TrimSuffix(value, uv.unit)

			number, err := strconv.ParseFloat(numberStr, 64)
			if err != nil {
				return 0, err
			}
			return number * uv.val, nil
		}
	}

	return 0, fmt.Errorf("failed to convert unit %s to bytes", value)
}

func getValueAsRatio(value string) (float64, error) {
	if strings.HasSuffix(value, "%") {
		percentValue, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(value, "%")))
		if err != nil {
			return 0, err
		}

		return float64(percentValue) / 100, nil
	}

	ratio, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	return ratio, nil
}
