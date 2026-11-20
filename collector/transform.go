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

	"github.com/prometheus/client_golang/prometheus"
)

// transformStates enumerates the possible states of a transform.
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/get-transform-stats.html
var transformStates = []string{"started", "indexing", "stopped", "stopping", "failed", "aborting", "waiting"}

// transformHealthStatuses enumerates the possible transform health statuses.
var transformHealthStatuses = []string{"green", "yellow", "red", "unknown"}

func transformLabels(t TransformStats) []string {
	return []string{t.ID}
}

var defaultTransformLabels = []string{"id"}

var (
	transformState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "state"),
		"State of the transform, one of: started, indexing, stopped, stopping, failed, aborting, waiting",
		append(defaultTransformLabels, "state"), nil,
	)
	transformHealthStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "health_status"),
		"Health status of the transform, one of: green, yellow, red, unknown",
		append(defaultTransformLabels, "status"), nil,
	)

	transformPagesProcessed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "pages_processed_total"),
		"The number of search or bulk index operations processed",
		defaultTransformLabels, nil,
	)
	transformDocumentsProcessed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "documents_processed_total"),
		"The number of documents processed",
		defaultTransformLabels, nil,
	)
	transformDocumentsIndexed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "documents_indexed_total"),
		"The number of documents indexed into the destination index",
		defaultTransformLabels, nil,
	)
	transformDocumentsDeleted = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "documents_deleted_total"),
		"The number of documents deleted from the destination index",
		defaultTransformLabels, nil,
	)
	transformTriggerCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "trigger_count_total"),
		"The number of times the transform has been triggered by the scheduler",
		defaultTransformLabels, nil,
	)
	transformIndexTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "index_time_seconds_total"),
		"The amount of time spent indexing in seconds",
		defaultTransformLabels, nil,
	)
	transformIndexTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "index_total"),
		"The number of index operations",
		defaultTransformLabels, nil,
	)
	transformIndexFailures = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "index_failures_total"),
		"The number of indexing failures",
		defaultTransformLabels, nil,
	)
	transformSearchTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "search_time_seconds_total"),
		"The amount of time spent searching in seconds",
		defaultTransformLabels, nil,
	)
	transformSearchTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "search_total"),
		"The number of search operations",
		defaultTransformLabels, nil,
	)
	transformSearchFailures = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "search_failures_total"),
		"The number of search failures",
		defaultTransformLabels, nil,
	)
	transformProcessingTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "processing_time_seconds_total"),
		"The amount of time spent processing results in seconds",
		defaultTransformLabels, nil,
	)
	transformProcessingTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "processing_total"),
		"The number of processing operations",
		defaultTransformLabels, nil,
	)
	transformDeleteTimeSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "delete_time_seconds_total"),
		"The amount of time spent deleting documents in seconds",
		defaultTransformLabels, nil,
	)
	transformExpAvgCheckpointDurationSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "exponential_avg_checkpoint_duration_seconds"),
		"The exponential moving average of the duration of the checkpoint, in seconds",
		defaultTransformLabels, nil,
	)
	transformExpAvgDocumentsIndexed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "exponential_avg_documents_indexed"),
		"The exponential moving average of the number of new documents that have been indexed",
		defaultTransformLabels, nil,
	)
	transformExpAvgDocumentsProcessed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "exponential_avg_documents_processed"),
		"The exponential moving average of the number of documents that have been processed",
		defaultTransformLabels, nil,
	)

	transformLastCheckpoint = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "checkpoint"),
		"The sequence number of the last completed checkpoint",
		defaultTransformLabels, nil,
	)
	transformOperationsBehind = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "transform_stats", "operations_behind"),
		"The number of operations in the source index that have not yet been processed",
		defaultTransformLabels, nil,
	)
)

func init() {
	registerCollector("transform", defaultDisabled, NewTransform)
}

// Transform information struct
type Transform struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

// NewTransform defines Transform Prometheus metrics
func NewTransform(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &Transform{
		logger: logger,
		hc:     hc,
		u:      u,
	}, nil
}

// TransformStatsResponse is a representation of the transform stats endpoint response
type TransformStatsResponse struct {
	Count      int64            `json:"count"`
	Transforms []TransformStats `json:"transforms"`
}

// TransformStats is a representation of the stats for a single transform
type TransformStats struct {
	ID            string                 `json:"id"`
	State         string                 `json:"state"`
	Stats         TransformIndexerStats  `json:"stats"`
	Checkpointing TransformCheckpointing `json:"checkpointing"`
	Health        *TransformHealth       `json:"health"`
}

// TransformIndexerStats holds the operational counters for a transform
type TransformIndexerStats struct {
	PagesProcessed                   int64   `json:"pages_processed"`
	DocumentsProcessed               int64   `json:"documents_processed"`
	DocumentsIndexed                 int64   `json:"documents_indexed"`
	DocumentsDeleted                 int64   `json:"documents_deleted"`
	TriggerCount                     int64   `json:"trigger_count"`
	IndexTimeInMillis                int64   `json:"index_time_in_ms"`
	IndexTotal                       int64   `json:"index_total"`
	IndexFailures                    int64   `json:"index_failures"`
	SearchTimeInMillis               int64   `json:"search_time_in_ms"`
	SearchTotal                      int64   `json:"search_total"`
	SearchFailures                   int64   `json:"search_failures"`
	ProcessingTimeInMillis           int64   `json:"processing_time_in_ms"`
	ProcessingTotal                  int64   `json:"processing_total"`
	DeleteTimeInMillis               int64   `json:"delete_time_in_ms"`
	ExpAvgCheckpointDurationInMillis float64 `json:"exponential_avg_checkpoint_duration_ms"`
	ExpAvgDocumentsIndexed           float64 `json:"exponential_avg_documents_indexed"`
	ExpAvgDocumentsProcessed         float64 `json:"exponential_avg_documents_processed"`
}

