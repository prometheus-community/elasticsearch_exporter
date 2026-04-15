// Copyright 2025 The Prometheus Authors
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

	"github.com/prometheus/client_golang/prometheus"
)

var (
	statusColors              = []string{"green", "yellow", "red"}
	defaultHealthReportLabels = []string{"cluster"}
)

var (
	healthReportTotalRepositories = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "total_repositories"),
		"The number of snapshot repositories",
		defaultHealthReportLabels, nil,
	)
	healthReportMaxShardsInClusterData = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "max_shards_in_cluster_data"),
		"The number of maximum shards in a cluster",
		defaultHealthReportLabels, nil,
	)
	healthReportMaxShardsInClusterFrozen = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "max_shards_in_cluster_frozen"),
		"The number of maximum frozen shards in a cluster",
		defaultHealthReportLabels, nil,
	)
	healthReportRestartingReplicas = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "restarting_replicas"),
		"The number of restarting replica shards",
		defaultHealthReportLabels, nil,
	)
	healthReportCreatingPrimaries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "creating_primaries"),
		"The number of creating primary shards",
		defaultHealthReportLabels, nil,
	)
	healthReportInitializingReplicas = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "initializing_replicas"),
		"The number of initializing replica shards",
		defaultHealthReportLabels, nil,
	)
	healthReportUnassignedReplicas = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "unassigned_replicas"),
		"The number of unassigned replica shards",
		defaultHealthReportLabels, nil,
	)
	healthReportStartedPrimaries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "started_primaries"),
		"The number of started primary shards",
		defaultHealthReportLabels, nil,
	)
	healthReportRestartingPrimaries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "restarting_primaries"),
		"The number of restarting primary shards",
		defaultHealthReportLabels, nil,
	)
	healthReportInitializingPrimaries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "initializing_primaries"),
		"The number of initializing primary shards",
		defaultHealthReportLabels, nil,
	)
	healthReportCreatingReplicas = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "creating_replicas"),
		"The number of creating replica shards",
		defaultHealthReportLabels, nil,
	)
	healthReportStartedReplicas = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "started_replicas"),
		"The number of started replica shards",
		defaultHealthReportLabels, nil,
	)
	healthReportUnassignedPrimaries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "unassigned_primaries"),
		"The number of unassigned primary shards",
		defaultHealthReportLabels, nil,
	)
	healthReportSlmPolicies = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "slm_policies"),
		"The number of SLM policies",
		defaultHealthReportLabels, nil,
	)
	healthReportIlmPolicies = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "ilm_policies"),
		"The number of ILM Policies",
		defaultHealthReportLabels, nil,
	)
	healthReportIlmStagnatingIndices = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "ilm_stagnating_indices"),
		"The number of stagnating indices",
		defaultHealthReportLabels, nil,
	)
	healthReportStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "status"),
		"Overall cluster status",
		[]string{"cluster", "color"}, nil,
	)
	healthReportMasterIsStableStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "master_is_stable_status"),
		"Master is stable status",
		[]string{"cluster", "color"}, nil,
	)
	healthReportRepositoryIntegrityStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "repository_integrity_status"),
		"Repository integrity status",
		[]string{"cluster", "color"}, nil,
	)
	healthReportDiskStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "disk_status"),
		"Disk status",
		[]string{"cluster", "color"}, nil,
	)
	healthReportShardsCapacityStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "shards_capacity_status"),
		"Shards capacity status",
		[]string{"cluster", "color"}, nil,
	)
	healthReportShardsAvailabiltystatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "shards_availabilty_status"),
		"Shards availabilty status",
		[]string{"cluster", "color"}, nil,
	)
	healthReportDataStreamLifecycleStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "data_stream_lifecycle_status"),
		"Data stream lifecycle status",
		[]string{"cluster", "color"}, nil,
	)
	healthReportSlmStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "slm_status"),
		"SLM status",
		[]string{"cluster", "color"}, nil,
	)
	healthReportIlmStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "health_report", "ilm_status"),
		"ILM status",
		[]string{"cluster", "color"}, nil,
	)
)

func init() {
	registerCollector("health-report", defaultDisabled, NewHealthReport)
}

type HealthReport struct {
	logger *slog.Logger
	client *http.Client
	url    *url.URL
}

func NewHealthReport(logger *slog.Logger, url *url.URL, client *http.Client) (Collector, error) {
	return &HealthReport{
		logger: logger,
		client: client,
		url:    url,
	}, nil
}

