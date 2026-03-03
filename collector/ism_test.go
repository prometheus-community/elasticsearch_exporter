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

func TestISM(t *testing.T) {
	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "2.12.0",
			file: "2.12.0.json",
			want: `
				# HELP elasticsearch_ism_index_failed Whether ISM is currently in a failed step/action for the index (OpenSearch Index State Management)
				# TYPE elasticsearch_ism_index_failed gauge
				elasticsearch_ism_index_failed{action="rollover",index="test-logs-001",policy_id="test-lifecycle-policy",state="hot",step="attempt_rollover",step_status="starting"} 0
				elasticsearch_ism_index_failed{action="allocation",index="test-logs-002",policy_id="test-lifecycle-policy",state="warm",step="attempt_allocation",step_status="failed"} 1
				# HELP elasticsearch_ism_index_status Status of ISM policy for index (OpenSearch Index State Management)
				# TYPE elasticsearch_ism_index_status gauge
				elasticsearch_ism_index_status{action="rollover",index="test-logs-001",policy_id="test-lifecycle-policy",state="hot",step="attempt_rollover",step_status="starting"} 1
				elasticsearch_ism_index_status{action="allocation",index="test-logs-002",policy_id="test-lifecycle-policy",state="warm",step="attempt_allocation",step_status="failed"} 1
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(path.Join("../fixtures/ism_explain", tt.file))
			if err != nil {
				t.Fatal(err)
			}

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sm := http.NewServeMux()
				sm.HandleFunc("/_plugins/_ism/explain/", func(w http.ResponseWriter, r *http.Request) {
					w.Write(data)
				})
				sm.ServeHTTP(w, r)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			c, err := NewISM(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestISM_FallbackOpendistroEndpoint(t *testing.T) {
	data, err := os.ReadFile(path.Join("../fixtures/ism_explain", "2.12.0.json"))
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sm := http.NewServeMux()
		// Intentionally do NOT register "/_plugins/_ism/explain/" so it 404s.
		sm.HandleFunc("/_opendistro/_ism/explain/", func(w http.ResponseWriter, r *http.Request) {
			w.Write(data)
		})
		sm.ServeHTTP(w, r)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewISM(promslog.NewNopLogger(), u, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	want := `
		# HELP elasticsearch_ism_index_failed Whether ISM is currently in a failed step/action for the index (OpenSearch Index State Management)
		# TYPE elasticsearch_ism_index_failed gauge
		elasticsearch_ism_index_failed{action="rollover",index="test-logs-001",policy_id="test-lifecycle-policy",state="hot",step="attempt_rollover",step_status="starting"} 0
		elasticsearch_ism_index_failed{action="allocation",index="test-logs-002",policy_id="test-lifecycle-policy",state="warm",step="attempt_allocation",step_status="failed"} 1
		# HELP elasticsearch_ism_index_status Status of ISM policy for index (OpenSearch Index State Management)
		# TYPE elasticsearch_ism_index_status gauge
		elasticsearch_ism_index_status{action="rollover",index="test-logs-001",policy_id="test-lifecycle-policy",state="hot",step="attempt_rollover",step_status="starting"} 1
		elasticsearch_ism_index_status{action="allocation",index="test-logs-002",policy_id="test-lifecycle-policy",state="warm",step="attempt_allocation",step_status="failed"} 1
	`

	if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(want)); err != nil {
		t.Fatal(err)
	}
}

