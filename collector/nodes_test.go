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
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
)

func TestNodesStats(t *testing.T) {
	for _, ver := range testElasticsearchVersions {
		filename := fmt.Sprintf("../fixtures/nodestats/%s.json", ver)
		data, _ := os.ReadFile(filename)

		handlers := map[string]http.Handler{
			"plain": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write(data); err != nil {
					t.Fatalf("failed write: %s", err)
				}
			}),
			"basicauth": &basicAuth{
				User: "elastic",
				Pass: "changeme",
				Next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if _, err := w.Write(data); err != nil {
						t.Fatalf("failed write: %s", err)
					}
				}),
			},
		}

		for hn, handler := range handlers {
			t.Run(fmt.Sprintf("%s/%s", hn, ver), func(t *testing.T) {

				ts := httptest.NewServer(handler)
				defer ts.Close()

				u, err := url.Parse(ts.URL)
				if err != nil {
					t.Fatalf("Failed to parse URL: %s", err)
				}
				u.User = url.UserPassword("elastic", "changeme")
				c := NewNodes(log.NewNopLogger(), http.DefaultClient, u, true, "_local")
				nsr, err := c.fetchAndDecodeNodeStats()
				if err != nil {
					t.Fatalf("Failed to fetch or decode node stats: %s", err)
				}
				t.Logf("[%s/%s] Node Stats Response: %+v", hn, ver, nsr)
				// TODO(@sysadmind): Add multinode fixture
				if nsr.ClusterName == "multinode" {
					for _, node := range nsr.Nodes {
						labels := defaultNodeLabelValues(nsr.ClusterName, node)
						esMasterNode := labels[3]
						esDataNode := labels[4]
						esIngestNode := labels[5]
						esClientNode := labels[6]
						t.Logf(
							"Node: %s - Master: %s - Data: %s - Ingest: %s - Client: %s",
							node.Name,
							esMasterNode,
							esDataNode,
							esIngestNode,
							esClientNode,
						)
						if strings.HasPrefix(node.Name, "elasticmaster") {
							if esMasterNode != "true" {
								t.Errorf("Master should be master")
							}
							if esDataNode == "true" {
								t.Errorf("Master should be not data")
							}
							if esIngestNode == "true" {
								t.Errorf("Master should be not ingest")
							}
						}
						if strings.HasPrefix(node.Name, "elasticdata") {
							if esMasterNode == "true" {
								t.Errorf("Data should not be master")
							}
							if esDataNode != "true" {
								t.Errorf("Data should be data")
							}
							if esIngestNode == "true" {
								t.Errorf("Data should be not ingest")
							}
						}
						if strings.HasPrefix(node.Name, "elasticin") {
							if esMasterNode == "true" {
								t.Errorf("Ingest should not be master")
							}
							if esDataNode == "true" {
								t.Errorf("Ingest should be data")
							}
							if esIngestNode != "true" {
								t.Errorf("Ingest should be not ingest")
							}
						}
						if strings.HasPrefix(node.Name, "elasticcli") {
							if esMasterNode == "true" {
								t.Errorf("CLI should not be master")
							}
							if esDataNode == "true" {
								t.Errorf("CLI should be data")
							}
							if esIngestNode == "true" {
								t.Errorf("CLI should be not ingest")
							}
						}
					}
				}
			})
		}
	}
}

type basicAuth struct {
	User string
	Pass string
	Next http.Handler
}

func (h *basicAuth) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	if h.User == pair[0] && h.Pass == pair[1] {
		return true
	}
	return false
}

func (h *basicAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.checkAuth(w, r) {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"ES\"")
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
		return
	}

	h.Next.ServeHTTP(w, r)
}