type HealthReportResponse struct {
	ClusterName string                 `json:"cluster_name"`
	Status      string                 `json:"status"`
	Indicators  HealthReportIndicators `json:"indicators"`
}

type HealthReportIndicators struct {
	MasterIsStable      HealthReportMasterIsStable      `json:"master_is_stable"`
	RepositoryIntegrity HealthReportRepositoryIntegrity `json:"repository_integrity"`
	Disk                HealthReportDisk                `json:"disk"`
	ShardsCapacity      HealthReportShardsCapacity      `json:"shards_capacity"`
	ShardsAvailability  HealthReportShardsAvailability  `json:"shards_availability"`
	DataStreamLifecycle HealthReportDataStreamLifecycle `json:"data_stream_lifecycle"`
	Slm                 HealthReportSlm                 `json:"slm"`
	Ilm                 HealthReportIlm                 `json:"ilm"`
}

type HealthReportMasterIsStable struct {
	Status  string                            `json:"status"`
	Symptom string                            `json:"symptom"`
	Details HealthReportMasterIsStableDetails `json:"details"`
}

type HealthReportMasterIsStableDetails struct {
	CurrentMaster HealthReportMasterIsStableDetailsNode   `json:"current_master"`
	RecentMasters []HealthReportMasterIsStableDetailsNode `json:"recent_masters"`
}

type HealthReportMasterIsStableDetailsNode struct {
	NodeID string `json:"node_id"`
	Name   string `json:"name"`
}

type HealthReportRepositoryIntegrity struct {
	Status  string                                  `json:"status"`
	Symptom string                                  `json:"symptom"`
	Details HealthReportRepositoriyIntegrityDetails `json:"details"`
}

type HealthReportRepositoriyIntegrityDetails struct {
	TotalRepositories int `json:"total_repositories"`
}

type HealthReportDisk struct {
	Status  string                  `json:"status"`
	Symptom string                  `json:"symptom"`
	Details HealthReportDiskDetails `json:"details"`
}

type HealthReportDiskDetails struct {
	IndicesWithReadonlyBlock     int `json:"indices_with_readonly_block"`
	NodesWithEnoughDiskSpace     int `json:"nodes_with_enough_disk_space"`
	NodesWithUnknownDiskStatus   int `json:"nodes_with_unknown_disk_status"`
	NodesOverHighWatermark       int `json:"nodes_over_high_watermark"`
	NodesOverFloodStageWatermark int `json:"nodes_over_flood_stage_watermark"`
}

type HealthReportShardsCapacity struct {
	Status  string                            `json:"status"`
	Symptom string                            `json:"symptom"`
	Details HealthReportShardsCapacityDetails `json:"details"`
}

type HealthReportShardsCapacityDetails struct {
	Data   HealthReportShardsCapacityDetailsMaxShards `json:"data"`
	Frozen HealthReportShardsCapacityDetailsMaxShards `json:"frozen"`
}

type HealthReportShardsCapacityDetailsMaxShards struct {
	MaxShardsInCluster int `json:"max_shards_in_cluster"`
}

type HealthReportShardsAvailability struct {
	Status  string                                `json:"status"`
	Symptom string                                `json:"symptom"`
	Details HealthReportShardsAvailabilityDetails `json:"details"`
}

type HealthReportShardsAvailabilityDetails struct {
	RestartingReplicas    int `json:"restarting_replicas"`
	CreatingPrimaries     int `json:"creating_primaries"`
	InitializingReplicas  int `json:"initializing_replicas"`
	UnassignedReplicas    int `json:"unassigned_replicas"`
	StartedPrimaries      int `json:"started_primaries"`
	RestartingPrimaries   int `json:"restarting_primaries"`
	InitializingPrimaries int `json:"initializing_primaries"`
	CreatingReplicas      int `json:"creating_replicas"`
	StartedReplicas       int `json:"started_replicas"`
	UnassignedPrimaries   int `json:"unassigned_primaries"`
}

type HealthReportDataStreamLifecycle struct {
	Status  string `json:"status"`
	Symptom string `json:"symptom"`
}

type HealthReportSlm struct {
	Status  string                 `json:"status"`
	Symptom string                 `json:"symptom"`
	Details HealthReportSlmDetails `json:"details"`
}

type HealthReportSlmDetails struct {
	SlmStatus string `json:"slm_status"`
	Policies  int    `json:"policies"`
}

type HealthReportIlm struct {
	Status  string                 `json:"status"`
	Symptom string                 `json:"symptom"`
	Details HealthReportIlmDetails `json:"details"`
}

