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
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

// clusterInfoResponse returns a minimal Elasticsearch root response for the given cluster.
func clusterInfoResponse(clusterName, clusterUUID, version, buildHash, buildDate, luceneVersion string) string {
	return `{
  "name": "node-1",
  "cluster_name": "` + clusterName + `",
  "cluster_uuid": "` + clusterUUID + `",
  "version": {
    "number": "` + version + `",
    "build_hash": "` + buildHash + `",
    "build_date": "` + buildDate + `",
    "build_snapshot": false,
    "lucene_version": "` + luceneVersion + `"
  },
  "tagline": "You Know, for Search"
}`
}

// setupClusterInfoState enables the cluster-info collector in the global
// collectorState (which is normally populated by kingpin.Parse) so the
// test can call NewElasticsearchCollector without a full application bootstrap.
// The returned cleanup function restores the original value.
func setupClusterInfoState(t *testing.T) func() {
	t.Helper()
	enabled := true
	originalState := collectorState["cluster-info"]
	collectorState["cluster-info"] = &enabled
	return func() {
		collectorState["cluster-info"] = originalState
	}
}

// TestNewElasticsearchCollectorMultiTarget verifies that successive calls to
// NewElasticsearchCollector with different target URLs each create a fresh
// collector instance bound to their own URL, so that multi-target probe
// requests return the correct elasticsearch_version metric for each target.
func TestNewElasticsearchCollectorMultiTarget(t *testing.T) {
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(clusterInfoResponse(
			"cluster-1", "uuid-cluster-1", "7.13.1", "abc123",
			"2021-05-28T17:40:59.346932922Z", "8.8.2",
		)))
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(clusterInfoResponse(
			"cluster-2", "uuid-cluster-2", "8.10.0", "def456",
			"2023-09-07T00:00:00.000Z", "9.7.0",
		)))
	}))
	defer ts2.Close()

	cleanup := setupClusterInfoState(t)
	defer cleanup()

	u1, err := url.Parse(ts1.URL)
	if err != nil {
		t.Fatal(err)
	}
	u2, err := url.Parse(ts2.URL)
	if err != nil {
		t.Fatal(err)
	}

	logger := promslog.NewNopLogger()

	// Simulate probe request for target 1.
	exp1, err := NewElasticsearchCollector(logger, []string{},
		WithElasticsearchURL(u1),
		WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate probe request for target 2.
	exp2, err := NewElasticsearchCollector(logger, []string{},
		WithElasticsearchURL(u2),
		WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		t.Fatal(err)
	}

	c1, ok := exp1.Collectors["cluster-info"]
	if !ok {
		t.Fatal("cluster-info collector not found in exp1")
	}
	c2, ok := exp2.Collectors["cluster-info"]
	if !ok {
		t.Fatal("cluster-info collector not found in exp2")
	}

	// Each call must produce a distinct collector instance.
	if c1 == c2 {
		t.Error("expected distinct cluster-info collector instances for different targets, but got the same instance")
	}

	// Verify cluster-1 returns its own version.
	want1 := `# HELP elasticsearch_version Elasticsearch version information.
# TYPE elasticsearch_version gauge
elasticsearch_version{build_date="2021-05-28T17:40:59.346932922Z",build_hash="abc123",cluster="cluster-1",cluster_uuid="uuid-cluster-1",lucene_version="8.8.2",version="7.13.1"} 1
`
	if err := testutil.CollectAndCompare(wrapCollector{c1}, strings.NewReader(want1)); err != nil {
		t.Errorf("cluster-1 metrics mismatch: %v", err)
	}

	// Verify cluster-2 returns its own version (not cluster-1's data).
	want2 := `# HELP elasticsearch_version Elasticsearch version information.
# TYPE elasticsearch_version gauge
elasticsearch_version{build_date="2023-09-07T00:00:00.000Z",build_hash="def456",cluster="cluster-2",cluster_uuid="uuid-cluster-2",lucene_version="9.7.0",version="8.10.0"} 1
`
	if err := testutil.CollectAndCompare(wrapCollector{c2}, strings.NewReader(want2)); err != nil {
		t.Errorf("cluster-2 metrics mismatch: %v", err)
	}
}
