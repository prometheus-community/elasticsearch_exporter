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
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultSnapshotLabels      = []string{"repository", "state", "version"}
	defaultSnapshotLabelValues = func(repositoryName string, snapshotStats SnapshotStatDataResponse) []string {
		return []string{repositoryName, snapshotStats.State, snapshotStats.Version}
	}
	defaultSnapshotRepositoryLabels      = []string{"repository"}
	defaultSnapshotRepositoryLabelValues = func(repositoryName string) []string {
		return []string{repositoryName}
	}
)

var (
	numIndices = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_number_of_indices"),
		"Number of indices in the last snapshot",
		defaultSnapshotLabels, nil,
	)
	snapshotStartTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_start_time_timestamp"),
		"Last snapshot start timestamp",
		defaultSnapshotLabels, nil,
	)
	snapshotEndTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_end_time_timestamp"),
		"Last snapshot end timestamp",
		defaultSnapshotLabels, nil,
	)
	snapshotNumFailures = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_number_of_failures"),
		"Last snapshot number of failures",
		defaultSnapshotLabels, nil,
	)
	snapshotNumShards = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_total_shards"),
		"Last snapshot total shards",
		defaultSnapshotLabels, nil,
	)
	snapshotFailedShards = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_failed_shards"),
		"Last snapshot failed shards",
		defaultSnapshotLabels, nil,
	)
	snapshotSuccessfulShards = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_successful_shards"),
		"Last snapshot successful shards",
		defaultSnapshotLabels, nil,
	)

	numSnapshots = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "snapshot_stats", "number_of_snapshots"),
		"Number of snapshots in a repository",
		defaultSnapshotRepositoryLabels, nil,
	)
	oldestSnapshotTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "snapshot_stats", "oldest_snapshot_timestamp"),
		"Timestamp of the oldest snapshot",
		defaultSnapshotRepositoryLabels, nil,
	)
	latestSnapshotTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "snapshot_stats", "latest_snapshot_timestamp_seconds"),
		"Timestamp of the latest SUCCESS or PARTIAL snapshot",
		defaultSnapshotRepositoryLabels, nil,
	)
)

func init() {
	registerCollector("snapshots", defaultDisabled, NewSnapshots)
}

// Snapshots information struct
type Snapshots struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

// NewSnapshots defines Snapshots Prometheus metrics
func NewSnapshots(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &Snapshots{
		logger: logger,
		u:      u,
		hc:     hc,
	}, nil
}

func (c *Snapshots) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	// indices
	snapshotsStatsResp := make(map[string]SnapshotStatsResponse)
	u := c.u.ResolveReference(&url.URL{Path: "/_snapshot"})

	var srr SnapshotRepositoriesResponse
	resp, err := getURL(ctx, c.hc, c.logger, u.String())
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, &srr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	for repository := range srr {
		pathPart := path.Join("/_snapshot", repository, "/_all")
		u := c.u.ResolveReference(&url.URL{Path: pathPart})
		var ssr SnapshotStatsResponse
		resp, err := getURL(ctx, c.hc, c.logger, u.String())
		if err != nil {
			continue
		}
		err = json.Unmarshal(resp, &ssr)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		snapshotsStatsResp[repository] = ssr
	}

	// Snapshots stats
	for repositoryName, snapshotStats := range snapshotsStatsResp {

		ch <- prometheus.MustNewConstMetric(
			numSnapshots,
			prometheus.GaugeValue,
			float64(len(snapshotStats.Snapshots)),
			defaultSnapshotRepositoryLabelValues(repositoryName)...,
		)

		oldest := float64(0)
		if len(snapshotStats.Snapshots) > 0 {
			oldest = float64(snapshotStats.Snapshots[0].StartTimeInMillis / 1000)
		}
		ch <- prometheus.MustNewConstMetric(
			oldestSnapshotTimestamp,
			prometheus.GaugeValue,
			oldest,
			defaultSnapshotRepositoryLabelValues(repositoryName)...,
		)

		latest := float64(0)
		for i := len(snapshotStats.Snapshots) - 1; i >= 0; i-- {
			var snap = snapshotStats.Snapshots[i]
			if snap.State == "SUCCESS" || snap.State == "PARTIAL" {
				latest = float64(snap.StartTimeInMillis / 1000)
				break
			}
		}
		ch <- prometheus.MustNewConstMetric(
			latestSnapshotTimestamp,
			prometheus.GaugeValue,
			latest,
			defaultSnapshotRepositoryLabelValues(repositoryName)...,
		)

		if len(snapshotStats.Snapshots) == 0 {
			continue
		}

		lastSnapshot := snapshotStats.Snapshots[len(snapshotStats.Snapshots)-1]
		ch <- prometheus.MustNewConstMetric(
			numIndices,
			prometheus.GaugeValue,
			float64(len(lastSnapshot.Indices)),
			defaultSnapshotLabelValues(repositoryName, lastSnapshot)...,
		)
		ch <- prometheus.MustNewConstMetric(
			snapshotStartTimestamp,
			prometheus.GaugeValue,
			float64(lastSnapshot.StartTimeInMillis/1000),
			defaultSnapshotLabelValues(repositoryName, lastSnapshot)...,
		)
		ch <- prometheus.MustNewConstMetric(
			snapshotEndTimestamp,
			prometheus.GaugeValue,
			float64(lastSnapshot.EndTimeInMillis/1000),
			defaultSnapshotLabelValues(repositoryName, lastSnapshot)...,
		)
		ch <- prometheus.MustNewConstMetric(
			snapshotNumFailures,
			prometheus.GaugeValue,
			float64(len(lastSnapshot.Failures)),
			defaultSnapshotLabelValues(repositoryName, lastSnapshot)...,
		)
		ch <- prometheus.MustNewConstMetric(
			snapshotNumShards,
			prometheus.GaugeValue,
			float64(lastSnapshot.Shards.Total),
			defaultSnapshotLabelValues(repositoryName, lastSnapshot)...,
		)
		ch <- prometheus.MustNewConstMetric(
			snapshotFailedShards,
			prometheus.GaugeValue,
			float64(lastSnapshot.Shards.Failed),
			defaultSnapshotLabelValues(repositoryName, lastSnapshot)...,
		)
		ch <- prometheus.MustNewConstMetric(
			snapshotSuccessfulShards,
			prometheus.GaugeValue,
			float64(lastSnapshot.Shards.Successful),
			defaultSnapshotLabelValues(repositoryName, lastSnapshot)...,
		)
	}

	return nil
}
