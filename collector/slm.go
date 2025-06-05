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
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/elasticsearch_exporter/pkg/clusterinfo"
)

var statuses = []string{"RUNNING", "STOPPING", "STOPPED"}

var (
	slmRetentionRunsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "retention_runs_total"),
		"Total retention runs",
		[]string{"cluster"}, nil,
	)
	slmRetentionFailedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "retention_failed_total"),
		"Total failed retention runs",
		[]string{"cluster"}, nil,
	)
	slmRetentionTimedOutTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "retention_timed_out_total"),
		"Total timed out retention runs",
		[]string{"cluster"}, nil,
	)
	slmRetentionDeletionTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "retention_deletion_time_seconds"),
		"Retention run deletion time",
		[]string{"cluster"}, nil,
	)
	slmTotalSnapshotsTaken = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "total_snapshots_taken_total"),
		"Total snapshots taken",
		[]string{"cluster"}, nil,
	)
	slmTotalSnapshotsFailed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "total_snapshots_failed_total"),
		"Total snapshots failed",
		[]string{"cluster"}, nil,
	)
	slmTotalSnapshotsDeleted = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "total_snapshots_deleted_total"),
		"Total snapshots deleted",
		[]string{"cluster"}, nil,
	)
	slmTotalSnapshotsDeleteFailed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "total_snapshot_deletion_failures_total"),
		"Total snapshot deletion failures",
		[]string{"cluster"}, nil,
	)

	slmOperationMode = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "operation_mode"),
		"Operating status of SLM",
		[]string{"cluster", "operation_mode"}, nil,
	)

	slmSnapshotsTaken = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "snapshots_taken_total"),
		"Total snapshots taken",
		[]string{
			"policy",
			"cluster",
		}, nil,
	)
	slmSnapshotsFailed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "snapshots_failed_total"),
		"Total snapshots failed",
		[]string{
			"policy",
			"cluster",
		}, nil,
	)
	slmSnapshotsDeleted = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "snapshots_deleted_total"),
		"Total snapshots deleted",
		[]string{
			"policy",
			"cluster",
		}, nil,
	)
	slmSnapshotsDeletionFailure = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "slm_stats", "snapshot_deletion_failures_total"),
		"Total snapshot deletion failures",
		[]string{
			"policy",
			"cluster",
		}, nil,
	)
)

func init() {
	registerCollector("slm", defaultDisabled, NewSLM)
}

// SLM information struct
type SLM struct {
	logger          *slog.Logger
	hc              *http.Client
	u               *url.URL
	clusterInfoCh   chan *clusterinfo.Response
	lastClusterInfo *clusterinfo.Response
}

// NewSLM defines SLM Prometheus metrics
func NewSLM(logger *slog.Logger, u *url.URL, hc *http.Client, ci *clusterinfo.Retriever) (Collector, error) {
	slm := &SLM{
		logger:        logger,
		hc:            hc,
		u:             u,
		clusterInfoCh: make(chan *clusterinfo.Response),
		lastClusterInfo: &clusterinfo.Response{
			ClusterName: "unknown_cluster",
		},
	}

	err := ci.RegisterConsumer(slm)
	if err != nil {
		return slm, err
	}

	return slm, nil
}

func (s *SLM) Describe(ch chan<- *prometheus.Desc) {
	ch <- slmRetentionRunsTotal
	ch <- slmRetentionFailedTotal
	ch <- slmRetentionTimedOutTotal
	ch <- slmRetentionDeletionTimeSeconds
	ch <- slmTotalSnapshotsTaken
	ch <- slmTotalSnapshotsFailed
	ch <- slmTotalSnapshotsDeleted
	ch <- slmTotalSnapshotsDeleteFailed
	ch <- slmOperationMode
	ch <- slmSnapshotsTaken
	ch <- slmSnapshotsFailed
	ch <- slmSnapshotsDeleted
	ch <- slmSnapshotsDeletionFailure
}

func (s *SLM) ClusterLabelUpdates() *chan *clusterinfo.Response {
	return &s.clusterInfoCh
}

func (s *SLM) SetClusterInfo(r *clusterinfo.Response) {
	s.lastClusterInfo = r
}

func (s *SLM) String() string {
	return namespace + "slm"
}

