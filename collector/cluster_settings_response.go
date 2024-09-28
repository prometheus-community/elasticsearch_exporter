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
	"github.com/mitchellh/mapstructure"
)

// clusterSettingsResponse is a representation of a Elasticsearch Cluster Settings
type clusterSettingsResponse struct {
	Defaults   clusterSettingsSection `json:"defaults"`
	Persistent clusterSettingsSection `json:"persistent"`
	Transient  clusterSettingsSection `json:"transient"`
}

// clusterSettingsSection is a representation of a Elasticsearch Cluster Settings
type clusterSettingsSection struct {
	Cluster clusterSettingsCluster `json:"cluster" mapstructure:",squash"`
}

// clusterSettingsCluster is a representation of a Elasticsearch clusterSettingsCluster Settings
type clusterSettingsCluster struct {
	Routing clusterSettingsRouting `json:"routing" mapstructure:",squash"`

	MaxShardsPerNode string `json:"max_shards_per_node" mapstructure:"cluster.max_shards_per_node"`
}

// clusterSettingsRouting is a representation of a Elasticsearch Cluster shard routing configuration
type clusterSettingsRouting struct {
	Allocation clusterSettingsAllocation `json:"allocation" mapstructure:",squash"`
}

// clusterSettingsAllocation is a representation of a Elasticsearch Cluster shard routing allocation settings
type clusterSettingsAllocation struct {
	Enabled string              `json:"enable" mapstructure:"cluster.routing.allocation.enable"`
	Disk    clusterSettingsDisk `json:"disk" mapstructure:",squash"`
}

// clusterSettingsDisk is a representation of a Elasticsearch Cluster shard routing disk allocation settings
type clusterSettingsDisk struct {
	ThresholdEnabled string                   `json:"threshold_enabled" mapstructure:"cluster.routing.allocation.disk.threshold_enabled"`
	Watermark        clusterSettingsWatermark `json:"watermark" mapstructure:",squash"`
}

// clusterSettingsWatermark is representation of Elasticsearch Cluster shard routing disk allocation watermark settings
type clusterSettingsWatermark struct {
	FloodStage string `json:"flood_stage" mapstructure:"cluster.routing.allocation.disk.watermark.flood_stage"`
	High       string `json:"high" mapstructure:"cluster.routing.allocation.disk.watermark.high"`
	Low        string `json:"low" mapstructure:"cluster.routing.allocation.disk.watermark.low"`
}

func (c *clusterSettingsSection) UnmarshalJSON(data []byte) error {
	var settings map[string]interface{}

	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	settings = flatten(settings)
	return mapstructure.Decode(settings, c)
}

func flatten(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k1, v1 := range m {
		if n, ok := v1.(map[string]interface{}); ok {
			for k2, v2 := range flatten(n) {
				result[k1+"."+k2] = v2
			}
		} else {
			result[k1] = v1
		}
	}
	return result
}
