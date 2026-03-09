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

func TestCCR(t *testing.T) {
	statsFile, err := os.Open(path.Join("../fixtures/ccr/stats", "7.17.0.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer statsFile.Close()

	infoFile, err := os.Open(path.Join("../fixtures/ccr/info", "7.17.0.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer infoFile.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/_ccr/stats":
			io.Copy(w, statsFile)
			return
		case "/_all/_ccr/info":
			io.Copy(w, infoFile)
			return
		}

		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewCCR(promslog.NewNopLogger(), u, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	const expected = `
		# HELP elasticsearch_ccr_auto_follow_failed_follow_indices_total Number of indices that auto-follow failed to follow
		# TYPE elasticsearch_ccr_auto_follow_failed_follow_indices_total counter
		elasticsearch_ccr_auto_follow_failed_follow_indices_total 2
		# HELP elasticsearch_ccr_auto_followed_cluster_last_seen_metadata_version Last seen metadata version for an auto-followed cluster
		# TYPE elasticsearch_ccr_auto_followed_cluster_last_seen_metadata_version gauge
		elasticsearch_ccr_auto_followed_cluster_last_seen_metadata_version{remote_cluster="remote_a"} 451
		# HELP elasticsearch_ccr_follow_index_global_checkpoint_lag Total global checkpoint lag for a follower index
		# TYPE elasticsearch_ccr_follow_index_global_checkpoint_lag gauge
		elasticsearch_ccr_follow_index_global_checkpoint_lag{follower_index="follower_index"} 256
		# HELP elasticsearch_ccr_follow_shard_successful_read_requests_total Successful read requests
		# TYPE elasticsearch_ccr_follow_shard_successful_read_requests_total counter
		elasticsearch_ccr_follow_shard_successful_read_requests_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 32
		# HELP elasticsearch_ccr_follow_shard_total_read_time_seconds_total Total read time in seconds
		# TYPE elasticsearch_ccr_follow_shard_total_read_time_seconds_total counter
		elasticsearch_ccr_follow_shard_total_read_time_seconds_total{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",shard_id="0"} 32.768
		# HELP elasticsearch_ccr_follower_index_status Follower index status where 1 means current state
		# TYPE elasticsearch_ccr_follower_index_status gauge
		elasticsearch_ccr_follower_index_status{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",status="active"} 1
		elasticsearch_ccr_follower_index_status{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a",status="paused"} 0
		elasticsearch_ccr_follower_index_status{follower_index="follower_paused",leader_index="leader_paused",remote_cluster="remote_b",status="active"} 0
		elasticsearch_ccr_follower_index_status{follower_index="follower_paused",leader_index="leader_paused",remote_cluster="remote_b",status="paused"} 1
		# HELP elasticsearch_ccr_follower_parameters_max_outstanding_read_requests Max outstanding read requests configured for a follower index
		# TYPE elasticsearch_ccr_follower_parameters_max_outstanding_read_requests gauge
		elasticsearch_ccr_follower_parameters_max_outstanding_read_requests{follower_index="follower_index",leader_index="leader_index",remote_cluster="remote_a"} 12
	`

	if err := testutil.CollectAndCompare(
		wrapCollector{c},
		strings.NewReader(expected),
		"elasticsearch_ccr_auto_follow_failed_follow_indices_total",
		"elasticsearch_ccr_auto_followed_cluster_last_seen_metadata_version",
		"elasticsearch_ccr_follow_index_global_checkpoint_lag",
		"elasticsearch_ccr_follow_shard_successful_read_requests_total",
		"elasticsearch_ccr_follow_shard_total_read_time_seconds_total",
		"elasticsearch_ccr_follower_index_status",
		"elasticsearch_ccr_follower_parameters_max_outstanding_read_requests",
	); err != nil {
		t.Fatal(err)
	}
}
