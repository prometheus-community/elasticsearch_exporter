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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func TestClusterInfo(t *testing.T) {
	// Testcases created using:
	//  docker run -p 9200:9200 -e "discovery.type=single-node" elasticsearch:${VERSION}
	//  curl http://localhost:9200/ > fixtures/cluster_info/${VERSION}.json

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "2.4.5",
			file: "../fixtures/clusterinfo/2.4.5.json",
			want: `# HELP elasticsearch_version Elasticsearch version information.
            # TYPE elasticsearch_version gauge
            elasticsearch_version{build_date="",build_hash="c849dd13904f53e63e88efc33b2ceeda0b6a1276",cluster="elasticsearch",cluster_uuid="3qps7bcWTqyzV49ApmPVfw",lucene_version="5.5.4",version="2.4.5"} 1
      `,
		},
		{
			name: "5.4.2",
			file: "../fixtures/clusterinfo/5.4.2.json",
			want: `# HELP elasticsearch_version Elasticsearch version information.
            # TYPE elasticsearch_version gauge
            elasticsearch_version{build_date="2017-06-15T02:29:28.122Z",build_hash="929b078",cluster="elasticsearch",cluster_uuid="kbqi7yhQT-WlPdGL2m0xJg",lucene_version="6.5.1",version="5.4.2"} 1
      `,
		},
		{
			name: "7.13.1",
			file: "../fixtures/clusterinfo/7.13.1.json",
			want: `# HELP elasticsearch_version Elasticsearch version information.
            # TYPE elasticsearch_version gauge
            elasticsearch_version{build_date="2021-05-28T17:40:59.346932922Z",build_hash="9a7758028e4ea59bcab41c12004603c5a7dd84a9",cluster="docker-cluster",cluster_uuid="aCMrCY1VQpqJ6U4Sw_xdiw",lucene_version="8.8.2",version="7.13.1"} 1
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

			c, err := NewClusterInfo(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestClusterInfoHTTPStatus(t *testing.T) {
	body, err := os.ReadFile("../fixtures/clusterinfo/7.13.1.json")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr string
	}{
		{
			name:   "ok_guard",
			status: http.StatusOK,
			body:   string(body),
			want: `# HELP elasticsearch_version Elasticsearch version information.
# TYPE elasticsearch_version gauge
elasticsearch_version{build_date="2021-05-28T17:40:59.346932922Z",build_hash="9a7758028e4ea59bcab41c12004603c5a7dd84a9",cluster="docker-cluster",cluster_uuid="aCMrCY1VQpqJ6U4Sw_xdiw",lucene_version="8.8.2",version="7.13.1"} 1
`,
		},
		{
			name:    "forbidden",
			status:  http.StatusForbidden,
			body:    `{"error":{"type":"exception","reason":"missing privilege"},"status":403}`,
			wantErr: "HTTP Request failed with code 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				if _, err := io.WriteString(w, tt.body); err != nil {
					t.Error(err)
				}
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}
			c, err := NewClusterInfo(promslog.NewNopLogger(), u, ts.Client())
			if err != nil {
				t.Fatal(err)
			}

			ch := make(chan prometheus.Metric, len(clusterInfoDesc))
			err = c.Update(context.Background(), &mockUpdateContext{}, ch)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Update() error = %v, want nil", err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Errorf("Update() error = %v, want %q", err, tt.wantErr)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Error(err)
			}
		})
	}
}
