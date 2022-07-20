// Copyright 2021 The Prometheus Authors
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
	"testing"

	"github.com/go-kit/log"
)

func TestClusterSettingsStats(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine
	//  curl http://localhost:9200/_cluster/settings/?include_defaults=true
	files := []string{"../fixtures/settings-5.4.2.json", "../fixtures/settings-merge-5.4.2.json"}
	for _, filename := range files {
		f, _ := os.Open(filename)
		defer f.Close()
		for hn, handler := range map[string]http.Handler{
			"plain": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(w, f)
			}),
		} {
			ts := httptest.NewServer(handler)
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}
			c := NewClusterSettings(log.NewNopLogger(), http.DefaultClient, u)
			nsr, err := c.fetchAndDecodeClusterSettingsStats()
			if err != nil {
				t.Fatalf("Failed to fetch or decode cluster settings stats: %s", err)
			}
			t.Logf("[%s/%s] Cluster Settings Stats Response: %+v", hn, filename, nsr)
			if nsr.Cluster.Routing.Allocation.Enabled != "ALL" {
				t.Errorf("Wrong setting for cluster routing allocation enabled")
			}
			if nsr.Cluster.MaxShardsPerNode != nil {
				t.Errorf("MaxShardsPerNode should be empty on older releases")
			}
		}
	}
}

func TestClusterMaxShardsPerNode(t *testing.T) {
	// settings-7.3.0.json testcase created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine
	//  curl http://localhost:9200/_cluster/settings/?include_defaults=true
	// settings-persistent-clustermaxshartspernode-7.17.json testcase created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION
	//  curl -X PUT http://localhost:9200/_cluster/settings -H 'Content-Type: application/json' -d '{"persistent":{"cluster.max_shards_per_node":1000}}'
	//  curl http://localhost:9200/_cluster/settings/?include_defaults=true
	files := []string{"../fixtures/settings-7.3.0.json", "../fixtures/settings-persistent-clustermaxshartspernode-7.17.5.json"}
	for _, filename := range files {
		f, _ := os.Open(filename)
		defer f.Close()
		for hn, handler := range map[string]http.Handler{
			"plain": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(w, f)
			}),
		} {
			ts := httptest.NewServer(handler)
			defer ts.Close()
			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}
			c := NewClusterSettings(log.NewNopLogger(), http.DefaultClient, u)
			nsr, err := c.fetchAndDecodeClusterSettingsStats()
			if err != nil {
				t.Fatalf("Failed to fetch or decode cluster settings stats: %s", err)
			}
			t.Logf("[%s/%s] Cluster Settings Stats Response: %+v", hn, filename, nsr)
			if nsr.Cluster.MaxShardsPerNode != "1000" {
				t.Errorf("Wrong value for MaxShardsPerNode")
			}
		}
	}
}