// TransformCheckpointing holds checkpoint progress information for a transform
type TransformCheckpointing struct {
	Last             TransformCheckpoint `json:"last"`
	OperationsBehind int64               `json:"operations_behind"`
}

// TransformCheckpoint represents a single transform checkpoint
type TransformCheckpoint struct {
	Checkpoint int64 `json:"checkpoint"`
}

// TransformHealth represents the health of a transform
type TransformHealth struct {
	Status string `json:"status"`
}

func (t *Transform) Update(ctx context.Context, _ UpdateContext, ch chan<- prometheus.Metric) error {
	u := t.u.ResolveReference(&url.URL{Path: "/_transform/_all/_stats"})
	var resp TransformStatsResponse

	if err := getAndDecodeURL(ctx, t.hc, t.logger, u.String(), &resp); err != nil {
		return err
	}

	for _, tr := range resp.Transforms {
		labels := transformLabels(tr)

		for _, state := range transformStates {
			var value float64
			if tr.State == state {
				value = 1
			}
			ch <- prometheus.MustNewConstMetric(
				transformState,
				prometheus.GaugeValue,
				value,
				append(labels, state)...,
			)
		}

		if tr.Health != nil {
			for _, status := range transformHealthStatuses {
				var value float64
				if tr.Health.Status == status {
					value = 1
				}
				ch <- prometheus.MustNewConstMetric(
					transformHealthStatus,
					prometheus.GaugeValue,
					value,
					append(labels, status)...,
				)
			}
		}

		s := tr.Stats
		ch <- prometheus.MustNewConstMetric(transformPagesProcessed, prometheus.CounterValue, float64(s.PagesProcessed), labels...)
		ch <- prometheus.MustNewConstMetric(transformDocumentsProcessed, prometheus.CounterValue, float64(s.DocumentsProcessed), labels...)
		ch <- prometheus.MustNewConstMetric(transformDocumentsIndexed, prometheus.CounterValue, float64(s.DocumentsIndexed), labels...)
		ch <- prometheus.MustNewConstMetric(transformDocumentsDeleted, prometheus.CounterValue, float64(s.DocumentsDeleted), labels...)
		ch <- prometheus.MustNewConstMetric(transformTriggerCount, prometheus.CounterValue, float64(s.TriggerCount), labels...)
		ch <- prometheus.MustNewConstMetric(transformIndexTimeSeconds, prometheus.CounterValue, float64(s.IndexTimeInMillis)/1000, labels...)
		ch <- prometheus.MustNewConstMetric(transformIndexTotal, prometheus.CounterValue, float64(s.IndexTotal), labels...)
		ch <- prometheus.MustNewConstMetric(transformIndexFailures, prometheus.CounterValue, float64(s.IndexFailures), labels...)
		ch <- prometheus.MustNewConstMetric(transformSearchTimeSeconds, prometheus.CounterValue, float64(s.SearchTimeInMillis)/1000, labels...)
		ch <- prometheus.MustNewConstMetric(transformSearchTotal, prometheus.CounterValue, float64(s.SearchTotal), labels...)
		ch <- prometheus.MustNewConstMetric(transformSearchFailures, prometheus.CounterValue, float64(s.SearchFailures), labels...)
		ch <- prometheus.MustNewConstMetric(transformProcessingTimeSeconds, prometheus.CounterValue, float64(s.ProcessingTimeInMillis)/1000, labels...)
		ch <- prometheus.MustNewConstMetric(transformProcessingTotal, prometheus.CounterValue, float64(s.ProcessingTotal), labels...)
		ch <- prometheus.MustNewConstMetric(transformDeleteTimeSeconds, prometheus.CounterValue, float64(s.DeleteTimeInMillis)/1000, labels...)
		ch <- prometheus.MustNewConstMetric(transformExpAvgCheckpointDurationSeconds, prometheus.GaugeValue, s.ExpAvgCheckpointDurationInMillis/1000, labels...)
		ch <- prometheus.MustNewConstMetric(transformExpAvgDocumentsIndexed, prometheus.GaugeValue, s.ExpAvgDocumentsIndexed, labels...)
		ch <- prometheus.MustNewConstMetric(transformExpAvgDocumentsProcessed, prometheus.GaugeValue, s.ExpAvgDocumentsProcessed, labels...)

		ch <- prometheus.MustNewConstMetric(transformLastCheckpoint, prometheus.GaugeValue, float64(tr.Checkpointing.Last.Checkpoint), labels...)
		ch <- prometheus.MustNewConstMetric(transformOperationsBehind, prometheus.GaugeValue, float64(tr.Checkpointing.OperationsBehind), labels...)
	}

	return nil
}