type HealthReportIlmDetails struct {
	Policies          int    `json:"policies"`
	StagnatingIndices int    `json:"stagnating_indices"`
	IlmStatus         string `json:"ilm_status"`
}

func statusValue(value string, color string) float64 {
	if value == color {
		return 1
	}
	return 0
}

func (c *HealthReport) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	u := c.url.ResolveReference(&url.URL{Path: "/_health_report"})
	var healthReportResponse HealthReportResponse

	if err := getAndDecodeURL(ctx, c.client, c.logger, u.String(), &healthReportResponse); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		healthReportTotalRepositories,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.RepositoryIntegrity.Details.TotalRepositories),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportMaxShardsInClusterData,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsCapacity.Details.Data.MaxShardsInCluster),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportMaxShardsInClusterFrozen,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsCapacity.Details.Frozen.MaxShardsInCluster),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportRestartingReplicas,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsAvailability.Details.RestartingReplicas),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportCreatingPrimaries,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsAvailability.Details.CreatingPrimaries),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportInitializingReplicas,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsAvailability.Details.InitializingReplicas),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportUnassignedReplicas,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsAvailability.Details.UnassignedReplicas),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportStartedPrimaries,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsAvailability.Details.StartedPrimaries),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportRestartingPrimaries,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsAvailability.Details.RestartingPrimaries),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportInitializingPrimaries,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsAvailability.Details.InitializingPrimaries),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportCreatingReplicas,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsAvailability.Details.CreatingReplicas),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportStartedReplicas,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsAvailability.Details.StartedReplicas),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportUnassignedPrimaries,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.ShardsAvailability.Details.UnassignedPrimaries),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportSlmPolicies,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.Slm.Details.Policies),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportIlmPolicies,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.Ilm.Details.Policies),
		healthReportResponse.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		healthReportIlmStagnatingIndices,
		prometheus.GaugeValue,
		float64(healthReportResponse.Indicators.Ilm.Details.StagnatingIndices),
		healthReportResponse.ClusterName,
	)

	for _, color := range statusColors {
		ch <- prometheus.MustNewConstMetric(
			healthReportStatus,
			prometheus.GaugeValue,
			statusValue(healthReportResponse.Status, color),
			healthReportResponse.ClusterName, color,
		)
		ch <- prometheus.MustNewConstMetric(
			healthReportMasterIsStableStatus,
			prometheus.GaugeValue,
			statusValue(healthReportResponse.Indicators.MasterIsStable.Status, color),
			healthReportResponse.ClusterName, color,
		)
		ch <- prometheus.MustNewConstMetric(
			healthReportRepositoryIntegrityStatus,
			prometheus.GaugeValue,
			statusValue(healthReportResponse.Indicators.RepositoryIntegrity.Status, color),
			healthReportResponse.ClusterName, color,
		)
		ch <- prometheus.MustNewConstMetric(
			healthReportDiskStatus,
			prometheus.GaugeValue,
			statusValue(healthReportResponse.Indicators.Disk.Status, color),
			healthReportResponse.ClusterName, color,
		)
		ch <- prometheus.MustNewConstMetric(
			healthReportShardsCapacityStatus,
			prometheus.GaugeValue,
			statusValue(healthReportResponse.Indicators.ShardsCapacity.Status, color),
			healthReportResponse.ClusterName, color,
		)
		ch <- prometheus.MustNewConstMetric(
			healthReportShardsAvailabiltystatus,
			prometheus.GaugeValue,
			statusValue(healthReportResponse.Indicators.ShardsAvailability.Status, color),
			healthReportResponse.ClusterName, color,
		)
		ch <- prometheus.MustNewConstMetric(
			healthReportDataStreamLifecycleStatus,
			prometheus.GaugeValue,
			statusValue(healthReportResponse.Indicators.DataStreamLifecycle.Status, color),
			healthReportResponse.ClusterName, color,
		)
		ch <- prometheus.MustNewConstMetric(
			healthReportSlmStatus,
			prometheus.GaugeValue,
			statusValue(healthReportResponse.Indicators.Slm.Status, color),
			healthReportResponse.ClusterName, color,
		)
		ch <- prometheus.MustNewConstMetric(
			healthReportIlmStatus,
			prometheus.GaugeValue,
			statusValue(healthReportResponse.Indicators.Ilm.Status, color),
			healthReportResponse.ClusterName, color,
		)
	}

	return nil
}
