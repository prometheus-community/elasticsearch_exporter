package collector

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

func TestNodesStats(t *testing.T) {
	for _, ver := range testElasticsearchVersions {
		filename := fmt.Sprintf("../fixtures/indexsettings/%s.json", ver)
		data, _ := ioutil.ReadFile(filename)

		handlers := map[string]http.Handler{
			"plain": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write(data)
			}),
			"basicauth": &basicAuth{
				User: "elastic",
				Pass: "changeme",
				Next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write(data)
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
				c := NewNodes(context.Background(), log.NewNopLogger(), http.DefaultClient, u, true, "_local", time.Nanosecond)
				nsr, err := c.updater.fetchAndDecodeNodeStats()
				if err != nil {
					t.Fatalf("Failed to fetch or decode node stats: %s", err)
				}
				t.Logf("[%s/%s] Node Stats Response: %+v", hn, ver, nsr)
				if nsr.ClusterName == "elasticsearch" {
					for _, nsnr := range nsr.Nodes {
						if nsnr.Indices.Docs.Count > 0 {
							t.Errorf("Wrong doc count")
						}
					}
					u.User = url.UserPassword("elastic", "changeme")
					c := NewNodes(context.Background(), log.NewNopLogger(), http.DefaultClient, u, true, "_local", time.Nanosecond)
					nsr, err := c.updater.fetchAndDecodeNodeStats()
					if err != nil {
						t.Fatalf("Failed to fetch or decode node stats: %s", err)
					}
					t.Logf("[%s/%s] Node Stats Response: %+v", hn, ver, nsr)
					if nsr.ClusterName == "elasticsearch" {
						for _, nsnr := range nsr.Nodes {
							if nsnr.Indices.Docs.Count > 0 {
								t.Errorf("Wrong doc count")
							}
						}
					}
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
