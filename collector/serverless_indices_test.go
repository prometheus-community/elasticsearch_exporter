package collector

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestServerlessIndicesUsesCatAPIAndPreservesStoreMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_cat/indices" {
			t.Errorf("path = %q, want /_cat/indices", r.URL.Path)
		}
		if r.URL.Query().Get("format") != "json" || r.URL.Query().Get("bytes") != "b" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = io.WriteString(w, `[{"index":"metrics-prometheus-000001","pri.store.size":"42","store.size":"84"}]`)
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	collector := NewServerlessIndices(slog.New(slog.NewTextHandler(io.Discard, nil)), server.Client(), u)
	want := `
# HELP elasticsearch_indices_store_size_bytes_primary Current total size of stored index data in bytes with only primary shards on all nodes
# TYPE elasticsearch_indices_store_size_bytes_primary gauge
elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index="metrics-prometheus-000001"} 42
# HELP elasticsearch_indices_store_size_bytes_total Current total size of stored index data in bytes with all shards on all nodes
# TYPE elasticsearch_indices_store_size_bytes_total gauge
elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index="metrics-prometheus-000001"} 84
`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(want)); err != nil {
		t.Fatal(err)
	}
}
