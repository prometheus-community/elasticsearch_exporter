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
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var ccrFollowerStatuses = []string{"active", "paused"}

var ccrShardLabels = []string{"remote_cluster", "leader_index", "follower_index", "shard"}

var (
	ccrLeaderGlobalCheckpoint = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "leader_global_checkpoint"),
		"Current global checkpoint on the leader known to the follower task",
		ccrShardLabels, nil,
	)
	ccrLeaderMaxSeqNo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "leader_max_seq_no"),
		"Current maximum sequence number on the leader known to the follower task",
		ccrShardLabels, nil,
	)
	ccrFollowerGlobalCheckpoint = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "follower_global_checkpoint"),
		"Current global checkpoint on the follower",
		ccrShardLabels, nil,
	)
	ccrFollowerMaxSeqNo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "follower_max_seq_no"),
		"Current maximum sequence number on the follower",
		ccrShardLabels, nil,
	)
	ccrLastRequestedSeqNo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "last_requested_seq_no"),
		"Starting sequence number of the last batch of operations requested from the leader",
		ccrShardLabels, nil,
	)
	ccrOutstandingReadRequests = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "outstanding_read_requests"),
		"Number of active read requests from the follower",
		ccrShardLabels, nil,
	)
	ccrOutstandingWriteRequests = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "outstanding_write_requests"),
		"Number of active bulk write requests on the follower",
		ccrShardLabels, nil,
	)
	ccrWriteBufferOperations = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "write_buffer_operations"),
		"Number of write operations queued on the follower",
		ccrShardLabels, nil,
	)
	ccrWriteBufferSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "write_buffer_size_bytes"),
		"Total size of operations queued for writing on the follower",
		ccrShardLabels, nil,
	)
	ccrFollowerMappingVersion = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "follower_mapping_version"),
		"Index mapping version the follower is synced up to",
		ccrShardLabels, nil,
	)
	ccrFollowerSettingsVersion = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "follower_settings_version"),
		"Index settings version the follower is synced up to",
		ccrShardLabels, nil,
	)
	ccrFollowerAliasesVersion = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "follower_aliases_version"),
		"Index aliases version the follower is synced up to",
		ccrShardLabels, nil,
	)
	ccrReadTimeSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "read_time_seconds_total"),
		"Total time reads were outstanding, from sending the request to the leader until the reply was returned to the follower",
		ccrShardLabels, nil,
	)
	ccrReadRemoteExecTimeSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "read_remote_exec_time_seconds_total"),
		"Total time reads spent executing on the remote cluster",
		ccrShardLabels, nil,
	)
	ccrSuccessfulReadRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "successful_read_requests_total"),
		"Total number of successful fetches from the leader",
		ccrShardLabels, nil,
	)
	ccrFailedReadRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "failed_read_requests_total"),
		"Total number of failed reads from the leader",
		ccrShardLabels, nil,
	)
	ccrOperationsReadTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "operations_read_total"),
		"Total number of operations read from the leader",
		ccrShardLabels, nil,
	)
	ccrReadBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "read_bytes_total"),
		"Estimated total bytes read from the leader, not accounting for compression",
		ccrShardLabels, nil,
	)
	ccrWriteTimeSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "write_time_seconds_total"),
		"Total time spent writing on the follower",
		ccrShardLabels, nil,
	)
	ccrSuccessfulWriteRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "successful_write_requests_total"),
		"Total number of bulk write requests executed on the follower",
		ccrShardLabels, nil,
	)
	ccrFailedWriteRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "failed_write_requests_total"),
		"Total number of failed bulk write requests executed on the follower",
		ccrShardLabels, nil,
	)
	ccrOperationsWrittenTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "operations_written_total"),
		"Total number of operations written on the follower",
		ccrShardLabels, nil,
	)
	ccrReadExceptions = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "read_exceptions"),
		"Number of failed reads currently being retried by the follower task",
		ccrShardLabels, nil,
	)
	ccrTimeSinceLastReadSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "time_since_last_read_seconds"),
		"Time since a read request was last sent to the leader",
		ccrShardLabels, nil,
	)
	ccrFatalException = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "fatal_exception"),
		"Whether the follower task stopped because of a fatal exception",
		ccrShardLabels, nil,
	)

	ccrAutoFollowSuccessfulFollowIndicesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "auto_follow_successful_follow_indices_total"),
		"Total number of indices the auto-follow coordinator successfully followed",
		nil, nil,
	)
	ccrAutoFollowFailedFollowIndicesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "auto_follow_failed_follow_indices_total"),
		"Total number of indices the auto-follow coordinator failed to follow",
		nil, nil,
	)
	ccrAutoFollowFailedRemoteClusterStateRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "auto_follow_failed_remote_cluster_state_requests_total"),
		"Total number of times the auto-follow coordinator failed to fetch the remote cluster state",
		nil, nil,
	)
	ccrAutoFollowRecentErrors = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "auto_follow_recent_errors"),
		"Number of recent auto-follow errors reported by the auto-follow coordinator",
		nil, nil,
	)
	ccrAutoFollowTimeSinceLastCheckSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "auto_follow_time_since_last_check_seconds"),
		"Time since the auto-follow coordinator last checked the remote cluster for new indices",
		[]string{"remote_cluster"}, nil,
	)
	ccrAutoFollowLastSeenMetadataVersion = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "auto_follow_last_seen_metadata_version"),
		"Last cluster metadata version of the remote cluster seen by the auto-follow coordinator",
		[]string{"remote_cluster"}, nil,
	)

	ccrFollowerIndexStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr", "follower_index_status"),
		"Status of the follower index (active or paused)",
		[]string{"remote_cluster", "leader_index", "follower_index", "status"}, nil,
	)
)

