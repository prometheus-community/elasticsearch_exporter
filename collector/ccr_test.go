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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func TestCCR(t *testing.T) {
	// Testcases created using two single-node clusters (leader and follower) on 8.17.0
	// with xpack.license.self_generated.type=trial:
	//  curl -XPUT http://follower:9200/_cluster/settings -H 'Content-Type: application/json' -d '{"persistent":{"cluster":{"remote":{"leader":{"seeds":["leader:9300"]}}}}}'
	//  curl -XPUT http://follower:9200/orders-replica/_ccr/follow -H 'Content-Type: application/json' -d '{"remote_cluster":"leader","leader_index":"orders"}'
	//  (same for audit-replica and doomed-replica)
	//  curl -XPOST http://follower:9200/audit-replica/_ccr/pause_follow
	//  curl -XDELETE http://leader:9200/doomed (the follower task stops with a fatal exception)
	//  curl -XPUT http://follower:9200/_ccr/auto_follow/metrics -H 'Content-Type: application/json' -d '{"remote_cluster":"leader","leader_index_patterns":["metrics_*"],"follow_index_pattern":"{{leader_index}}_copy"}'
	//  (pre-creating metrics_clash_copy on the follower makes auto-following metrics_clash fail)
	//  curl http://follower:9200/_ccr/stats (Numbers manually tweaked so that every value is distinct)
	//  curl http://follower:9200/_all/_ccr/info

	tests := []struct {
		name   string
		file   string
		prefix string
		want   string
	}{
		{
			name: "8.17.0",
			file: "8.17.0.json",
			want: `# HELP elasticsearch_ccr_auto_follow_failed_follow_indices_total Total number of indices the auto-follow coordinator failed to follow
			# TYPE elasticsearch_ccr_auto_follow_failed_follow_indices_total counter
			elasticsearch_ccr_auto_follow_failed_follow_indices_total 3
			# HELP elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total Total number of times the auto-follow coordinator failed to fetch the remote cluster state
			# TYPE elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total counter
			elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total 4
			# HELP elasticsearch_ccr_auto_follow_last_seen_metadata_version Last cluster metadata version of the remote cluster seen by the auto-follow coordinator
			# TYPE elasticsearch_ccr_auto_follow_last_seen_metadata_version gauge
			elasticsearch_ccr_auto_follow_last_seen_metadata_version{remote_cluster="leader"} 80
			# HELP elasticsearch_ccr_auto_follow_recent_errors Number of recent auto-follow errors reported by the auto-follow coordinator
			# TYPE elasticsearch_ccr_auto_follow_recent_errors gauge
			elasticsearch_ccr_auto_follow_recent_errors 2
			# HELP elasticsearch_ccr_auto_follow_successful_follow_indices_total Total number of indices the auto-follow coordinator successfully followed
			# TYPE elasticsearch_ccr_auto_follow_successful_follow_indices_total counter
			elasticsearch_ccr_auto_follow_successful_follow_indices_total 7
			# HELP elasticsearch_ccr_auto_follow_time_since_last_check_seconds Time since the auto-follow coordinator last checked the remote cluster for new indices
			# TYPE elasticsearch_ccr_auto_follow_time_since_last_check_seconds gauge
			elasticsearch_ccr_auto_follow_time_since_last_check_seconds{remote_cluster="leader"} 2.012
			# HELP elasticsearch_ccr_failed_read_requests_total Total number of failed reads from the leader
			# TYPE elasticsearch_ccr_failed_read_requests_total counter
			elasticsearch_ccr_failed_read_requests_total{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 2
			elasticsearch_ccr_failed_read_requests_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 7
			elasticsearch_ccr_failed_read_requests_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 17
			# HELP elasticsearch_ccr_failed_write_requests_total Total number of failed bulk write requests executed on the follower
			# TYPE elasticsearch_ccr_failed_write_requests_total counter
			elasticsearch_ccr_failed_write_requests_total{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 0
			elasticsearch_ccr_failed_write_requests_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 8
			elasticsearch_ccr_failed_write_requests_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 19
			# HELP elasticsearch_ccr_fatal_exception Whether the follower task stopped because of a fatal exception
			# TYPE elasticsearch_ccr_fatal_exception gauge
			elasticsearch_ccr_fatal_exception{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 1
			elasticsearch_ccr_fatal_exception{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 0
			elasticsearch_ccr_fatal_exception{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 0
			# HELP elasticsearch_ccr_follower_aliases_version Index aliases version the follower is synced up to
			# TYPE elasticsearch_ccr_follower_aliases_version gauge
			elasticsearch_ccr_follower_aliases_version{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 1
			elasticsearch_ccr_follower_aliases_version{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 6
			elasticsearch_ccr_follower_aliases_version{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 13
			# HELP elasticsearch_ccr_follower_global_checkpoint Current global checkpoint on the follower
			# TYPE elasticsearch_ccr_follower_global_checkpoint gauge
			elasticsearch_ccr_follower_global_checkpoint{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 19
			elasticsearch_ccr_follower_global_checkpoint{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 1490
			elasticsearch_ccr_follower_global_checkpoint{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 1608
			# HELP elasticsearch_ccr_follower_index_status Status of the follower index (active or paused)
			# TYPE elasticsearch_ccr_follower_index_status gauge
			elasticsearch_ccr_follower_index_status{follower_index="audit-replica",leader_index="audit",remote_cluster="leader",status="active"} 0
			elasticsearch_ccr_follower_index_status{follower_index="audit-replica",leader_index="audit",remote_cluster="leader",status="paused"} 1
			elasticsearch_ccr_follower_index_status{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",status="active"} 1
			elasticsearch_ccr_follower_index_status{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",status="paused"} 0
			elasticsearch_ccr_follower_index_status{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",status="active"} 1
			elasticsearch_ccr_follower_index_status{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",status="paused"} 0
			# HELP elasticsearch_ccr_follower_mapping_version Index mapping version the follower is synced up to
			# TYPE elasticsearch_ccr_follower_mapping_version gauge
			elasticsearch_ccr_follower_mapping_version{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 2
			elasticsearch_ccr_follower_mapping_version{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 5
			elasticsearch_ccr_follower_mapping_version{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 3
			# HELP elasticsearch_ccr_follower_max_seq_no Current maximum sequence number on the follower
			# TYPE elasticsearch_ccr_follower_max_seq_no gauge
			elasticsearch_ccr_follower_max_seq_no{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 20
			elasticsearch_ccr_follower_max_seq_no{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 1502
			elasticsearch_ccr_follower_max_seq_no{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 1611
			# HELP elasticsearch_ccr_follower_settings_version Index settings version the follower is synced up to
			# TYPE elasticsearch_ccr_follower_settings_version gauge
			elasticsearch_ccr_follower_settings_version{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 1
			elasticsearch_ccr_follower_settings_version{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 4
			elasticsearch_ccr_follower_settings_version{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 11
			# HELP elasticsearch_ccr_last_requested_seq_no Starting sequence number of the last batch of operations requested from the leader
			# TYPE elasticsearch_ccr_last_requested_seq_no gauge
			elasticsearch_ccr_last_requested_seq_no{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 22
			elasticsearch_ccr_last_requested_seq_no{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 1503
			elasticsearch_ccr_last_requested_seq_no{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 1609
			# HELP elasticsearch_ccr_leader_global_checkpoint Current global checkpoint on the leader known to the follower task
			# TYPE elasticsearch_ccr_leader_global_checkpoint gauge
			elasticsearch_ccr_leader_global_checkpoint{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 21
			elasticsearch_ccr_leader_global_checkpoint{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 1536
			elasticsearch_ccr_leader_global_checkpoint{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 1610
			# HELP elasticsearch_ccr_leader_max_seq_no Current maximum sequence number on the leader known to the follower task
			# TYPE elasticsearch_ccr_leader_max_seq_no gauge
			elasticsearch_ccr_leader_max_seq_no{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 23
			elasticsearch_ccr_leader_max_seq_no{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 1540
			elasticsearch_ccr_leader_max_seq_no{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 1612
			# HELP elasticsearch_ccr_operations_read_total Total number of operations read from the leader
			# TYPE elasticsearch_ccr_operations_read_total counter
			elasticsearch_ccr_operations_read_total{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 20
			elasticsearch_ccr_operations_read_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 1495
			elasticsearch_ccr_operations_read_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 1605
			# HELP elasticsearch_ccr_operations_written_total Total number of operations written on the follower
			# TYPE elasticsearch_ccr_operations_written_total counter
			elasticsearch_ccr_operations_written_total{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 19
			elasticsearch_ccr_operations_written_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 1489
			elasticsearch_ccr_operations_written_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 1601
			# HELP elasticsearch_ccr_outstanding_read_requests Number of active read requests from the follower
			# TYPE elasticsearch_ccr_outstanding_read_requests gauge
			elasticsearch_ccr_outstanding_read_requests{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 0
			elasticsearch_ccr_outstanding_read_requests{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 3
			elasticsearch_ccr_outstanding_read_requests{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 1
			# HELP elasticsearch_ccr_outstanding_write_requests Number of active bulk write requests on the follower
			# TYPE elasticsearch_ccr_outstanding_write_requests gauge
			elasticsearch_ccr_outstanding_write_requests{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 0
			elasticsearch_ccr_outstanding_write_requests{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 2
			elasticsearch_ccr_outstanding_write_requests{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 4
			# HELP elasticsearch_ccr_read_bytes_total Estimated total bytes read from the leader, not accounting for compression
			# TYPE elasticsearch_ccr_read_bytes_total counter
			elasticsearch_ccr_read_bytes_total{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 952
			elasticsearch_ccr_read_bytes_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 145678
			elasticsearch_ccr_read_bytes_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 156789
			# HELP elasticsearch_ccr_read_exceptions Number of failed reads currently being retried by the follower task
			# TYPE elasticsearch_ccr_read_exceptions gauge
			elasticsearch_ccr_read_exceptions{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 1
			elasticsearch_ccr_read_exceptions{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 1
			elasticsearch_ccr_read_exceptions{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 0
			# HELP elasticsearch_ccr_read_remote_exec_time_seconds_total Total time reads spent executing on the remote cluster
			# TYPE elasticsearch_ccr_read_remote_exec_time_seconds_total counter
			elasticsearch_ccr_read_remote_exec_time_seconds_total{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 0.014
			elasticsearch_ccr_read_remote_exec_time_seconds_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 23.456
			elasticsearch_ccr_read_remote_exec_time_seconds_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 8.765
			# HELP elasticsearch_ccr_read_time_seconds_total Total time reads were outstanding, from sending the request to the leader until the reply was returned to the follower
			# TYPE elasticsearch_ccr_read_time_seconds_total counter
			elasticsearch_ccr_read_time_seconds_total{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 10.102
			elasticsearch_ccr_read_time_seconds_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 123.456
			elasticsearch_ccr_read_time_seconds_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 98.765
			# HELP elasticsearch_ccr_successful_read_requests_total Total number of successful fetches from the leader
			# TYPE elasticsearch_ccr_successful_read_requests_total counter
			elasticsearch_ccr_successful_read_requests_total{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 9
			elasticsearch_ccr_successful_read_requests_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 87
			elasticsearch_ccr_successful_read_requests_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 91
			# HELP elasticsearch_ccr_successful_write_requests_total Total number of bulk write requests executed on the follower
			# TYPE elasticsearch_ccr_successful_write_requests_total counter
			elasticsearch_ccr_successful_write_requests_total{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 6
			elasticsearch_ccr_successful_write_requests_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 64
			elasticsearch_ccr_successful_write_requests_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 72
			# HELP elasticsearch_ccr_time_since_last_read_seconds Time since a read request was last sent to the leader
			# TYPE elasticsearch_ccr_time_since_last_read_seconds gauge
			elasticsearch_ccr_time_since_last_read_seconds{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 83.21
			elasticsearch_ccr_time_since_last_read_seconds{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 1.25
			elasticsearch_ccr_time_since_last_read_seconds{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 0.87
			# HELP elasticsearch_ccr_write_buffer_operations Number of write operations queued on the follower
			# TYPE elasticsearch_ccr_write_buffer_operations gauge
			elasticsearch_ccr_write_buffer_operations{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 0
			elasticsearch_ccr_write_buffer_operations{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 12
			elasticsearch_ccr_write_buffer_operations{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 2
			# HELP elasticsearch_ccr_write_buffer_size_bytes Total size of operations queued for writing on the follower
			# TYPE elasticsearch_ccr_write_buffer_size_bytes gauge
			elasticsearch_ccr_write_buffer_size_bytes{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 0
			elasticsearch_ccr_write_buffer_size_bytes{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 4096
			elasticsearch_ccr_write_buffer_size_bytes{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 1024
			# HELP elasticsearch_ccr_write_time_seconds_total Total time spent writing on the follower
			# TYPE elasticsearch_ccr_write_time_seconds_total counter
			elasticsearch_ccr_write_time_seconds_total{follower_index="doomed-replica",leader_index="doomed",remote_cluster="leader",shard="0"} 0.005
			elasticsearch_ccr_write_time_seconds_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="0"} 3.456
			elasticsearch_ccr_write_time_seconds_total{follower_index="orders-replica",leader_index="orders",remote_cluster="leader",shard="1"} 4.321
`,
		},
		{
			name: "8.17.0 no followers",
			file: "8.17.0-no-followers.json",
			want: `# HELP elasticsearch_ccr_auto_follow_failed_follow_indices_total Total number of indices the auto-follow coordinator failed to follow
			# TYPE elasticsearch_ccr_auto_follow_failed_follow_indices_total counter
			elasticsearch_ccr_auto_follow_failed_follow_indices_total 0
			# HELP elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total Total number of times the auto-follow coordinator failed to fetch the remote cluster state
			# TYPE elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total counter
			elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total 0
			# HELP elasticsearch_ccr_auto_follow_recent_errors Number of recent auto-follow errors reported by the auto-follow coordinator
			# TYPE elasticsearch_ccr_auto_follow_recent_errors gauge
			elasticsearch_ccr_auto_follow_recent_errors 0
			# HELP elasticsearch_ccr_auto_follow_successful_follow_indices_total Total number of indices the auto-follow coordinator successfully followed
			# TYPE elasticsearch_ccr_auto_follow_successful_follow_indices_total counter
			elasticsearch_ccr_auto_follow_successful_follow_indices_total 0
`,
		},
		{
			name:   "8.17.0 no followers with path prefix",
			file:   "8.17.0-no-followers.json",
			prefix: "/elasticsearch",
			want: `# HELP elasticsearch_ccr_auto_follow_failed_follow_indices_total Total number of indices the auto-follow coordinator failed to follow
			# TYPE elasticsearch_ccr_auto_follow_failed_follow_indices_total counter
			elasticsearch_ccr_auto_follow_failed_follow_indices_total 0
			# HELP elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total Total number of times the auto-follow coordinator failed to fetch the remote cluster state
			# TYPE elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total counter
			elasticsearch_ccr_auto_follow_failed_remote_cluster_state_requests_total 0
			# HELP elasticsearch_ccr_auto_follow_recent_errors Number of recent auto-follow errors reported by the auto-follow coordinator
			# TYPE elasticsearch_ccr_auto_follow_recent_errors gauge
			elasticsearch_ccr_auto_follow_recent_errors 0
			# HELP elasticsearch_ccr_auto_follow_successful_follow_indices_total Total number of indices the auto-follow coordinator successfully followed
			# TYPE elasticsearch_ccr_auto_follow_successful_follow_indices_total counter
			elasticsearch_ccr_auto_follow_successful_follow_indices_total 0
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fStatsPath := path.Join("../fixtures/ccr/stats/", tt.file)
			fStats, err := os.Open(fStatsPath)
			if err != nil {
				t.Fatal(err)
			}
			defer fStats.Close()

			fInfoPath := path.Join("../fixtures/ccr/info/", tt.file)
			fInfo, err := os.Open(fInfoPath)
			if err != nil {
				t.Fatal(err)
			}
			defer fInfo.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case tt.prefix + "/_ccr/stats":
					io.Copy(w, fStats)
					return
				case tt.prefix + "/_all/_ccr/info":
					io.Copy(w, fInfo)
					return
				}

				http.Error(w, "Not Found", http.StatusNotFound)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL + tt.prefix)
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

func TestCCRUnlicensed(t *testing.T) {
	// Clusters without a license that includes CCR reject the stats request.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"type":"security_exception","reason":"current license is non-compliant for [ccr]"},"status":403}`, http.StatusForbidden)
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

	ch := make(chan prometheus.Metric, 100)
	if err := c.Update(context.Background(), &mockUpdateContext{}, ch); err == nil {
		t.Fatal("expected an error for a forbidden response, got nil")
	}
	close(ch)

	if n := len(ch); n != 0 {
		t.Fatalf("expected no metrics for a forbidden response, got %d", n)
	}
}
