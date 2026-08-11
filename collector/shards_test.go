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
			want: `# HELP elasticsearch_node_shards_state Shard state allocated per node by index (0=unassigned, 10=primary started, 11=primary initializing, 12=primary relocating, 20=replica initializing, 21=replica started, 22=replica relocating).
			# TYPE elasticsearch_node_shards_state gauge
			elasticsearch_node_shards_state{cluster="unknown_cluster",index=".geoip_databases",node="35dfca79831a",shard="0"} 10
			elasticsearch_node_shards_state{cluster="unknown_cluster",index="otherindex",node="35dfca79831a",shard="0"} 10
			elasticsearch_node_shards_state{cluster="unknown_cluster",index="testindex",node="35dfca79831a",shard="0"} 10
			elasticsearch_node_shards_state{cluster="unknown_cluster",index="otherindex",node="-",shard="0"} 0
			elasticsearch_node_shards_state{cluster="unknown_cluster",index="testindex",node="-",shard="0"} 0
			# HELP elasticsearch_node_shards_json_parse_failures Number of errors while parsing JSON.
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

func TestShards_encodeState(t *testing.T) {
	node := "test-node"
	tests := []struct {
		name  string
		shard ShardResponse
		want  float64
	}{
		{
			name:  "unassigned_nil_node",
			shard: ShardResponse{Node: nil, State: "STARTED", Prirep: "p"},
			want:  0,
		},
		{
			name:  "unassigned_state",
			shard: ShardResponse{Node: nil, State: "UNASSIGNED", Prirep: "p"},
			want:  0,
		},
		{
			name:  "primary_started",
			shard: ShardResponse{Node: &node, State: "STARTED", Prirep: "p"},
			want:  10,
		},
		{
			name:  "primary_initializing",
			shard: ShardResponse{Node: &node, State: "INITIALIZING", Prirep: "p"},
			want:  11,
		},
		{
			name:  "primary_relocating",
			shard: ShardResponse{Node: &node, State: "RELOCATING", Prirep: "p"},
			want:  12,
		},
		{
			name:  "replica_started",
			shard: ShardResponse{Node: &node, State: "STARTED", Prirep: "r"},
			want:  20,
		},
		{
			name:  "replica_initializing",
			shard: ShardResponse{Node: &node, State: "INITIALIZING", Prirep: "r"},
			want:  21,
		},
		{
			name:  "replica_relocating",
			shard: ShardResponse{Node: &node, State: "RELOCATING", Prirep: "r"},
			want:  22,
		},
		{
			name:  "unknown_prirep",
			shard: ShardResponse{Node: &node, State: "STARTED", Prirep: "x"},
			want:  0,
		},
		{
			name:  "unknown_state",
			shard: ShardResponse{Node: &node, State: "UNKNOWN", Prirep: "p"},
			want:  0,
		},
	}

	s := &Shards{logger: promslog.NewNopLogger()}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.encodeState(tt.shard); got != tt.want {
				t.Errorf("encodeState() = %v, want %v", got, tt.want)
			}
		})
	}
}
