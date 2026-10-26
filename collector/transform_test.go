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

func TestTransform(t *testing.T) {
	// Testcases created using:
	//  curl http://127.0.0.1:9200/_transform/_all/_stats (Numbers manually tweaked)

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "8.10.0",
			file: "8.10.0.json",
			want: `# HELP elasticsearch_transform_stats_checkpoint The sequence number of the last completed checkpoint
            # TYPE elasticsearch_transform_stats_checkpoint gauge
            elasticsearch_transform_stats_checkpoint{id="ecommerce_transform"} 7
            # HELP elasticsearch_transform_stats_delete_time_seconds_total The amount of time spent deleting documents in seconds
            # TYPE elasticsearch_transform_stats_delete_time_seconds_total counter
            elasticsearch_transform_stats_delete_time_seconds_total{id="ecommerce_transform"} 0.05
            # HELP elasticsearch_transform_stats_documents_deleted_total The number of documents deleted from the destination index
            # TYPE elasticsearch_transform_stats_documents_deleted_total counter
            elasticsearch_transform_stats_documents_deleted_total{id="ecommerce_transform"} 5
            # HELP elasticsearch_transform_stats_documents_indexed_total The number of documents indexed into the destination index
            # TYPE elasticsearch_transform_stats_documents_indexed_total counter
            elasticsearch_transform_stats_documents_indexed_total{id="ecommerce_transform"} 1000
            # HELP elasticsearch_transform_stats_documents_processed_total The number of documents processed
            # TYPE elasticsearch_transform_stats_documents_processed_total counter
            elasticsearch_transform_stats_documents_processed_total{id="ecommerce_transform"} 5000
            # HELP elasticsearch_transform_stats_exponential_avg_checkpoint_duration_seconds The exponential moving average of the duration of the checkpoint, in seconds
            # TYPE elasticsearch_transform_stats_exponential_avg_checkpoint_duration_seconds gauge
            elasticsearch_transform_stats_exponential_avg_checkpoint_duration_seconds{id="ecommerce_transform"} 2
            # HELP elasticsearch_transform_stats_exponential_avg_documents_indexed The exponential moving average of the number of new documents that have been indexed
            # TYPE elasticsearch_transform_stats_exponential_avg_documents_indexed gauge
            elasticsearch_transform_stats_exponential_avg_documents_indexed{id="ecommerce_transform"} 100
            # HELP elasticsearch_transform_stats_exponential_avg_documents_processed The exponential moving average of the number of documents that have been processed
            # TYPE elasticsearch_transform_stats_exponential_avg_documents_processed gauge
            elasticsearch_transform_stats_exponential_avg_documents_processed{id="ecommerce_transform"} 500
            # HELP elasticsearch_transform_stats_health_status Health status of the transform, one of: green, yellow, red, unknown
            # TYPE elasticsearch_transform_stats_health_status gauge
            elasticsearch_transform_stats_health_status{id="ecommerce_transform",status="green"} 1
            elasticsearch_transform_stats_health_status{id="ecommerce_transform",status="red"} 0
            elasticsearch_transform_stats_health_status{id="ecommerce_transform",status="unknown"} 0
            elasticsearch_transform_stats_health_status{id="ecommerce_transform",status="yellow"} 0
            # HELP elasticsearch_transform_stats_index_failures_total The number of indexing failures
            # TYPE elasticsearch_transform_stats_index_failures_total counter
            elasticsearch_transform_stats_index_failures_total{id="ecommerce_transform"} 1
            # HELP elasticsearch_transform_stats_index_time_seconds_total The amount of time spent indexing in seconds
            # TYPE elasticsearch_transform_stats_index_time_seconds_total counter
            elasticsearch_transform_stats_index_time_seconds_total{id="ecommerce_transform"} 2.5
            # HELP elasticsearch_transform_stats_index_total The number of index operations
            # TYPE elasticsearch_transform_stats_index_total counter
            elasticsearch_transform_stats_index_total{id="ecommerce_transform"} 100
            # HELP elasticsearch_transform_stats_operations_behind The number of operations in the source index that have not yet been processed
            # TYPE elasticsearch_transform_stats_operations_behind gauge
            elasticsearch_transform_stats_operations_behind{id="ecommerce_transform"} 42
            # HELP elasticsearch_transform_stats_pages_processed_total The number of search or bulk index operations processed
            # TYPE elasticsearch_transform_stats_pages_processed_total counter
            elasticsearch_transform_stats_pages_processed_total{id="ecommerce_transform"} 10
            # HELP elasticsearch_transform_stats_processing_time_seconds_total The amount of time spent processing results in seconds
            # TYPE elasticsearch_transform_stats_processing_time_seconds_total counter
            elasticsearch_transform_stats_processing_time_seconds_total{id="ecommerce_transform"} 0.45
            # HELP elasticsearch_transform_stats_processing_total The number of processing operations
            # TYPE elasticsearch_transform_stats_processing_total counter
            elasticsearch_transform_stats_processing_total{id="ecommerce_transform"} 110
            # HELP elasticsearch_transform_stats_search_failures_total The number of search failures
            # TYPE elasticsearch_transform_stats_search_failures_total counter
            elasticsearch_transform_stats_search_failures_total{id="ecommerce_transform"} 2
            # HELP elasticsearch_transform_stats_search_time_seconds_total The amount of time spent searching in seconds
            # TYPE elasticsearch_transform_stats_search_time_seconds_total counter
            elasticsearch_transform_stats_search_time_seconds_total{id="ecommerce_transform"} 3.5
            # HELP elasticsearch_transform_stats_search_total The number of search operations
            # TYPE elasticsearch_transform_stats_search_total counter
            elasticsearch_transform_stats_search_total{id="ecommerce_transform"} 120
            # HELP elasticsearch_transform_stats_state State of the transform, one of: started, indexing, stopped, stopping, failed, aborting, waiting
            # TYPE elasticsearch_transform_stats_state gauge
            elasticsearch_transform_stats_state{id="ecommerce_transform",state="aborting"} 0
            elasticsearch_transform_stats_state{id="ecommerce_transform",state="failed"} 0
            elasticsearch_transform_stats_state{id="ecommerce_transform",state="indexing"} 0
            elasticsearch_transform_stats_state{id="ecommerce_transform",state="started"} 1
            elasticsearch_transform_stats_state{id="ecommerce_transform",state="stopped"} 0
            elasticsearch_transform_stats_state{id="ecommerce_transform",state="stopping"} 0
            elasticsearch_transform_stats_state{id="ecommerce_transform",state="waiting"} 0
            # HELP elasticsearch_transform_stats_trigger_count_total The number of times the transform has been triggered by the scheduler
            # TYPE elasticsearch_transform_stats_trigger_count_total counter
            elasticsearch_transform_stats_trigger_count_total{id="ecommerce_transform"} 12
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fStatsPath := path.Join("../fixtures/transform/stats/", tt.file)
			fStats, err := os.Open(fStatsPath)
			if err != nil {
				t.Fatal(err)
			}
			defer fStats.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/_transform/_all/_stats":
					io.Copy(w, fStats)
					return
				}

				http.Error(w, "Not Found", http.StatusNotFound)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}

			c, err := NewTransform(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