// SLMStatsResponse is a representation of the SLM stats
type SLMStatsResponse struct {
	RetentionRuns                 int64         `json:"retention_runs"`
	RetentionFailed               int64         `json:"retention_failed"`
	RetentionTimedOut             int64         `json:"retention_timed_out"`
	RetentionDeletionTime         string        `json:"retention_deletion_time"`
	RetentionDeletionTimeMillis   int64         `json:"retention_deletion_time_millis"`
	TotalSnapshotsTaken           int64         `json:"total_snapshots_taken"`
	TotalSnapshotsFailed          int64         `json:"total_snapshots_failed"`
	TotalSnapshotsDeleted         int64         `json:"total_snapshots_deleted"`
	TotalSnapshotDeletionFailures int64         `json:"total_snapshot_deletion_failures"`
	PolicyStats                   []PolicyStats `json:"policy_stats"`
}

// PolicyStats is a representation of SLM stats for specific policies
type PolicyStats struct {
	Policy                   string `json:"policy"`
	SnapshotsTaken           int64  `json:"snapshots_taken"`
	SnapshotsFailed          int64  `json:"snapshots_failed"`
	SnapshotsDeleted         int64  `json:"snapshots_deleted"`
	SnapshotDeletionFailures int64  `json:"snapshot_deletion_failures"`
}

// SLMStatusResponse is a representation of the SLM status
type SLMStatusResponse struct {
	OperationMode string `json:"operation_mode"`
}

func (s *SLM) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	u := s.u.ResolveReference(&url.URL{Path: "/_slm/status"})
	var slmStatusResp SLMStatusResponse

	resp, err := getURL(ctx, s.hc, s.logger, u.String())
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, &slmStatusResp)
	if err != nil {
		return err
	}

	u = s.u.ResolveReference(&url.URL{Path: "/_slm/stats"})
	var slmStatsResp SLMStatsResponse

	resp, err = getURL(ctx, s.hc, s.logger, u.String())
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, &slmStatsResp)
	if err != nil {
		return err
	}

	for _, status := range statuses {
		var value float64
		if slmStatusResp.OperationMode == status {
			value = 1
		}
		ch <- prometheus.MustNewConstMetric(
			slmOperationMode,
			prometheus.GaugeValue,
			value,
			s.lastClusterInfo.ClusterName,
			status,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		slmRetentionRunsTotal,
		prometheus.CounterValue,
		float64(slmStatsResp.RetentionRuns),
		s.lastClusterInfo.ClusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		slmRetentionFailedTotal,
		prometheus.CounterValue,
		float64(slmStatsResp.RetentionFailed),
		s.lastClusterInfo.ClusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		slmRetentionTimedOutTotal,
		prometheus.CounterValue,
		float64(slmStatsResp.RetentionTimedOut),
		s.lastClusterInfo.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		slmRetentionDeletionTimeSeconds,
		prometheus.GaugeValue,
		float64(slmStatsResp.RetentionDeletionTimeMillis)/1000,
		s.lastClusterInfo.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		slmTotalSnapshotsTaken,
		prometheus.CounterValue,
		float64(slmStatsResp.TotalSnapshotsTaken),
		s.lastClusterInfo.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		slmTotalSnapshotsFailed,
		prometheus.CounterValue,
		float64(slmStatsResp.TotalSnapshotsFailed),
		s.lastClusterInfo.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		slmTotalSnapshotsDeleted,
		prometheus.CounterValue,
		float64(slmStatsResp.TotalSnapshotsDeleted),
		s.lastClusterInfo.ClusterName,
	)
	ch <- prometheus.MustNewConstMetric(
		slmTotalSnapshotsDeleteFailed,
		prometheus.CounterValue,
		float64(slmStatsResp.TotalSnapshotDeletionFailures),
		s.lastClusterInfo.ClusterName,
	)

	for _, policy := range slmStatsResp.PolicyStats {
		ch <- prometheus.MustNewConstMetric(
			slmSnapshotsTaken,
			prometheus.CounterValue,
			float64(policy.SnapshotsTaken),
			policy.Policy,
			s.lastClusterInfo.ClusterName,
		)
		ch <- prometheus.MustNewConstMetric(
			slmSnapshotsFailed,
			prometheus.CounterValue,
			float64(policy.SnapshotsFailed),
			policy.Policy,
			s.lastClusterInfo.ClusterName,
		)
		ch <- prometheus.MustNewConstMetric(
			slmSnapshotsDeleted,
			prometheus.CounterValue,
			float64(policy.SnapshotsDeleted),
			policy.Policy,
			s.lastClusterInfo.ClusterName,
		)
		ch <- prometheus.MustNewConstMetric(
			slmSnapshotsDeletionFailure,
			prometheus.CounterValue,
			float64(policy.SnapshotDeletionFailures),
			policy.Policy,
			s.lastClusterInfo.ClusterName,
		)
	}

	return nil
}
