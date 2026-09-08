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
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ccrDetailedMetrics bool

	ccrAutoFollowFailedFollowIndicesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_auto_follow", "failed_follow_indices_total"),
		"Number of indices that auto-follow failed to follow",
		nil,
		nil,
	)
	ccrAutoFollowFailedRemoteClusterStateRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_auto_follow", "failed_remote_cluster_state_requests_total"),
		"Number of failed remote cluster state requests from auto-follow",
		nil,
		nil,
	)
	ccrAutoFollowSuccessfulFollowIndicesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_auto_follow", "successful_follow_indices_total"),
		"Number of indices auto-follow successfully followed",
		nil,
		nil,
	)
	ccrAutoFollowRecentErrors = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_auto_follow", "recent_errors"),
		"Number of recent auto-follow errors currently reported",
		nil,
		nil,
	)
	ccrAutoFollowedClusterLastSeenMetadataVersion = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_auto_followed_cluster", "last_seen_metadata_version"),
		"Last seen metadata version for an auto-followed cluster",
		[]string{"remote_cluster"},
		nil,
	)
	ccrAutoFollowedClusterTimeSinceLastCheckSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_auto_followed_cluster", "time_since_last_check_seconds"),
		"Time since last auto-follow check in seconds for an auto-followed cluster",
		[]string{"remote_cluster"},
		nil,
	)
	ccrFollowIndexGlobalCheckpointLag = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_index", "global_checkpoint_lag"),
		"Total global checkpoint lag for a follower index",
		[]string{"follower_index"},
		nil,
	)
	ccrFollowShardLeaderGlobalCheckpoint = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "leader_global_checkpoint"),
		"Leader global checkpoint",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardLeaderMaxSeqNo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "leader_max_seq_no"),
		"Leader max sequence number",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardFollowerGlobalCheckpoint = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "follower_global_checkpoint"),
		"Follower global checkpoint",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardFollowerMaxSeqNo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "follower_max_seq_no"),
		"Follower max sequence number",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardLastRequestedSeqNo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "last_requested_seq_no"),
		"Last requested sequence number",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardOutstandingReadRequests = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "outstanding_read_requests"),
		"Outstanding read requests",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardOutstandingWriteRequests = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "outstanding_write_requests"),
		"Outstanding write requests",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardWriteBufferOperationCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "write_buffer_operation_count"),
		"Write buffer operation count",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardFollowerMappingVersion = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "follower_mapping_version"),
		"Follower mapping version",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardFollowerSettingsVersion = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "follower_settings_version"),
		"Follower settings version",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardFollowerAliasesVersion = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "follower_aliases_version"),
		"Follower aliases version",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardTotalReadTimeSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "total_read_time_seconds_total"),
		"Total read time in seconds",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardTotalReadRemoteExecTimeSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "total_read_remote_exec_time_seconds_total"),
		"Total remote read execution time in seconds",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardSuccessfulReadRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "successful_read_requests_total"),
		"Successful read requests",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardFailedReadRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "failed_read_requests_total"),
		"Failed read requests",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardOperationsReadTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "operations_read_total"),
		"Read operations",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardBytesReadTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "bytes_read_total"),
		"Read bytes",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardTotalWriteTimeSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "total_write_time_seconds_total"),
		"Total write time in seconds",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardWriteBufferSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "write_buffer_size_bytes"),
		"Write buffer size in bytes",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardSuccessfulWriteRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "successful_write_requests_total"),
		"Successful write requests",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardFailedWriteRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "failed_write_requests_total"),
		"Failed write requests",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardOperationsWrittenTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "operations_written_total"),
		"Write operations",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardTimeSinceLastReadSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "time_since_last_read_seconds"),
		"Time since last read in seconds",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowShardReadExceptionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follow_shard", "read_exceptions_total"),
		"Number of read exceptions",
		[]string{"remote_cluster", "leader_index", "follower_index", "shard_id"},
		nil,
	)
	ccrFollowerStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follower_index", "status"),
		"Follower index status where 1 means current state",
		[]string{"follower_index", "leader_index", "remote_cluster", "status"},
		nil,
	)
	ccrFollowerParamsMaxReadRequestOperationCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follower_parameters", "max_read_request_operation_count"),
		"Max read request operation count configured for a follower index",
		[]string{"follower_index", "leader_index", "remote_cluster"},
		nil,
	)
	ccrFollowerParamsMaxOutstandingReadRequests = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follower_parameters", "max_outstanding_read_requests"),
		"Max outstanding read requests configured for a follower index",
		[]string{"follower_index", "leader_index", "remote_cluster"},
		nil,
	)
	ccrFollowerParamsMaxWriteRequestOperationCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follower_parameters", "max_write_request_operation_count"),
		"Max write request operation count configured for a follower index",
		[]string{"follower_index", "leader_index", "remote_cluster"},
		nil,
	)
	ccrFollowerParamsMaxOutstandingWriteRequests = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follower_parameters", "max_outstanding_write_requests"),
		"Max outstanding write requests configured for a follower index",
		[]string{"follower_index", "leader_index", "remote_cluster"},
		nil,
	)
	ccrFollowerParamsMaxWriteBufferCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ccr_follower_parameters", "max_write_buffer_count"),
		"Max write buffer count configured for a follower index",
		[]string{"follower_index", "leader_index", "remote_cluster"},
		nil,
	)
)

