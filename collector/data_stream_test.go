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

func TestDataStream(t *testing.T) {
	tests := []struct {
		name           string
		dsStatsFile    string
		indexStatsFile string
		want           string
	}{
		{
			name:           "7.15.0",
			dsStatsFile:    "../fixtures/datastream/7.15.0.json",
			indexStatsFile: "../fixtures/datastream/7.15.0-index-stats.json",
			want: `# HELP elasticsearch_data_stream_backing_indices_total Number of backing indices
            # TYPE elasticsearch_data_stream_backing_indices_total counter
            elasticsearch_data_stream_backing_indices_total{data_stream="bar"} 2
            elasticsearch_data_stream_backing_indices_total{data_stream="foo"} 5
            # HELP elasticsearch_data_stream_store_size_bytes Store size of data stream
            # TYPE elasticsearch_data_stream_store_size_bytes counter
            elasticsearch_data_stream_store_size_bytes{data_stream="bar"} 6.7382272e+08
            elasticsearch_data_stream_store_size_bytes{data_stream="foo"} 4.29205396e+08
            # HELP elasticsearch_data_stream_stats_docs_total Total number of documents in the data stream
            # TYPE elasticsearch_data_stream_stats_docs_total gauge
            elasticsearch_data_stream_stats_docs_total{data_stream="bar"} 1300
            elasticsearch_data_stream_stats_docs_total{data_stream="foo"} 1000
            # HELP elasticsearch_data_stream_stats_indexing_delete_time_seconds_total Total time in seconds spent deleting documents from the data stream
            # TYPE elasticsearch_data_stream_stats_indexing_delete_time_seconds_total counter
            elasticsearch_data_stream_stats_indexing_delete_time_seconds_total{data_stream="bar"} 0.5
            elasticsearch_data_stream_stats_indexing_delete_time_seconds_total{data_stream="foo"} 0.25
            # HELP elasticsearch_data_stream_stats_indexing_delete_total Total number of documents deleted from the data stream
            # TYPE elasticsearch_data_stream_stats_indexing_delete_total counter
            elasticsearch_data_stream_stats_indexing_delete_total{data_stream="bar"} 100
            elasticsearch_data_stream_stats_indexing_delete_total{data_stream="foo"} 50
            # HELP elasticsearch_data_stream_stats_indexing_index_current Number of documents currently being indexed to the data stream
            # TYPE elasticsearch_data_stream_stats_indexing_index_current gauge
            elasticsearch_data_stream_stats_indexing_index_current{data_stream="bar"} 3
            elasticsearch_data_stream_stats_indexing_index_current{data_stream="foo"} 3
            # HELP elasticsearch_data_stream_stats_indexing_index_time_seconds_total Total time in seconds spent indexing documents to the data stream
            # TYPE elasticsearch_data_stream_stats_indexing_index_time_seconds_total counter
            elasticsearch_data_stream_stats_indexing_index_time_seconds_total{data_stream="bar"} 6.5
            elasticsearch_data_stream_stats_indexing_index_time_seconds_total{data_stream="foo"} 5
            # HELP elasticsearch_data_stream_stats_indexing_index_total Total number of documents indexed to the data stream
            # TYPE elasticsearch_data_stream_stats_indexing_index_total counter
            elasticsearch_data_stream_stats_indexing_index_total{data_stream="bar"} 1300
            elasticsearch_data_stream_stats_indexing_index_total{data_stream="foo"} 1000
            # HELP elasticsearch_data_stream_stats_search_fetch_time_seconds_total Total time in seconds spent on search fetch operations on the data stream
            # TYPE elasticsearch_data_stream_stats_search_fetch_time_seconds_total counter
            elasticsearch_data_stream_stats_search_fetch_time_seconds_total{data_stream="bar"} 1.6
            elasticsearch_data_stream_stats_search_fetch_time_seconds_total{data_stream="foo"} 1.2
            # HELP elasticsearch_data_stream_stats_search_fetch_total Total number of search fetch operations on the data stream
            # TYPE elasticsearch_data_stream_stats_search_fetch_total counter
            elasticsearch_data_stream_stats_search_fetch_total{data_stream="bar"} 800
            elasticsearch_data_stream_stats_search_fetch_total{data_stream="foo"} 600
            # HELP elasticsearch_data_stream_stats_search_query_time_seconds_total Total time in seconds spent on search queries on the data stream
            # TYPE elasticsearch_data_stream_stats_search_query_time_seconds_total counter
            elasticsearch_data_stream_stats_search_query_time_seconds_total{data_stream="bar"} 5
            elasticsearch_data_stream_stats_search_query_time_seconds_total{data_stream="foo"} 5
            # HELP elasticsearch_data_stream_stats_search_query_total Total number of search queries executed on the data stream
            # TYPE elasticsearch_data_stream_stats_search_query_total counter
            elasticsearch_data_stream_stats_search_query_total{data_stream="bar"} 1000
            elasticsearch_data_stream_stats_search_query_total{data_stream="foo"} 1000
			`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fDsStats, err := os.Open(tt.dsStatsFile)
			if err != nil {
				t.Fatal(err)
			}
			defer fDsStats.Close()

			fIndexStats, err := os.Open(tt.indexStatsFile)
			if err != nil {
				t.Fatal(err)
			}
			defer fIndexStats.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/_data_stream/*/_stats":
					io.Copy(w, fDsStats)
				case strings.HasSuffix(r.URL.Path, "/_stats"):
					io.Copy(w, fIndexStats)
				default:
					http.Error(w, "Not Found", http.StatusNotFound)
				}
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			c, err := NewDataStream(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
