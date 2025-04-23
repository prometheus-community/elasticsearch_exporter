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

func TestSLM(t *testing.T) {
	// Testcases created using:

	//  docker run -d -p 9200:9200 -e discovery.type=single-node -e path.repo=/tmp/backups docker.elastic.co/elasticsearch/elasticsearch:7.15.0-arm64
	//  curl -XPUT http://127.0.0.1:9200/_snapshot/my_repository -H 'Content-Type: application/json' -d '{"type":"url","settings":{"url":"file:/tmp/backups"}}'
	//  curl -XPUT http://127.0.0.1:9200/_slm/policy/everything -H 'Content-Type: application/json' -d '{"schedule":"0 */15 * * * ?","name":"<everything-{now/d}>","repository":"my_repository","config":{"indices":".*","include_global_state":true,"ignore_unavailable":true},"retention":{"expire_after":"7d"}}'
	//  curl http://127.0.0.1:9200/_slm/stats (Numbers manually tweaked)

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "7.15.0",
			file: "7.15.0.json",
			want: `# HELP elasticsearch_slm_stats_operation_mode Operating status of SLM
            # TYPE elasticsearch_slm_stats_operation_mode gauge
            elasticsearch_slm_stats_operation_mode{operation_mode="RUNNING"} 0
            elasticsearch_slm_stats_operation_mode{operation_mode="STOPPED"} 0
            elasticsearch_slm_stats_operation_mode{operation_mode="STOPPING"} 0
            # HELP elasticsearch_slm_stats_retention_deletion_time_seconds Retention run deletion time
            # TYPE elasticsearch_slm_stats_retention_deletion_time_seconds gauge
            elasticsearch_slm_stats_retention_deletion_time_seconds 72.491
            # HELP elasticsearch_slm_stats_retention_failed_total Total failed retention runs
            # TYPE elasticsearch_slm_stats_retention_failed_total counter
            elasticsearch_slm_stats_retention_failed_total 0
            # HELP elasticsearch_slm_stats_retention_runs_total Total retention runs
            # TYPE elasticsearch_slm_stats_retention_runs_total counter
            elasticsearch_slm_stats_retention_runs_total 9
            # HELP elasticsearch_slm_stats_retention_timed_out_total Total timed out retention runs
            # TYPE elasticsearch_slm_stats_retention_timed_out_total counter
            elasticsearch_slm_stats_retention_timed_out_total 0
            # HELP elasticsearch_slm_stats_snapshot_deletion_failures_total Total snapshot deletion failures
            # TYPE elasticsearch_slm_stats_snapshot_deletion_failures_total counter
            elasticsearch_slm_stats_snapshot_deletion_failures_total{policy="everything"} 0
            # HELP elasticsearch_slm_stats_snapshots_deleted_total Total snapshots deleted
            # TYPE elasticsearch_slm_stats_snapshots_deleted_total counter
            elasticsearch_slm_stats_snapshots_deleted_total{policy="everything"} 20
            # HELP elasticsearch_slm_stats_snapshots_failed_total Total snapshots failed
            # TYPE elasticsearch_slm_stats_snapshots_failed_total counter
            elasticsearch_slm_stats_snapshots_failed_total{policy="everything"} 2
            # HELP elasticsearch_slm_stats_snapshots_taken_total Total snapshots taken
            # TYPE elasticsearch_slm_stats_snapshots_taken_total counter
            elasticsearch_slm_stats_snapshots_taken_total{policy="everything"} 50
            # HELP elasticsearch_slm_stats_total_snapshot_deletion_failures_total Total snapshot deletion failures
            # TYPE elasticsearch_slm_stats_total_snapshot_deletion_failures_total counter
            elasticsearch_slm_stats_total_snapshot_deletion_failures_total 0
            # HELP elasticsearch_slm_stats_total_snapshots_deleted_total Total snapshots deleted
            # TYPE elasticsearch_slm_stats_total_snapshots_deleted_total counter
            elasticsearch_slm_stats_total_snapshots_deleted_total 20
            # HELP elasticsearch_slm_stats_total_snapshots_failed_total Total snapshots failed
            # TYPE elasticsearch_slm_stats_total_snapshots_failed_total counter
            elasticsearch_slm_stats_total_snapshots_failed_total 2
            # HELP elasticsearch_slm_stats_total_snapshots_taken_total Total snapshots taken
            # TYPE elasticsearch_slm_stats_total_snapshots_taken_total counter
            elasticsearch_slm_stats_total_snapshots_taken_total 103
						`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fStatsPath := path.Join("../fixtures/slm/stats/", tt.file)
			fStats, err := os.Open(fStatsPath)
			if err != nil {
				t.Fatal(err)
			}
			defer fStats.Close()

			fStatusPath := path.Join("../fixtures/slm/status/", tt.file)
			fStatus, err := os.Open(fStatusPath)
			if err != nil {
				t.Fatal(err)
			}
			defer fStatus.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/_slm/stats":
					io.Copy(w, fStats)
					return
				case "/_slm/status":
					io.Copy(w, fStatus)
					return
				}

				http.Error(w, "Not Found", http.StatusNotFound)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}

			s, err := NewSLM(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{s}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})

	}

}