func init() {
	registerCollector("ccr", defaultDisabled, NewCCR)
}

// CCR information struct
type CCR struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

// NewCCR defines CCR Prometheus metrics
func NewCCR(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &CCR{
		logger: logger,
		hc:     hc,
		u:      u,
	}, nil
}

// ccrStatsResponse is a representation of the cross-cluster replication stats
type ccrStatsResponse struct {
	AutoFollowStats ccrAutoFollowStats `json:"auto_follow_stats"`
	FollowStats     ccrFollowStats     `json:"follow_stats"`
}

// ccrAutoFollowStats is a representation of the auto-follow coordinator stats
type ccrAutoFollowStats struct {
	NumberOfFailedFollowIndices              int64                    `json:"number_of_failed_follow_indices"`
	NumberOfFailedRemoteClusterStateRequests int64                    `json:"number_of_failed_remote_cluster_state_requests"`
	NumberOfSuccessfulFollowIndices          int64                    `json:"number_of_successful_follow_indices"`
	RecentAutoFollowErrors                   []json.RawMessage        `json:"recent_auto_follow_errors"`
	AutoFollowedClusters                     []ccrAutoFollowedCluster `json:"auto_followed_clusters"`
}

// ccrAutoFollowedCluster is a representation of a remote cluster watched by the auto-follow coordinator
type ccrAutoFollowedCluster struct {
	ClusterName              string `json:"cluster_name"`
	TimeSinceLastCheckMillis int64  `json:"time_since_last_check_millis"`
	LastSeenMetadataVersion  int64  `json:"last_seen_metadata_version"`
}

// ccrFollowStats is a representation of the shard-level stats of all follower indices
type ccrFollowStats struct {
	Indices []ccrFollowIndexStats `json:"indices"`
}

// ccrFollowIndexStats is a representation of the stats of a single follower index
type ccrFollowIndexStats struct {
	Index  string                `json:"index"`
	Shards []ccrFollowShardStats `json:"shards"`
}

