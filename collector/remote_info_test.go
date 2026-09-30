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
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func TestRemoteInfo(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine
	//  curl http://localhost:9200/_remote/info

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "7.15.0",
			file: "../fixtures/remote_info/7.15.0.json",
			want: `
				# HELP elasticsearch_remote_info_max_connections_per_cluster Maximum number of connections allowed per remote cluster
				# TYPE elasticsearch_remote_info_max_connections_per_cluster gauge
				elasticsearch_remote_info_max_connections_per_cluster{remote_cluster="cluster_remote_1"} 10
				elasticsearch_remote_info_max_connections_per_cluster{remote_cluster="cluster_remote_2"} 5
				# HELP elasticsearch_remote_info_num_nodes_connected Number of nodes connected to the remote cluster
				# TYPE elasticsearch_remote_info_num_nodes_connected gauge
				elasticsearch_remote_info_num_nodes_connected{remote_cluster="cluster_remote_1"} 3
				elasticsearch_remote_info_num_nodes_connected{remote_cluster="cluster_remote_2"} 0
				# HELP elasticsearch_remote_info_num_proxy_sockets_connected Number of proxy sockets connected to the remote cluster
				# TYPE elasticsearch_remote_info_num_proxy_sockets_connected gauge
				elasticsearch_remote_info_num_proxy_sockets_connected{remote_cluster="cluster_remote_1"} 5
				elasticsearch_remote_info_num_proxy_sockets_connected{remote_cluster="cluster_remote_2"} 0
			`,
		},
		{
			name: "8.0.0",
			file: "../fixtures/remote_info/8.0.0.json",
			want: `
				# HELP elasticsearch_remote_info_max_connections_per_cluster Maximum number of connections allowed per remote cluster
				# TYPE elasticsearch_remote_info_max_connections_per_cluster gauge
				elasticsearch_remote_info_max_connections_per_cluster{remote_cluster="prod_cluster"} 30
				# HELP elasticsearch_remote_info_num_nodes_connected Number of nodes connected to the remote cluster
				# TYPE elasticsearch_remote_info_num_nodes_connected gauge
				elasticsearch_remote_info_num_nodes_connected{remote_cluster="prod_cluster"} 15
				# HELP elasticsearch_remote_info_num_proxy_sockets_connected Number of proxy sockets connected to the remote cluster
				# TYPE elasticsearch_remote_info_num_proxy_sockets_connected gauge
				elasticsearch_remote_info_num_proxy_sockets_connected{remote_cluster="prod_cluster"} 25
			`,
		},
		{
			name: "empty",
			file: "../fixtures/remote_info/empty.json",
			want: `
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.file)
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
				t.Fatal(err)
			}

			c, err := NewRemoteInfo(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRemoteInfoError(t *testing.T) {
	// Test error handling when endpoint is unavailable
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewRemoteInfo(promslog.NewNopLogger(), u, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	expected := ``

	if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(expected)); err != nil {
		t.Fatal(err)
	}
}

func TestRemoteInfoJSONParseError(t *testing.T) {
	// Test JSON parse error handling
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewRemoteInfo(promslog.NewNopLogger(), u, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	expected := ``

	if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(expected)); err != nil {
		t.Fatal(err)
	}
}
