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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/log"
)

func TestSLM(t *testing.T) {
	// Testcases created using:

	//  docker run -d -p 9200:9200 -e discovery.type=single-node -e path.repo=/tmp/backups docker.elastic.co/elasticsearch/elasticsearch:7.15.0-arm64
	//  curl -XPUT http://127.0.0.1:9200/_snapshot/my_repository -H 'Content-Type: application/json' -d '{"type":"url","settings":{"url":"file:/tmp/backups"}}'
	//  curl -XPUT http://127.0.0.1:9200/_slm/policy/everything -H 'Content-Type: application/json' -d '{"schedule":"0 */15 * * * ?","name":"<everything-{now/d}>","repository":"my_repository","config":{"indices":".*","include_global_state":true,"ignore_unavailable":true},"retention":{"expire_after":"7d"}}'
	//  curl http://127.0.0.1:9200/_slm/stats (Numbers manually tweaked)

	tcs := map[string]string{
		"7.15.0": `{"retention_runs":9,"retention_failed":0,"retention_timed_out":0,"retention_deletion_time":"1.2m","retention_deletion_time_millis":72491,"total_snapshots_taken":103,"total_snapshots_failed":2,"total_snapshots_deleted":20,"total_snapshot_deletion_failures":0,"policy_stats":[{"policy":"everything","snapshots_taken":50,"snapshots_failed":2,"snapshots_deleted":20,"snapshot_deletion_failures":0}]}`,
	}
	for ver, out := range tcs {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, out)
		}))
		defer ts.Close()

		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatalf("Failed to parse URL: %s", err)
		}
		s := NewSLM(log.NewNopLogger(), http.DefaultClient, u)
		stats, err := s.fetchAndDecodeSLMStats()
		if err != nil {
			t.Fatalf("Failed to fetch or decode snapshots stats: %s", err)
		}
		t.Logf("[%s] SLM Response: %+v", ver, stats)
		slmStats := stats
		policyStats := stats.PolicyStats[0]

		if slmStats.TotalSnapshotsTaken != 103 {
			t.Errorf("Bad number of total snapshots taken")
		}

		if policyStats.SnapshotsTaken != 50 {
			t.Errorf("Bad number of policy snapshots taken")
		}
	}

}