// ccrFollowShardStats is a representation of the stats of a single shard follow task
type ccrFollowShardStats struct {
	RemoteCluster                 string            `json:"remote_cluster"`
	LeaderIndex                   string            `json:"leader_index"`
	FollowerIndex                 string            `json:"follower_index"`
	ShardID                       int64             `json:"shard_id"`
	LeaderGlobalCheckpoint        int64             `json:"leader_global_checkpoint"`
	LeaderMaxSeqNo                int64             `json:"leader_max_seq_no"`
	FollowerGlobalCheckpoint      int64             `json:"follower_global_checkpoint"`
	FollowerMaxSeqNo              int64             `json:"follower_max_seq_no"`
	LastRequestedSeqNo            int64             `json:"last_requested_seq_no"`
	OutstandingReadRequests       int64             `json:"outstanding_read_requests"`
	OutstandingWriteRequests      int64             `json:"outstanding_write_requests"`
	WriteBufferOperationCount     int64             `json:"write_buffer_operation_count"`
	WriteBufferSizeInBytes        int64             `json:"write_buffer_size_in_bytes"`
	FollowerMappingVersion        int64             `json:"follower_mapping_version"`
	FollowerSettingsVersion       int64             `json:"follower_settings_version"`
	FollowerAliasesVersion        int64             `json:"follower_aliases_version"`
	TotalReadTimeMillis           int64             `json:"total_read_time_millis"`
	TotalReadRemoteExecTimeMillis int64             `json:"total_read_remote_exec_time_millis"`
	SuccessfulReadRequests        int64             `json:"successful_read_requests"`
	FailedReadRequests            int64             `json:"failed_read_requests"`
	OperationsRead                int64             `json:"operations_read"`
	BytesRead                     int64             `json:"bytes_read"`
	TotalWriteTimeMillis          int64             `json:"total_write_time_millis"`
	SuccessfulWriteRequests       int64             `json:"successful_write_requests"`
	FailedWriteRequests           int64             `json:"failed_write_requests"`
	OperationsWritten             int64             `json:"operations_written"`
	ReadExceptions                []json.RawMessage `json:"read_exceptions"`
	TimeSinceLastReadMillis       int64             `json:"time_since_last_read_millis"`
	FatalException                json.RawMessage   `json:"fatal_exception"`
}

// ccrFollowInfoResponse is a representation of the follower info of all follower indices
type ccrFollowInfoResponse struct {
	FollowerIndices []ccrFollowerIndexInfo `json:"follower_indices"`
}

// ccrFollowerIndexInfo is a representation of the follower info of a single follower index
type ccrFollowerIndexInfo struct {
	FollowerIndex string `json:"follower_index"`
	RemoteCluster string `json:"remote_cluster"`
	LeaderIndex   string `json:"leader_index"`
	Status        string `json:"status"`
}