func init() {
	kingpin.Flag(
		"collector.ccr.detailed",
		"Enable high-cardinality CCR metrics (per-shard and follower parameter metrics).",
	).Default("false").BoolVar(&ccrDetailedMetrics)

	registerCollector("ccr", defaultDisabled, NewCCR)
}

// CCR collects metrics from Elasticsearch cross-cluster replication APIs.
type CCR struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

func NewCCR(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &CCR{
		logger: logger,
		hc:     hc,
		u:      u,
	}, nil
}

type CCRStatsResponse struct {
	AutoFollowStats CCRAutoFollowStats `json:"auto_follow_stats"`
	FollowStats     CCRFollowStats     `json:"follow_stats"`
}

type CCRAutoFollowStats struct {
	NumberOfFailedFollowIndices             int64                     `json:"number_of_failed_follow_indices"`
	NumberOfFailedRemoteClusterStateRequest int64                     `json:"number_of_failed_remote_cluster_state_requests"`
	NumberOfSuccessfulFollowIndices         int64                     `json:"number_of_successful_follow_indices"`
	RecentAutoFollowErrors                  []CCRAutoFollowError      `json:"recent_auto_follow_errors"`
	AutoFollowedClusters                    []CCRAutoFollowedClusters `json:"auto_followed_clusters"`
}

type CCRAutoFollowError struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type CCRAutoFollowedClusters struct {
	ClusterName             string `json:"cluster_name"`
	LastSeenMetadataVersion int64  `json:"last_seen_metadata_version"`
	TimeSinceLastCheckMs    int64  `json:"time_since_last_check_millis"`
}

type CCRFollowStats struct {
	Indices []CCRFollowIndexStatsResponse `json:"indices"`
}

type CCRFollowIndexStatsResponse struct {
	Index                    string                  `json:"index"`
	TotalGlobalCheckpointLag int64                   `json:"total_global_checkpoint_lag"`
	Shards                   []CCRShardStatsResponse `json:"shards"`
}

type CCRShardStatsResponse struct {
	RemoteCluster                 string        `json:"remote_cluster"`
	LeaderIndex                   string        `json:"leader_index"`
	FollowerIndex                 string        `json:"follower_index"`
	ShardID                       int64         `json:"shard_id"`
	LeaderGlobalCheckpoint        int64         `json:"leader_global_checkpoint"`
	LeaderMaxSeqNo                int64         `json:"leader_max_seq_no"`
	FollowerGlobalCheckpoint      int64         `json:"follower_global_checkpoint"`
	FollowerMaxSeqNo              int64         `json:"follower_max_seq_no"`
	LastRequestedSeqNo            int64         `json:"last_requested_seq_no"`
	OutstandingReadRequests       int64         `json:"outstanding_read_requests"`
	OutstandingWriteRequests      int64         `json:"outstanding_write_requests"`
	WriteBufferOperationCount     int64         `json:"write_buffer_operation_count"`
	FollowerMappingVersion        int64         `json:"follower_mapping_version"`
	FollowerSettingsVersion       int64         `json:"follower_settings_version"`
	FollowerAliasesVersion        int64         `json:"follower_aliases_version"`
	TotalReadTimeMillis           int64         `json:"total_read_time_millis"`
	TotalReadRemoteExecTimeMillis int64         `json:"total_read_remote_exec_time_millis"`
	SuccessfulReadRequests        int64         `json:"successful_read_requests"`
	FailedReadRequests            int64         `json:"failed_read_requests"`
	OperationsRead                int64         `json:"operations_read"`
	BytesRead                     int64         `json:"bytes_read"`
	TotalWriteTimeMillis          int64         `json:"total_write_time_millis"`
	WriteBufferSizeBytes          int64         `json:"write_buffer_size_in_bytes"`
	SuccessfulWriteRequests       int64         `json:"successful_write_requests"`
	FailedWriteRequests           int64         `json:"failed_write_requests"`
	OperationsWritten             int64         `json:"operations_written"`
	ReadExceptions                []interface{} `json:"read_exceptions"`
	TimeSinceLastReadMillis       int64         `json:"time_since_last_read_millis"`
}

