package collector

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
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
			if nsr.Cluster.MaxShardsPerNode != "" {
				t.Errorf("MaxShardsPerNode should be empty on older releases")
			}
		}
	}
}

func TestClusterMaxShardsPerNode(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine
	//  curl http://localhost:9200/_cluster/settings/?include_defaults=true
	files := []string{"../fixtures/settings-7.3.0.json"}
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
