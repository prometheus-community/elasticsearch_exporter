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
// collectorState (which is normally populated by kingpin.Parse) and clears any
// previously cached instance so the test starts from a clean slate.
// The returned cleanup function restores the original values.
func setupClusterInfoState(t *testing.T) func() {
	t.Helper()

	// Override collectorState so cluster-info appears enabled even without
	// kingpin.Parse (which is not called in unit tests).
	enabled := true
	originalState := collectorState["cluster-info"]
	collectorState["cluster-info"] = &enabled

	// Snapshot and clear the global collector cache for this collector.
	initiatedCollectorsMtx.Lock()
	originalCached, hadCached := initiatedCollectors["cluster-info"]
	delete(initiatedCollectors, "cluster-info")
	initiatedCollectorsMtx.Unlock()

	return func() {
		collectorState["cluster-info"] = originalState
		initiatedCollectorsMtx.Lock()
		if hadCached {
			initiatedCollectors["cluster-info"] = originalCached
		} else {
			delete(initiatedCollectors, "cluster-info")
		}
		initiatedCollectorsMtx.Unlock()
	}
}

// TestWithSkipCacheVerifyBug shows that without WithSkipCache(true) the global
// collector cache causes the second target's collector to be the same instance
// as the first target's collector (the bug).
func TestWithSkipCacheVerifyBug(t *testing.T) {
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

	// First call: creates and caches the cluster-info collector for target 1.
	exp1, err := NewElasticsearchCollector(logger, []string{},
		WithElasticsearchURL(u1),
		WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		t.Fatal(err)
	}
	c1, ok := exp1.Collectors["cluster-info"]
	if !ok {
		t.Fatal("cluster-info collector not found in exp1")
	}

	// Second call with a different URL but without skipCache:
	// the cache returns the same collector that was bound to target 1.
	exp2, err := NewElasticsearchCollector(logger, []string{},
		WithElasticsearchURL(u2),
		WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		t.Fatal(err)
	}
	c2, ok := exp2.Collectors["cluster-info"]
	if !ok {
		t.Fatal("cluster-info collector not found in exp2")
	}

	// Without skipCache the two instances must be the same (demonstrating the bug).
	if c1 != c2 {
		t.Error("expected the same (cached) collector instance when skipCache is false, but got distinct instances")
	}
}

// TestWithSkipCacheFixesMultiTarget verifies that WithSkipCache(true) creates a
// fresh collector per call, so each probe request queries its own target and
// returns the correct elasticsearch_version metric for that target.
func TestWithSkipCacheFixesMultiTarget(t *testing.T) {
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

	// Probe for target 1 with skipCache=true.
	exp1, err := NewElasticsearchCollector(logger, []string{},
		WithElasticsearchURL(u1),
		WithHTTPClient(http.DefaultClient),
		WithSkipCache(true),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Probe for target 2 with skipCache=true.
	exp2, err := NewElasticsearchCollector(logger, []string{},
		WithElasticsearchURL(u2),
		WithHTTPClient(http.DefaultClient),
		WithSkipCache(true),
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

	// With skipCache=true each call must produce a distinct collector instance.
	if c1 == c2 {
		t.Error("WithSkipCache(true) should create distinct collector instances for different targets, but the same instance was returned")
	}

	// Verify collector 1 returns the correct version for cluster-1.
	want1 := `# HELP elasticsearch_version Elasticsearch version information.
# TYPE elasticsearch_version gauge
elasticsearch_version{build_date="2021-05-28T17:40:59.346932922Z",build_hash="abc123",cluster="cluster-1",cluster_uuid="uuid-cluster-1",lucene_version="8.8.2",version="7.13.1"} 1
`
	if err := testutil.CollectAndCompare(wrapCollector{c1}, strings.NewReader(want1)); err != nil {
		t.Errorf("cluster-1 metrics mismatch: %v", err)
	}

	// Verify collector 2 returns the correct version for cluster-2.
	// Before the fix this would show cluster-1's data due to cache reuse.
	want2 := `# HELP elasticsearch_version Elasticsearch version information.
# TYPE elasticsearch_version gauge
elasticsearch_version{build_date="2023-09-07T00:00:00.000Z",build_hash="def456",cluster="cluster-2",cluster_uuid="uuid-cluster-2",lucene_version="9.7.0",version="8.10.0"} 1
`
	if err := testutil.CollectAndCompare(wrapCollector{c2}, strings.NewReader(want2)); err != nil {
		t.Errorf("cluster-2 metrics mismatch (cache was not bypassed): %v", err)
	}
}