type CCRFollowInfoResponse struct {
	FollowerIndices []CCRFollowerIndexInfo `json:"follower_indices"`
}

type CCRFollowerIndexInfo struct {
	FollowerIndex string               `json:"follower_index"`
	LeaderIndex   string               `json:"leader_index"`
	RemoteCluster string               `json:"remote_cluster"`
	Status        string               `json:"status"`
	Parameters    *CCRFollowerSettings `json:"parameters"`
}

type CCRFollowerSettings struct {
	MaxReadRequestOperationCount  int64 `json:"max_read_request_operation_count"`
	MaxOutstandingReadRequests    int64 `json:"max_outstanding_read_requests"`
	MaxWriteRequestOperationCount int64 `json:"max_write_request_operation_count"`
	MaxOutstandingWriteRequests   int64 `json:"max_outstanding_write_requests"`
	MaxWriteBufferCount           int64 `json:"max_write_buffer_count"`
}

func (c *CCR) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	statsURL := c.u.ResolveReference(&url.URL{Path: "/_ccr/stats"})
	statsResp, err := getURL(ctx, c.hc, c.logger, statsURL.String())
	if err != nil {
		return fmt.Errorf("failed to load CCR stats: %w", err)
	}

	var stats CCRStatsResponse
	if err := json.Unmarshal(statsResp, &stats); err != nil {
		return fmt.Errorf("failed to decode CCR stats response: %w", err)
	}

	infoURL := c.u.ResolveReference(&url.URL{Path: "/_all/_ccr/info"})
	infoResp, err := getURL(ctx, c.hc, c.logger, infoURL.String())
	if err != nil {
		return fmt.Errorf("failed to load CCR follow info: %w", err)
	}

	var followInfo CCRFollowInfoResponse
	if err := json.Unmarshal(infoResp, &followInfo); err != nil {
		return fmt.Errorf("failed to decode CCR follow info response: %w", err)
	}

	c.collectAutoFollowStats(ch, stats.AutoFollowStats)
	c.collectFollowStats(ch, stats.FollowStats)
	c.collectFollowerInfo(ch, followInfo)

	return nil
}

func (c *CCR) collectAutoFollowStats(ch chan<- prometheus.Metric, stats CCRAutoFollowStats) {
	ch <- prometheus.MustNewConstMetric(
		ccrAutoFollowFailedFollowIndicesTotal,
		prometheus.CounterValue,
		float64(stats.NumberOfFailedFollowIndices),
	)
	ch <- prometheus.MustNewConstMetric(
		ccrAutoFollowFailedRemoteClusterStateRequestsTotal,
		prometheus.CounterValue,
		float64(stats.NumberOfFailedRemoteClusterStateRequest),
	)
	ch <- prometheus.MustNewConstMetric(
		ccrAutoFollowSuccessfulFollowIndicesTotal,
		prometheus.CounterValue,
		float64(stats.NumberOfSuccessfulFollowIndices),
	)
	ch <- prometheus.MustNewConstMetric(
		ccrAutoFollowRecentErrors,
		prometheus.GaugeValue,
		float64(len(stats.RecentAutoFollowErrors)),
	)

	for _, cluster := range stats.AutoFollowedClusters {
		ch <- prometheus.MustNewConstMetric(
			ccrAutoFollowedClusterLastSeenMetadataVersion,
			prometheus.GaugeValue,
			float64(cluster.LastSeenMetadataVersion),
			cluster.ClusterName,
		)
		ch <- prometheus.MustNewConstMetric(
			ccrAutoFollowedClusterTimeSinceLastCheckSeconds,
			prometheus.GaugeValue,
			float64(cluster.TimeSinceLastCheckMs)/1000,
			cluster.ClusterName,
		)
	}
}

