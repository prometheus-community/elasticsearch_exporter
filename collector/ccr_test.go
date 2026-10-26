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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func TestCCRMinimal(t *testing.T) {
	previous := ccrDetailedMetrics
	ccrDetailedMetrics = false
	defer func() {
		ccrDetailedMetrics = previous
	}()

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "7.17.0",
			file: "7.17.0.json",
			want: `# HELP elasticsearch_ccr_auto_follow_failed_follow_indices_total Number of indices that auto-follow failed to follow
# TYPE elasticsearch_ccr_auto_follow_failed_follow_indices_total counter
elasticsearch_ccr_auto_follow_failed_follow_indices_total 2
# HELP elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total Number of failed remote cluster state requests from auto-follow
# TYPE elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total counter
elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total 1
# HELP elasticsearch_ccr_auto_follow_recent_errors Number of recent auto-follow errors currently reported
# TYPE elasticsearch_ccr_auto_follow_recent_errors gauge
elasticsearch_ccr_auto_follow_recent_errors 1
# HELP elasticsearch_ccr_auto_follow_successful_follow_indices_total Number of indices auto-follow successfully followed
# TYPE elasticsearch_ccr_auto_follow_successful_follow_indices_total counter
elasticsearch_ccr_auto_follow_successful_follow_indices_total 8
# HELP elasticsearch_ccr_auto_followed_cluster_last_seen_metadata_version Last seen metadata version for an auto-followed cluster
# TYPE elasticsearch_ccr_auto_followed_cluster_last_seen_metadata_version gauge
elasticsearch_ccr_auto_followed_cluster_last_seen_metadata_version{remote_cluster="remote_a"} 451
# HELP elasticsearch_ccr_auto_followed_cluster_time_since_last_check_seconds Time since last auto-follow check in seconds for an auto-followed cluster
# TYPE elasticsearch_ccr_auto_followed_cluster_time_since_last_check_seconds gauge
elasticsearch_ccr_auto_followed_cluster_time_since_last_check_seconds{remote_cluster="remote_a"} 1.2
# HELP elasticsearch_ccr_follow_index_global_checkpoint_lag Total global checkpoint lag for a follower index
# TYPE elasticsearch_ccr_follow_index_global_checkpoint_lag gauge
elasticsearch_ccr_follow_index_global_checkpoint_lag{follower_index="follower_index"} 256
# HELP elasticsearch_ccr_follower_index_status Follower index status where 1 means current state
# TYPE elasticsearch_ccr_follower_index_status gauge
elasticsearch_ccr_follower_index_status{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",status="active"} 1
elasticsearch_ccr_follower_index_status{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",status="paused"} 0
elasticsearch_ccr_follower_index_status{follower_index="follower_paused",leader_index="leader_paused",remote_cluster="remote_b",status="active"} 0
elasticsearch_ccr_follower_index_status{follower_index="follower_paused",leader_index="leader_paused",remote_cluster="remote_b",status="paused"} 1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fStats, err := os.Open(path.Join("../fixtures/ccr/stats/", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			defer fStats.Close()

			fInfo, err := os.Open(path.Join("../fixtures/ccr/info/", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			defer fInfo.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/_ccr/stats":
					io.Copy(w, fStats)
					return
				case "/_all/_ccr/info":
					io.Copy(w, fInfo)
					return
				}
				http.Error(w, "Not Found", http.StatusNotFound)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}

			c, err := NewCCR(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCCRDetailed(t *testing.T) {
	previous := ccrDetailedMetrics
	ccrDetailedMetrics = true
	defer func() {
		ccrDetailedMetrics = previous
	}()

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "7.17.0",
			file: "7.17.0.json",
			want: `# HELP elasticsearch_ccr_auto_follow_failed_follow_indices_total Number of indices that auto-follow failed to follow
# TYPE elasticsearch_ccr_auto_follow_failed_follow_indices_total counter
elasticsearch_ccr_auto_follow_failed_follow_indices_total 2
# HELP elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total Number of failed remote cluster state requests from auto-follow
# TYPE elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total counter
elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total 1
# HELP elasticsearch_ccr_auto_follow_recent_errors Number of recent auto-follow errors currently reported
# TYPE elasticsearch_ccr_auto_follow_recent_errors gauge
elasticsearch_ccr_auto_follow_recent_errors 1
# HELP elasticsearch_ccr_auto_follow_successful_follow_indices_total Number of indices auto-follow successfully followed
# TYPE elasticsearch_ccr_auto_follow_successful_follow_indices_total counter
elasticsearch_ccr_auto_follow_successful_follow_indices_total 8
# HELP elasticsearch_ccr_auto_followed_cluster_last_seen_metadata_version Last seen metadata version for an auto-followed cluster
# TYPE elasticsearch_ccr_auto_followed_cluster_last_seen_metadata_version gauge
elasticsearch_ccr_auto_followed_cluster_last_seen_metadata_version{remote_cluster="remote_a"} 451
# HELP elasticsearch_ccr_auto_followed_cluster_time_since_last_check_seconds Time since last auto-follow check in seconds for an auto-followed cluster
# TYPE elasticsearch_ccr_auto_followed_cluster_time_since_last_check_seconds gauge
elasticsearch_ccr_auto_followed_cluster_time_since_last_check_seconds{remote_cluster="remote_a"} 1.2
# HELP elasticsearch_ccr_follow_index_global_checkpoint_lag Total global checkpoint lag for a follower index
# TYPE elasticsearch_ccr_follow_index_global_checkpoint_lag gauge
elasticsearch_ccr_follow_index_global_checkpoint_lag{follower_index="follower_index"} 256
# HELP elasticsearch_ccr_follow_shard_bytes_read_total Read bytes
# TYPE elasticsearch_ccr_follow_shard_bytes_read_total counter
elasticsearch_ccr_follow_shard_bytes_read_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 32768
# HELP elasticsearch_ccr_follow_shard_failed_read_requests_total Failed read requests
# TYPE elasticsearch_ccr_follow_shard_failed_read_requests_total counter
elasticsearch_ccr_follow_shard_failed_read_requests_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 0
# HELP elasticsearch_ccr_follow_shard_failed_write_requests_total Failed write requests
# TYPE elasticsearch_ccr_follow_shard_failed_write_requests_total counter
elasticsearch_ccr_follow_shard_failed_write_requests_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 0
# HELP elasticsearch_ccr_follow_shard_follower_aliases_version Follower aliases version
# TYPE elasticsearch_ccr_follow_shard_follower_aliases_version gauge
elasticsearch_ccr_follow_shard_follower_aliases_version{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 8
# HELP elasticsearch_ccr_follow_shard_follower_global_checkpoint Follower global checkpoint
# TYPE elasticsearch_ccr_follow_shard_follower_global_checkpoint gauge
elasticsearch_ccr_follow_shard_follower_global_checkpoint{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 768
# HELP elasticsearch_ccr_follow_shard_follower_mapping_version Follower mapping version
# TYPE elasticsearch_ccr_follow_shard_follower_mapping_version gauge
elasticsearch_ccr_follow_shard_follower_mapping_version{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 4
# HELP elasticsearch_ccr_follow_shard_follower_max_seq_no Follower max sequence number
# TYPE elasticsearch_ccr_follow_shard_follower_max_seq_no gauge
elasticsearch_ccr_follow_shard_follower_max_seq_no{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 896
# HELP elasticsearch_ccr_follow_shard_follower_settings_version Follower settings version
# TYPE elasticsearch_ccr_follow_shard_follower_settings_version gauge
elasticsearch_ccr_follow_shard_follower_settings_version{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 2
# HELP elasticsearch_ccr_follow_shard_last_requested_seq_no Last requested sequence number
# TYPE elasticsearch_ccr_follow_shard_last_requested_seq_no gauge
elasticsearch_ccr_follow_shard_last_requested_seq_no{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 897
# HELP elasticsearch_ccr_follow_shard_leader_global_checkpoint Leader global checkpoint
# TYPE elasticsearch_ccr_follow_shard_leader_global_checkpoint gauge
elasticsearch_ccr_follow_shard_leader_global_checkpoint{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 1024
# HELP elasticsearch_ccr_follow_shard_leader_max_seq_no Leader max sequence number
# TYPE elasticsearch_ccr_follow_shard_leader_max_seq_no gauge
elasticsearch_ccr_follow_shard_leader_max_seq_no{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 1536
# HELP elasticsearch_ccr_follow_shard_operations_read_total Read operations
# TYPE elasticsearch_ccr_follow_shard_operations_read_total counter
elasticsearch_ccr_follow_shard_operations_read_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 896
# HELP elasticsearch_ccr_follow_shard_operations_written_total Write operations
# TYPE elasticsearch_ccr_follow_shard_operations_written_total counter
elasticsearch_ccr_follow_shard_operations_written_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 832
# HELP elasticsearch_ccr_follow_shard_outstanding_read_requests Outstanding read requests
# TYPE elasticsearch_ccr_follow_shard_outstanding_read_requests gauge
elasticsearch_ccr_follow_shard_outstanding_read_requests{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 8
# HELP elasticsearch_ccr_follow_shard_outstanding_write_requests Outstanding write requests
# TYPE elasticsearch_ccr_follow_shard_outstanding_write_requests gauge
elasticsearch_ccr_follow_shard_outstanding_write_requests{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 2
# HELP elasticsearch_ccr_follow_shard_read_exceptions_total Number of read exceptions
# TYPE elasticsearch_ccr_follow_shard_read_exceptions_total counter
elasticsearch_ccr_follow_shard_read_exceptions_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 0
# HELP elasticsearch_ccr_follow_shard_successful_read_requests_total Successful read requests
# TYPE elasticsearch_ccr_follow_shard_successful_read_requests_total counter
elasticsearch_ccr_follow_shard_successful_read_requests_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 32
# HELP elasticsearch_ccr_follow_shard_successful_write_requests_total Successful write requests
# TYPE elasticsearch_ccr_follow_shard_successful_write_requests_total counter
elasticsearch_ccr_follow_shard_successful_write_requests_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 16
# HELP elasticsearch_ccr_follow_shard_time_since_last_read_seconds Time since last read in seconds
# TYPE elasticsearch_ccr_follow_shard_time_since_last_read_seconds gauge
elasticsearch_ccr_follow_shard_time_since_last_read_seconds{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 0.008
# HELP elasticsearch_ccr_follow_shard_total_read_remote_exec_time_seconds_total Total remote read execution time in seconds
# TYPE elasticsearch_ccr_follow_shard_total_read_remote_exec_time_seconds_total counter
elasticsearch_ccr_follow_shard_total_read_remote_exec_time_seconds_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 16.384
# HELP elasticsearch_ccr_follow_shard_total_read_time_seconds_total Total read time in seconds
# TYPE elasticsearch_ccr_follow_shard_total_read_time_seconds_total counter
elasticsearch_ccr_follow_shard_total_read_time_seconds_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 32.768
# HELP elasticsearch_ccr_follow_shard_total_write_time_seconds_total Total write time in seconds
# TYPE elasticsearch_ccr_follow_shard_total_write_time_seconds_total counter
elasticsearch_ccr_follow_shard_total_write_time_seconds_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 16.384
# HELP elasticsearch_ccr_follow_shard_write_buffer_operation_count Write buffer operation count
# TYPE elasticsearch_ccr_follow_shard_write_buffer_operation_count gauge
elasticsearch_ccr_follow_shard_write_buffer_operation_count{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 64
# HELP elasticsearch_ccr_follow_shard_write_buffer_size_bytes Write buffer size in bytes
# TYPE elasticsearch_ccr_follow_shard_write_buffer_size_bytes gauge
elasticsearch_ccr_follow_shard_write_buffer_size_bytes{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 1536
# HELP elasticsearch_ccr_follower_index_status Follower index status where 1 means current state
# TYPE elasticsearch_ccr_follower_index_status gauge
elasticsearch_ccr_follower_index_status{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",status="active"} 1
elasticsearch_ccr_follower_index_status{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",status="paused"} 0
elasticsearch_ccr_follower_index_status{follower_index="follower_paused",leader_index="leader_paused",remote_cluster="remote_b",status="active"} 0
elasticsearch_ccr_follower_index_status{follower_index="follower_paused",leader_index="leader_paused",remote_cluster="remote_b",status="paused"} 1
# HELP elasticsearch_ccr_follower_parameters_max_outstanding_read_requests Max outstanding read requests configured for a follower index
# TYPE elasticsearch_ccr_follower_parameters_max_outstanding_read_requests gauge
elasticsearch_ccr_follower_parameters_max_outstanding_read_requests{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a"} 12
# HELP elasticsearch_ccr_follower_parameters_max_outstanding_write_requests Max outstanding write requests configured for a follower index
# TYPE elasticsearch_ccr_follower_parameters_max_outstanding_write_requests gauge
elasticsearch_ccr_follower_parameters_max_outstanding_write_requests{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a"} 9
# HELP elasticsearch_ccr_follower_parameters_max_read_request_operation_count Max read request operation count configured for a follower index
# TYPE elasticsearch_ccr_follower_parameters_max_read_request_operation_count gauge
elasticsearch_ccr_follower_parameters_max_read_request_operation_count{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a"} 5120
# HELP elasticsearch_ccr_follower_parameters_max_write_buffer_count Max write buffer count configured for a follower index
# TYPE elasticsearch_ccr_follower_parameters_max_write_buffer_count gauge
elasticsearch_ccr_follower_parameters_max_write_buffer_count{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a"} 2.147483647e+09
# HELP elasticsearch_ccr_follower_parameters_max_write_request_operation_count Max write request operation count configured for a follower index
# TYPE elasticsearch_ccr_follower_parameters_max_write_request_operation_count gauge
elasticsearch_ccr_follower_parameters_max_write_request_operation_count{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a"} 5120
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fStats, err := os.Open(path.Join("../fixtures/ccr/stats/", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			defer fStats.Close()

			fInfo, err := os.Open(path.Join("../fixtures/ccr/info/", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			defer fInfo.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/_ccr/stats":
					io.Copy(w, fStats)
					return
				case "/_all/_ccr/info":
					io.Copy(w, fInfo)
					return
				}
				http.Error(w, "Not Found", http.StatusNotFound)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}

			c, err := NewCCR(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
