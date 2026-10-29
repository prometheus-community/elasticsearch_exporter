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
	"errors"
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

func TestShards(t *testing.T) {
	// Testcases created using:
	// docker run --rm -d -p 9200:9200 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:$VERSION
	// curl -XPUT http://localhost:9200/testindex
	// curl -XPUT http://localhost:9200/otherindex
	// curl http://localhost:9200/_cat/shards?format=json > fixtures/shards/$VERSION.json

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "7.15.0",
			file: "7.15.0.json",
			want: `# HELP elasticsearch_node_shards_json_parse_failures Number of errors while parsing JSON.
             # TYPE elasticsearch_node_shards_json_parse_failures counter
             elasticsearch_node_shards_json_parse_failures 0
             # HELP elasticsearch_node_shards_total Total shards per node
             # TYPE elasticsearch_node_shards_total gauge
             elasticsearch_node_shards_total{cluster="unknown_cluster",node="35dfca79831a"} 3
						 `,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(path.Join("../fixtures/shards/", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(w, f)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}

			s, err := NewShards(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{s}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// A response cut off right after a complete shard object must count as a
// parse failure and emit no partial per-node totals.
func TestShardsTruncatedResponse(t *testing.T) {
	fixture, err := os.ReadFile("../fixtures/shards/7.15.0.json")
	if err != nil {
		t.Fatal(err)
	}
	truncated := strings.TrimSuffix(strings.TrimSpace(string(fixture)), "]")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, truncated)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %s", err)
	}

	s, err := NewShards(promslog.NewNopLogger(), u, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	want := `# HELP elasticsearch_node_shards_json_parse_failures Number of errors while parsing JSON.
             # TYPE elasticsearch_node_shards_json_parse_failures counter
             elasticsearch_node_shards_json_parse_failures 1
						 `
	if err := testutil.CollectAndCompare(wrapCollector{s}, strings.NewReader(want)); err != nil {
		t.Fatal(err)
	}
}

func TestShardsClusterName(t *testing.T) {
	tests := []struct {
		name string
		uc   *mockUpdateContext
		want string
	}{
		{
			name: "cluster name from update context",
			uc:   &mockUpdateContext{clusterName: "test-cluster"},
			want: `# HELP elasticsearch_node_shards_json_parse_failures Number of errors while parsing JSON.
             # TYPE elasticsearch_node_shards_json_parse_failures counter
             elasticsearch_node_shards_json_parse_failures 0
             # HELP elasticsearch_node_shards_total Total shards per node
             # TYPE elasticsearch_node_shards_total gauge
             elasticsearch_node_shards_total{cluster="test-cluster",node="35dfca79831a"} 3
						 `,
		},
		{
			name: "cluster info error falls back to unknown_cluster",
			uc:   &mockUpdateContext{err: errors.New("cluster info unavailable")},
			want: `# HELP elasticsearch_node_shards_json_parse_failures Number of errors while parsing JSON.
             # TYPE elasticsearch_node_shards_json_parse_failures counter
             elasticsearch_node_shards_json_parse_failures 0
             # HELP elasticsearch_node_shards_total Total shards per node
             # TYPE elasticsearch_node_shards_total gauge
             elasticsearch_node_shards_total{cluster="unknown_cluster",node="35dfca79831a"} 3
						 `,
		},
	}

	fixture, err := os.ReadFile("../fixtures/shards/7.15.0.json")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/_cat/shards" {
					http.NotFound(w, r)
					return
				}
				w.Write(fixture)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}

			s, err := NewShards(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			// The scrape must still succeed when cluster info is unavailable.
			ch := make(chan prometheus.Metric, 10)
			if err := s.Update(context.Background(), tt.uc, ch); err != nil {
				t.Fatalf("Update returned an error: %s", err)
			}

			if err := testutil.CollectAndCompare(wrapCollectorWithContext{s, tt.uc}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestShardsHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewShards(promslog.NewNopLogger(), u, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan prometheus.Metric, 10)
	err = c.Update(context.Background(), &mockUpdateContext{}, ch)
	close(ch)
	if err == nil {
		t.Fatal("expected an error for a non-200 response, got nil")
	}

	// Only the json_parse_failures counter is emitted, and it is not
	// incremented because the body was never decoded.
	if n := len(ch); n != 1 {
		t.Fatalf("expected only json_parse_failures on error, got %d metrics", n)
	}
	jsonParseFailures := c.(*Shards).jsonParseFailures
	if m := <-ch; m != jsonParseFailures {
		t.Fatalf("expected json_parse_failures, got %s", m.Desc())
	}
	if v := testutil.ToFloat64(jsonParseFailures); v != 0 {
		t.Fatalf("expected json_parse_failures 0, got %v", v)
	}
}
