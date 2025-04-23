// Copyright 2024 The Prometheus Authors
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

			s := NewShards(promslog.NewNopLogger(), http.DefaultClient, u)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(s, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})

	}
}