func (c *CCR) collectFollowStats(ch chan<- prometheus.Metric, stats CCRFollowStats) {
	for _, indexStats := range stats.Indices {
		ch <- prometheus.MustNewConstMetric(
			ccrFollowIndexGlobalCheckpointLag,
			prometheus.GaugeValue,
			float64(indexStats.TotalGlobalCheckpointLag),
			indexStats.Index,
		)

		if !ccrDetailedMetrics {
			continue
		}

		for _, shard := range indexStats.Shards {
			shardLabels := []string{
				shard.RemoteCluster,
				shard.LeaderIndex,
				shard.FollowerIndex,
				strconv.FormatInt(shard.ShardID, 10),
			}
			ch <- prometheus.MustNewConstMetric(ccrFollowShardLeaderGlobalCheckpoint, prometheus.GaugeValue, float64(shard.LeaderGlobalCheckpoint), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardLeaderMaxSeqNo, prometheus.GaugeValue, float64(shard.LeaderMaxSeqNo), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardFollowerGlobalCheckpoint, prometheus.GaugeValue, float64(shard.FollowerGlobalCheckpoint), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardFollowerMaxSeqNo, prometheus.GaugeValue, float64(shard.FollowerMaxSeqNo), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardLastRequestedSeqNo, prometheus.GaugeValue, float64(shard.LastRequestedSeqNo), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardOutstandingReadRequests, prometheus.GaugeValue, float64(shard.OutstandingReadRequests), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardOutstandingWriteRequests, prometheus.GaugeValue, float64(shard.OutstandingWriteRequests), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardWriteBufferOperationCount, prometheus.GaugeValue, float64(shard.WriteBufferOperationCount), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardFollowerMappingVersion, prometheus.GaugeValue, float64(shard.FollowerMappingVersion), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardFollowerSettingsVersion, prometheus.GaugeValue, float64(shard.FollowerSettingsVersion), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardFollowerAliasesVersion, prometheus.GaugeValue, float64(shard.FollowerAliasesVersion), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardTotalReadTimeSecondsTotal, prometheus.CounterValue, float64(shard.TotalReadTimeMillis)/1000, shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardTotalReadRemoteExecTimeSecondsTotal, prometheus.CounterValue, float64(shard.TotalReadRemoteExecTimeMillis)/1000, shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardSuccessfulReadRequestsTotal, prometheus.CounterValue, float64(shard.SuccessfulReadRequests), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardFailedReadRequestsTotal, prometheus.CounterValue, float64(shard.FailedReadRequests), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardOperationsReadTotal, prometheus.CounterValue, float64(shard.OperationsRead), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardBytesReadTotal, prometheus.CounterValue, float64(shard.BytesRead), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardTotalWriteTimeSecondsTotal, prometheus.CounterValue, float64(shard.TotalWriteTimeMillis)/1000, shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardWriteBufferSizeBytes, prometheus.GaugeValue, float64(shard.WriteBufferSizeBytes), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardSuccessfulWriteRequestsTotal, prometheus.CounterValue, float64(shard.SuccessfulWriteRequests), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardFailedWriteRequestsTotal, prometheus.CounterValue, float64(shard.FailedWriteRequests), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardOperationsWrittenTotal, prometheus.CounterValue, float64(shard.OperationsWritten), shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardTimeSinceLastReadSeconds, prometheus.GaugeValue, float64(shard.TimeSinceLastReadMillis)/1000, shardLabels...)
			ch <- prometheus.MustNewConstMetric(ccrFollowShardReadExceptionsTotal, prometheus.CounterValue, float64(len(shard.ReadExceptions)), shardLabels...)
		}
	}
}

func (c *CCR) collectFollowerInfo(ch chan<- prometheus.Metric, info CCRFollowInfoResponse) {
	followerStatuses := []string{"active", "paused"}
	for _, follower := range info.FollowerIndices {
		for _, status := range followerStatuses {
			ch <- prometheus.MustNewConstMetric(
				ccrFollowerStatus,
				prometheus.GaugeValue,
				bool2Float(follower.Status == status),
				follower.FollowerIndex,
				follower.LeaderIndex,
				follower.RemoteCluster,
				status,
			)
		}

		if !ccrDetailedMetrics {
			continue
		}

		if follower.Parameters == nil {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			ccrFollowerParamsMaxReadRequestOperationCount,
			prometheus.GaugeValue,
			float64(follower.Parameters.MaxReadRequestOperationCount),
			follower.FollowerIndex,
			follower.LeaderIndex,
			follower.RemoteCluster,
		)
		ch <- prometheus.MustNewConstMetric(
			ccrFollowerParamsMaxOutstandingReadRequests,
			prometheus.GaugeValue,
			float64(follower.Parameters.MaxOutstandingReadRequests),
			follower.FollowerIndex,
			follower.LeaderIndex,
			follower.RemoteCluster,
		)
		ch <- prometheus.MustNewConstMetric(
			ccrFollowerParamsMaxWriteRequestOperationCount,
			prometheus.GaugeValue,
			float64(follower.Parameters.MaxWriteRequestOperationCount),
			follower.FollowerIndex,
			follower.LeaderIndex,
			follower.RemoteCluster,
		)
		ch <- prometheus.MustNewConstMetric(
			ccrFollowerParamsMaxOutstandingWriteRequests,
			prometheus.GaugeValue,
			float64(follower.Parameters.MaxOutstandingWriteRequests),
			follower.FollowerIndex,
			follower.LeaderIndex,
			follower.RemoteCluster,
		)
		ch <- prometheus.MustNewConstMetric(
			ccrFollowerParamsMaxWriteBufferCount,
			prometheus.GaugeValue,
			float64(follower.Parameters.MaxWriteBufferCount),
			follower.FollowerIndex,
			follower.LeaderIndex,
			follower.RemoteCluster,
		)
	}
}