func (c *CCR) Update(ctx context.Context, uc UpdateContext, ch chan<- prometheus.Metric) error {
	u := c.u.JoinPath("_ccr", "stats")
	var ccrStatsResp ccrStatsResponse

	if err := getAndDecodeURL(ctx, c.hc, c.logger, u.String(), &ccrStatsResp); err != nil {
		return err
	}

	// Paused follower indices are absent from the stats response, so the
	// follower info is needed to report them.
	u = c.u.JoinPath("_all", "_ccr", "info")
	var ccrInfoResp ccrFollowInfoResponse

	if err := getAndDecodeURL(ctx, c.hc, c.logger, u.String(), &ccrInfoResp); err != nil {
		return err
	}

	autoFollow := ccrStatsResp.AutoFollowStats

	ch <- prometheus.MustNewConstMetric(
		ccrAutoFollowSuccessfulFollowIndicesTotal,
		prometheus.CounterValue,
		float64(autoFollow.NumberOfSuccessfulFollowIndices),
	)
	ch <- prometheus.MustNewConstMetric(
		ccrAutoFollowFailedFollowIndicesTotal,
		prometheus.CounterValue,
		float64(autoFollow.NumberOfFailedFollowIndices),
	)
	ch <- prometheus.MustNewConstMetric(
		ccrAutoFollowFailedRemoteClusterStateRequestsTotal,
		prometheus.CounterValue,
		float64(autoFollow.NumberOfFailedRemoteClusterStateRequests),
	)
	ch <- prometheus.MustNewConstMetric(
		ccrAutoFollowRecentErrors,
		prometheus.GaugeValue,
		float64(len(autoFollow.RecentAutoFollowErrors)),
	)

	for _, cluster := range autoFollow.AutoFollowedClusters {
		ch <- prometheus.MustNewConstMetric(
			ccrAutoFollowTimeSinceLastCheckSeconds,
			prometheus.GaugeValue,
			float64(cluster.TimeSinceLastCheckMillis)/1000,
			cluster.ClusterName,
		)
		ch <- prometheus.MustNewConstMetric(
			ccrAutoFollowLastSeenMetadataVersion,
			prometheus.GaugeValue,
			float64(cluster.LastSeenMetadataVersion),
			cluster.ClusterName,
		)
	}

	for _, index := range ccrStatsResp.FollowStats.Indices {
		for _, shard := range index.Shards {
			labels := []string{
				shard.RemoteCluster,
				shard.LeaderIndex,
				shard.FollowerIndex,
				strconv.FormatInt(shard.ShardID, 10),
			}

			ch <- prometheus.MustNewConstMetric(
				ccrLeaderGlobalCheckpoint,
				prometheus.GaugeValue,
				float64(shard.LeaderGlobalCheckpoint),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrLeaderMaxSeqNo,
				prometheus.GaugeValue,
				float64(shard.LeaderMaxSeqNo),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrFollowerGlobalCheckpoint,
				prometheus.GaugeValue,
				float64(shard.FollowerGlobalCheckpoint),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrFollowerMaxSeqNo,
				prometheus.GaugeValue,
				float64(shard.FollowerMaxSeqNo),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrLastRequestedSeqNo,
				prometheus.GaugeValue,
				float64(shard.LastRequestedSeqNo),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrOutstandingReadRequests,
				prometheus.GaugeValue,
				float64(shard.OutstandingReadRequests),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrOutstandingWriteRequests,
				prometheus.GaugeValue,
				float64(shard.OutstandingWriteRequests),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrWriteBufferOperations,
				prometheus.GaugeValue,
				float64(shard.WriteBufferOperationCount),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrWriteBufferSizeBytes,
				prometheus.GaugeValue,
				float64(shard.WriteBufferSizeInBytes),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrFollowerMappingVersion,
				prometheus.GaugeValue,
				float64(shard.FollowerMappingVersion),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrFollowerSettingsVersion,
				prometheus.GaugeValue,
				float64(shard.FollowerSettingsVersion),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrFollowerAliasesVersion,
				prometheus.GaugeValue,
				float64(shard.FollowerAliasesVersion),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrReadTimeSecondsTotal,
				prometheus.CounterValue,
				float64(shard.TotalReadTimeMillis)/1000,
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrReadRemoteExecTimeSecondsTotal,
				prometheus.CounterValue,
				float64(shard.TotalReadRemoteExecTimeMillis)/1000,
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrSuccessfulReadRequestsTotal,
				prometheus.CounterValue,
				float64(shard.SuccessfulReadRequests),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrFailedReadRequestsTotal,
				prometheus.CounterValue,
				float64(shard.FailedReadRequests),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrOperationsReadTotal,
				prometheus.CounterValue,
				float64(shard.OperationsRead),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrReadBytesTotal,
				prometheus.CounterValue,
				float64(shard.BytesRead),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrWriteTimeSecondsTotal,
				prometheus.CounterValue,
				float64(shard.TotalWriteTimeMillis)/1000,
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrSuccessfulWriteRequestsTotal,
				prometheus.CounterValue,
				float64(shard.SuccessfulWriteRequests),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrFailedWriteRequestsTotal,
				prometheus.CounterValue,
				float64(shard.FailedWriteRequests),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrOperationsWrittenTotal,
				prometheus.CounterValue,
				float64(shard.OperationsWritten),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrReadExceptions,
				prometheus.GaugeValue,
				float64(len(shard.ReadExceptions)),
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrTimeSinceLastReadSeconds,
				prometheus.GaugeValue,
				float64(shard.TimeSinceLastReadMillis)/1000,
				labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				ccrFatalException,
				prometheus.GaugeValue,
				bool2Float(len(shard.FatalException) > 0 && string(shard.FatalException) != "null"),
				labels...,
			)
		}
	}

	for _, follower := range ccrInfoResp.FollowerIndices {
		for _, status := range ccrFollowerStatuses {
			var value float64
			if follower.Status == status {
				value = 1
			}
			ch <- prometheus.MustNewConstMetric(
				ccrFollowerIndexStatus,
				prometheus.GaugeValue,
				value,
				follower.RemoteCluster,
				follower.LeaderIndex,
				follower.FollowerIndex,
				status,
			)
		}
	}

	return nil
}
