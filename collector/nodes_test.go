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
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func TestNodesStats(t *testing.T) {
	tests := []struct {
		name string
		file string
		want string
	}{
		// {
		// 	name: "5.4.2",
		// 	file: "../fixtures/nodestats/5.4.2.json",
		// 	want: ``,
		// },
		{
			name: "5.6.16",
			file: "../fixtures/nodestats/5.6.16.json",
			want: `# HELP elasticsearch_breakers_estimated_size_bytes Estimated size in bytes of breaker
            # TYPE elasticsearch_breakers_estimated_size_bytes gauge
            elasticsearch_breakers_estimated_size_bytes{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            elasticsearch_breakers_estimated_size_bytes{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            elasticsearch_breakers_estimated_size_bytes{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            elasticsearch_breakers_estimated_size_bytes{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_breakers_limit_size_bytes Limit size in bytes for breaker
            # TYPE elasticsearch_breakers_limit_size_bytes gauge
            elasticsearch_breakers_limit_size_bytes{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1.246652006e+09
            elasticsearch_breakers_limit_size_bytes{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 2.077753344e+09
            elasticsearch_breakers_limit_size_bytes{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1.45442734e+09
            elasticsearch_breakers_limit_size_bytes{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1.246652006e+09
            # HELP elasticsearch_breakers_overhead Overhead of circuit breakers
            # TYPE elasticsearch_breakers_overhead counter
            elasticsearch_breakers_overhead{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1.03
            elasticsearch_breakers_overhead{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1
            elasticsearch_breakers_overhead{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1
            elasticsearch_breakers_overhead{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1
            # HELP elasticsearch_breakers_tripped tripped for breaker
            # TYPE elasticsearch_breakers_tripped counter
            elasticsearch_breakers_tripped{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            elasticsearch_breakers_tripped{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            elasticsearch_breakers_tripped{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            elasticsearch_breakers_tripped{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_filesystem_data_available_bytes Available space on block device in bytes
            # TYPE elasticsearch_filesystem_data_available_bytes gauge
            elasticsearch_filesystem_data_available_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",mount="/usr/share/elasticsearch/data (/dev/mapper/vg0-root)",name="bVrN1Hx",path="/usr/share/elasticsearch/data/nodes/0"} 7.7533405184e+10
            # HELP elasticsearch_filesystem_data_free_bytes Free space on block device in bytes
            # TYPE elasticsearch_filesystem_data_free_bytes gauge
            elasticsearch_filesystem_data_free_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",mount="/usr/share/elasticsearch/data (/dev/mapper/vg0-root)",name="bVrN1Hx",path="/usr/share/elasticsearch/data/nodes/0"} 7.7533405184e+10
            # HELP elasticsearch_filesystem_data_size_bytes Size of block device in bytes
            # TYPE elasticsearch_filesystem_data_size_bytes gauge
            elasticsearch_filesystem_data_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",mount="/usr/share/elasticsearch/data (/dev/mapper/vg0-root)",name="bVrN1Hx",path="/usr/share/elasticsearch/data/nodes/0"} 4.76630163456e+11
            # HELP elasticsearch_filesystem_io_stats_device_operations_count Count of disk operations
            # TYPE elasticsearch_filesystem_io_stats_device_operations_count counter
            elasticsearch_filesystem_io_stats_device_operations_count{cluster="elasticsearch",device="dm-2",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 2517
            # HELP elasticsearch_filesystem_io_stats_device_read_operations_count Count of disk read operations
            # TYPE elasticsearch_filesystem_io_stats_device_read_operations_count counter
            elasticsearch_filesystem_io_stats_device_read_operations_count{cluster="elasticsearch",device="dm-2",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 706
            # HELP elasticsearch_filesystem_io_stats_device_read_size_kilobytes_sum Total kilobytes read from disk
            # TYPE elasticsearch_filesystem_io_stats_device_read_size_kilobytes_sum counter
            elasticsearch_filesystem_io_stats_device_read_size_kilobytes_sum{cluster="elasticsearch",device="dm-2",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 12916
            # HELP elasticsearch_filesystem_io_stats_device_write_operations_count Count of disk write operations
            # TYPE elasticsearch_filesystem_io_stats_device_write_operations_count counter
            elasticsearch_filesystem_io_stats_device_write_operations_count{cluster="elasticsearch",device="dm-2",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1811
            # HELP elasticsearch_filesystem_io_stats_device_write_size_kilobytes_sum Total kilobytes written to disk
            # TYPE elasticsearch_filesystem_io_stats_device_write_size_kilobytes_sum counter
            elasticsearch_filesystem_io_stats_device_write_size_kilobytes_sum{cluster="elasticsearch",device="dm-2",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 17760
            # HELP elasticsearch_indices_completion_size_in_bytes Completion in bytes
            # TYPE elasticsearch_indices_completion_size_in_bytes counter
            elasticsearch_indices_completion_size_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_docs Count of documents on this node
            # TYPE elasticsearch_indices_docs gauge
            elasticsearch_indices_docs{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 5
            # HELP elasticsearch_indices_docs_deleted Count of deleted documents on this node
            # TYPE elasticsearch_indices_docs_deleted gauge
            elasticsearch_indices_docs_deleted{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_fielddata_evictions Evictions from field data
            # TYPE elasticsearch_indices_fielddata_evictions counter
            elasticsearch_indices_fielddata_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_fielddata_memory_size_bytes Field data cache memory usage in bytes
            # TYPE elasticsearch_indices_fielddata_memory_size_bytes gauge
            elasticsearch_indices_fielddata_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_filter_cache_evictions Evictions from filter cache
            # TYPE elasticsearch_indices_filter_cache_evictions counter
            elasticsearch_indices_filter_cache_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_filter_cache_memory_size_bytes Filter cache memory usage in bytes
            # TYPE elasticsearch_indices_filter_cache_memory_size_bytes gauge
            elasticsearch_indices_filter_cache_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_flush_time_seconds Cumulative flush time in seconds
            # TYPE elasticsearch_indices_flush_time_seconds counter
            elasticsearch_indices_flush_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_flush_total Total flushes
            # TYPE elasticsearch_indices_flush_total counter
            elasticsearch_indices_flush_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_get_exists_time_seconds Total time get exists in seconds
            # TYPE elasticsearch_indices_get_exists_time_seconds counter
            elasticsearch_indices_get_exists_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_get_exists_total Total get exists operations
            # TYPE elasticsearch_indices_get_exists_total counter
            elasticsearch_indices_get_exists_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_get_missing_time_seconds Total time of get missing in seconds
            # TYPE elasticsearch_indices_get_missing_time_seconds counter
            elasticsearch_indices_get_missing_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_get_missing_total Total get missing
            # TYPE elasticsearch_indices_get_missing_total counter
            elasticsearch_indices_get_missing_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_get_time_seconds Total get time in seconds
            # TYPE elasticsearch_indices_get_time_seconds counter
            elasticsearch_indices_get_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_get_total Total get
            # TYPE elasticsearch_indices_get_total counter
            elasticsearch_indices_get_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_indexing_delete_time_seconds_total Total time indexing delete in seconds
            # TYPE elasticsearch_indices_indexing_delete_time_seconds_total counter
            elasticsearch_indices_indexing_delete_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_indexing_delete_total Total indexing deletes
            # TYPE elasticsearch_indices_indexing_delete_total counter
            elasticsearch_indices_indexing_delete_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_indexing_index_time_seconds_total Cumulative index time in seconds
            # TYPE elasticsearch_indices_indexing_index_time_seconds_total counter
            elasticsearch_indices_indexing_index_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0.039
            # HELP elasticsearch_indices_indexing_index_total Total index calls
            # TYPE elasticsearch_indices_indexing_index_total counter
            elasticsearch_indices_indexing_index_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 5
            # HELP elasticsearch_indices_indexing_is_throttled Indexing throttling
            # TYPE elasticsearch_indices_indexing_is_throttled gauge
            elasticsearch_indices_indexing_is_throttled{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_indexing_throttle_time_seconds_total Cumulative indexing throttling time
            # TYPE elasticsearch_indices_indexing_throttle_time_seconds_total counter
            elasticsearch_indices_indexing_throttle_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_merges_current Current merges
            # TYPE elasticsearch_indices_merges_current gauge
            elasticsearch_indices_merges_current{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_merges_current_size_in_bytes Size of a current merges in bytes
            # TYPE elasticsearch_indices_merges_current_size_in_bytes gauge
            elasticsearch_indices_merges_current_size_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_merges_docs_total Cumulative docs merged
            # TYPE elasticsearch_indices_merges_docs_total counter
            elasticsearch_indices_merges_docs_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_merges_total Total merges
            # TYPE elasticsearch_indices_merges_total counter
            elasticsearch_indices_merges_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_merges_total_size_bytes_total Total merge size in bytes
            # TYPE elasticsearch_indices_merges_total_size_bytes_total counter
            elasticsearch_indices_merges_total_size_bytes_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_merges_total_throttled_time_seconds_total Total throttled time of merges in seconds
            # TYPE elasticsearch_indices_merges_total_throttled_time_seconds_total counter
            elasticsearch_indices_merges_total_throttled_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_merges_total_time_seconds_total Total time spent merging in seconds
            # TYPE elasticsearch_indices_merges_total_time_seconds_total counter
            elasticsearch_indices_merges_total_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_query_cache_cache_size Query cache cache size
            # TYPE elasticsearch_indices_query_cache_cache_size gauge
            elasticsearch_indices_query_cache_cache_size{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_query_cache_cache_total Query cache cache count
            # TYPE elasticsearch_indices_query_cache_cache_total counter
            elasticsearch_indices_query_cache_cache_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_query_cache_count Query cache count
            # TYPE elasticsearch_indices_query_cache_count counter
            elasticsearch_indices_query_cache_count{cache="hit",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_query_cache_evictions Evictions from query cache
            # TYPE elasticsearch_indices_query_cache_evictions counter
            elasticsearch_indices_query_cache_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_query_cache_memory_size_bytes Query cache memory usage in bytes
            # TYPE elasticsearch_indices_query_cache_memory_size_bytes gauge
            elasticsearch_indices_query_cache_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_query_cache_total Query cache total count
            # TYPE elasticsearch_indices_query_cache_total counter
            elasticsearch_indices_query_cache_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_query_miss_count Query miss count
            # TYPE elasticsearch_indices_query_miss_count counter
            elasticsearch_indices_query_miss_count{cache="miss",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_refresh_external_time_seconds_total Total time spent external refreshing in seconds
            # TYPE elasticsearch_indices_refresh_external_time_seconds_total counter
            elasticsearch_indices_refresh_external_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_refresh_external_total Total external refreshes
            # TYPE elasticsearch_indices_refresh_external_total counter
            elasticsearch_indices_refresh_external_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_refresh_time_seconds_total Total time spent refreshing in seconds
            # TYPE elasticsearch_indices_refresh_time_seconds_total counter
            elasticsearch_indices_refresh_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0.086
            # HELP elasticsearch_indices_refresh_total Total refreshes
            # TYPE elasticsearch_indices_refresh_total counter
            elasticsearch_indices_refresh_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 65
            # HELP elasticsearch_indices_request_cache_count Request cache count
            # TYPE elasticsearch_indices_request_cache_count counter
            elasticsearch_indices_request_cache_count{cache="hit",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_request_cache_evictions Evictions from request cache
            # TYPE elasticsearch_indices_request_cache_evictions counter
            elasticsearch_indices_request_cache_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_request_cache_memory_size_bytes Request cache memory usage in bytes
            # TYPE elasticsearch_indices_request_cache_memory_size_bytes gauge
            elasticsearch_indices_request_cache_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_request_miss_count Request miss count
            # TYPE elasticsearch_indices_request_miss_count counter
            elasticsearch_indices_request_miss_count{cache="miss",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_search_fetch_time_seconds Total search fetch time in seconds
            # TYPE elasticsearch_indices_search_fetch_time_seconds counter
            elasticsearch_indices_search_fetch_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_search_fetch_total Total number of fetches
            # TYPE elasticsearch_indices_search_fetch_total counter
            elasticsearch_indices_search_fetch_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_search_query_time_seconds Total search query time in seconds
            # TYPE elasticsearch_indices_search_query_time_seconds counter
            elasticsearch_indices_search_query_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_search_query_total Total number of queries
            # TYPE elasticsearch_indices_search_query_total counter
            elasticsearch_indices_search_query_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_search_scroll_time_seconds Total scroll time in seconds
            # TYPE elasticsearch_indices_search_scroll_time_seconds counter
            elasticsearch_indices_search_scroll_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_search_scroll_total Total number of scrolls
            # TYPE elasticsearch_indices_search_scroll_total counter
            elasticsearch_indices_search_scroll_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_search_suggest_time_seconds Total suggest time in seconds
            # TYPE elasticsearch_indices_search_suggest_time_seconds counter
            elasticsearch_indices_search_suggest_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_search_suggest_total Total number of suggests
            # TYPE elasticsearch_indices_search_suggest_total counter
            elasticsearch_indices_search_suggest_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_segments_count Count of index segments on this node
            # TYPE elasticsearch_indices_segments_count gauge
            elasticsearch_indices_segments_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 5
            # HELP elasticsearch_indices_segments_doc_values_memory_in_bytes Count of doc values memory
            # TYPE elasticsearch_indices_segments_doc_values_memory_in_bytes gauge
            elasticsearch_indices_segments_doc_values_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 460
            # HELP elasticsearch_indices_segments_fixed_bit_set_memory_in_bytes Count of fixed bit set
            # TYPE elasticsearch_indices_segments_fixed_bit_set_memory_in_bytes gauge
            elasticsearch_indices_segments_fixed_bit_set_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_segments_index_writer_memory_in_bytes Count of memory for index writer on this node
            # TYPE elasticsearch_indices_segments_index_writer_memory_in_bytes gauge
            elasticsearch_indices_segments_index_writer_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_segments_memory_bytes Current memory size of segments in bytes
            # TYPE elasticsearch_indices_segments_memory_bytes gauge
            elasticsearch_indices_segments_memory_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 12940
            # HELP elasticsearch_indices_segments_norms_memory_in_bytes Count of memory used by norms
            # TYPE elasticsearch_indices_segments_norms_memory_in_bytes gauge
            elasticsearch_indices_segments_norms_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 960
            # HELP elasticsearch_indices_segments_points_memory_in_bytes Point values memory usage in bytes
            # TYPE elasticsearch_indices_segments_points_memory_in_bytes gauge
            elasticsearch_indices_segments_points_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_segments_stored_fields_memory_in_bytes Count of stored fields memory
            # TYPE elasticsearch_indices_segments_stored_fields_memory_in_bytes gauge
            elasticsearch_indices_segments_stored_fields_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1560
            # HELP elasticsearch_indices_segments_term_vectors_memory_in_bytes Term vectors memory usage in bytes
            # TYPE elasticsearch_indices_segments_term_vectors_memory_in_bytes gauge
            elasticsearch_indices_segments_term_vectors_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_segments_terms_memory_in_bytes Count of terms in memory for this node
            # TYPE elasticsearch_indices_segments_terms_memory_in_bytes gauge
            elasticsearch_indices_segments_terms_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 9960
            # HELP elasticsearch_indices_segments_version_map_memory_in_bytes Version map memory usage in bytes
            # TYPE elasticsearch_indices_segments_version_map_memory_in_bytes gauge
            elasticsearch_indices_segments_version_map_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_store_size_bytes Current size of stored index data in bytes
            # TYPE elasticsearch_indices_store_size_bytes gauge
            elasticsearch_indices_store_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 5721
            # HELP elasticsearch_indices_store_throttle_time_seconds_total Throttle time for index store in seconds
            # TYPE elasticsearch_indices_store_throttle_time_seconds_total counter
            elasticsearch_indices_store_throttle_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_indices_translog_operations Total translog operations
            # TYPE elasticsearch_indices_translog_operations counter
            elasticsearch_indices_translog_operations{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 5
            # HELP elasticsearch_indices_translog_size_in_bytes Total translog size in bytes
            # TYPE elasticsearch_indices_translog_size_in_bytes gauge
            elasticsearch_indices_translog_size_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1743
            # HELP elasticsearch_indices_warmer_time_seconds_total Total warmer time in seconds
            # TYPE elasticsearch_indices_warmer_time_seconds_total counter
            elasticsearch_indices_warmer_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0.001
            # HELP elasticsearch_indices_warmer_total Total warmer count
            # TYPE elasticsearch_indices_warmer_total counter
            elasticsearch_indices_warmer_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 35
            # HELP elasticsearch_jvm_buffer_pool_used_bytes JVM buffer currently used
            # TYPE elasticsearch_jvm_buffer_pool_used_bytes gauge
            elasticsearch_jvm_buffer_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="direct"} 2.52727869e+08
            elasticsearch_jvm_buffer_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="mapped"} 15007
            # HELP elasticsearch_jvm_gc_collection_seconds_count Count of JVM GC runs
            # TYPE elasticsearch_jvm_gc_collection_seconds_count counter
            elasticsearch_jvm_gc_collection_seconds_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="old",host="127.0.0.1",name="bVrN1Hx"} 1
            elasticsearch_jvm_gc_collection_seconds_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="young",host="127.0.0.1",name="bVrN1Hx"} 2
            # HELP elasticsearch_jvm_gc_collection_seconds_sum GC run time in seconds
            # TYPE elasticsearch_jvm_gc_collection_seconds_sum counter
            elasticsearch_jvm_gc_collection_seconds_sum{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="old",host="127.0.0.1",name="bVrN1Hx"} 0.109
            elasticsearch_jvm_gc_collection_seconds_sum{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="young",host="127.0.0.1",name="bVrN1Hx"} 0.143
            # HELP elasticsearch_jvm_memory_committed_bytes JVM memory currently committed by area
            # TYPE elasticsearch_jvm_memory_committed_bytes gauge
            elasticsearch_jvm_memory_committed_bytes{area="heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 2.077753344e+09
            elasticsearch_jvm_memory_committed_bytes{area="non-heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 7.5362304e+07
            # HELP elasticsearch_jvm_memory_max_bytes JVM memory max
            # TYPE elasticsearch_jvm_memory_max_bytes gauge
            elasticsearch_jvm_memory_max_bytes{area="heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 2.077753344e+09
            # HELP elasticsearch_jvm_memory_pool_max_bytes JVM memory max by pool
            # TYPE elasticsearch_jvm_memory_pool_max_bytes counter
            elasticsearch_jvm_memory_pool_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="old"} 1.449590784e+09
            elasticsearch_jvm_memory_pool_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="survivor"} 6.9730304e+07
            elasticsearch_jvm_memory_pool_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="young"} 5.58432256e+08
            # HELP elasticsearch_jvm_memory_pool_peak_max_bytes JVM memory peak max by pool
            # TYPE elasticsearch_jvm_memory_pool_peak_max_bytes counter
            elasticsearch_jvm_memory_pool_peak_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="old"} 1.449590784e+09
            elasticsearch_jvm_memory_pool_peak_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="survivor"} 6.9730304e+07
            elasticsearch_jvm_memory_pool_peak_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="young"} 5.58432256e+08
            # HELP elasticsearch_jvm_memory_pool_peak_used_bytes JVM memory peak used by pool
            # TYPE elasticsearch_jvm_memory_pool_peak_used_bytes counter
            elasticsearch_jvm_memory_pool_peak_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="old"} 2.10051288e+08
            elasticsearch_jvm_memory_pool_peak_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="survivor"} 6.9730304e+07
            elasticsearch_jvm_memory_pool_peak_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="young"} 5.58432256e+08
            # HELP elasticsearch_jvm_memory_pool_used_bytes JVM memory currently used by pool
            # TYPE elasticsearch_jvm_memory_pool_used_bytes gauge
            elasticsearch_jvm_memory_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="old"} 2.10051288e+08
            elasticsearch_jvm_memory_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="survivor"} 6.9730304e+07
            elasticsearch_jvm_memory_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",pool="young"} 5.3925336e+07
            # HELP elasticsearch_jvm_memory_used_bytes JVM memory currently used by area
            # TYPE elasticsearch_jvm_memory_used_bytes gauge
            elasticsearch_jvm_memory_used_bytes{area="heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 3.33706928e+08
            elasticsearch_jvm_memory_used_bytes{area="non-heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 7.0212664e+07
            # HELP elasticsearch_jvm_uptime_seconds JVM process uptime in seconds
            # TYPE elasticsearch_jvm_uptime_seconds gauge
            elasticsearch_jvm_uptime_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="mapped"} 14.845
            # HELP elasticsearch_nodes_roles Node roles
            # TYPE elasticsearch_nodes_roles gauge
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="client"} 1
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="data"} 1
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="data_cold"} 0
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="data_content"} 0
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="data_frozen"} 0
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="data_hot"} 0
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="data_warm"} 0
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="ingest"} 1
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="master"} 1
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="ml"} 0
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="remote_cluster_client"} 0
            elasticsearch_nodes_roles{cluster="elasticsearch",host="127.0.0.1",name="bVrN1Hx",role="transform"} 0
            # HELP elasticsearch_os_cpu_percent Percent CPU used by OS
            # TYPE elasticsearch_os_cpu_percent gauge
            elasticsearch_os_cpu_percent{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 23
            # HELP elasticsearch_os_load1 Shortterm load average
            # TYPE elasticsearch_os_load1 gauge
            elasticsearch_os_load1{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 45.68
            # HELP elasticsearch_os_load15 Longterm load average
            # TYPE elasticsearch_os_load15 gauge
            elasticsearch_os_load15{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 96.01
            # HELP elasticsearch_os_load5 Midterm load average
            # TYPE elasticsearch_os_load5 gauge
            elasticsearch_os_load5{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 150.34
            # HELP elasticsearch_os_mem_actual_free_bytes Amount of free physical memory in bytes
            # TYPE elasticsearch_os_mem_actual_free_bytes gauge
            elasticsearch_os_mem_actual_free_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_os_mem_actual_used_bytes Amount of used physical memory in bytes
            # TYPE elasticsearch_os_mem_actual_used_bytes gauge
            elasticsearch_os_mem_actual_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_os_mem_free_bytes Amount of free physical memory in bytes
            # TYPE elasticsearch_os_mem_free_bytes gauge
            elasticsearch_os_mem_free_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 7.009173504e+09
            # HELP elasticsearch_os_mem_used_bytes Amount of used physical memory in bytes
            # TYPE elasticsearch_os_mem_used_bytes gauge
            elasticsearch_os_mem_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 2.6614063104e+10
            # HELP elasticsearch_process_cpu_percent Percent CPU used by process
            # TYPE elasticsearch_process_cpu_percent gauge
            elasticsearch_process_cpu_percent{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 8
            # HELP elasticsearch_process_cpu_seconds_total Process CPU time in seconds
            # TYPE elasticsearch_process_cpu_seconds_total counter
            elasticsearch_process_cpu_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 14.51
            # HELP elasticsearch_process_max_files_descriptors Max file descriptors
            # TYPE elasticsearch_process_max_files_descriptors gauge
            elasticsearch_process_max_files_descriptors{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 1.048576e+06
            # HELP elasticsearch_process_mem_resident_size_bytes Resident memory in use by process in bytes
            # TYPE elasticsearch_process_mem_resident_size_bytes gauge
            elasticsearch_process_mem_resident_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_process_mem_share_size_bytes Shared memory in use by process in bytes
            # TYPE elasticsearch_process_mem_share_size_bytes gauge
            elasticsearch_process_mem_share_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_process_mem_virtual_size_bytes Total virtual memory used in bytes
            # TYPE elasticsearch_process_mem_virtual_size_bytes gauge
            elasticsearch_process_mem_virtual_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 8.293711872e+09
            # HELP elasticsearch_process_open_files_count Open file descriptors
            # TYPE elasticsearch_process_open_files_count gauge
            elasticsearch_process_open_files_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 308
            # HELP elasticsearch_thread_pool_active_count Thread Pool threads active
            # TYPE elasticsearch_thread_pool_active_count gauge
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="bulk"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_started"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_store"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="flush"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="force_merge"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="generic"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="get"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="index"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="listener"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="management"} 1
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="refresh"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="search"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="snapshot"} 0
            elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="warmer"} 0
            # HELP elasticsearch_thread_pool_completed_count Thread Pool operations completed
            # TYPE elasticsearch_thread_pool_completed_count counter
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="bulk"} 5
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_started"} 0
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_store"} 0
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="flush"} 0
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="force_merge"} 0
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="generic"} 38
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="get"} 0
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="index"} 0
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="listener"} 5
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="management"} 2
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="refresh"} 31
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="search"} 0
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="snapshot"} 0
            elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="warmer"} 0
            # HELP elasticsearch_thread_pool_largest_count Thread Pool largest threads count
            # TYPE elasticsearch_thread_pool_largest_count gauge
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="bulk"} 5
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_started"} 0
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_store"} 0
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="flush"} 0
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="force_merge"} 0
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="generic"} 4
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="get"} 0
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="index"} 0
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="listener"} 4
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="management"} 1
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="refresh"} 1
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="search"} 0
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="snapshot"} 0
            elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="warmer"} 0
            # HELP elasticsearch_thread_pool_queue_count Thread Pool operations queued
            # TYPE elasticsearch_thread_pool_queue_count gauge
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="bulk"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_started"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_store"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="flush"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="force_merge"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="generic"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="get"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="index"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="listener"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="management"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="refresh"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="search"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="snapshot"} 0
            elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="warmer"} 0
            # HELP elasticsearch_thread_pool_rejected_count Thread Pool operations rejected
            # TYPE elasticsearch_thread_pool_rejected_count counter
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="bulk"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_started"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_store"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="flush"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="force_merge"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="generic"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="get"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="index"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="listener"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="management"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="refresh"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="search"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="snapshot"} 0
            elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="warmer"} 0
            # HELP elasticsearch_thread_pool_threads_count Thread Pool current threads count
            # TYPE elasticsearch_thread_pool_threads_count gauge
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="bulk"} 5
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_started"} 0
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="fetch_shard_store"} 0
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="flush"} 0
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="force_merge"} 0
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="generic"} 4
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="get"} 0
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="index"} 0
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="listener"} 4
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="management"} 1
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="refresh"} 1
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="search"} 0
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="snapshot"} 0
            elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx",type="warmer"} 0
            # HELP elasticsearch_transport_rx_packets_total Count of packets received
            # TYPE elasticsearch_transport_rx_packets_total counter
            elasticsearch_transport_rx_packets_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_transport_rx_size_bytes_total Total number of bytes received
            # TYPE elasticsearch_transport_rx_size_bytes_total counter
            elasticsearch_transport_rx_size_bytes_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_transport_tx_packets_total Count of packets sent
            # TYPE elasticsearch_transport_tx_packets_total counter
            elasticsearch_transport_tx_packets_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
            # HELP elasticsearch_transport_tx_size_bytes_total Total number of bytes sent
            # TYPE elasticsearch_transport_tx_size_bytes_total counter
            elasticsearch_transport_tx_size_bytes_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="127.0.0.1",name="bVrN1Hx"} 0
`,
		},
		{
			name: "6.8.8",
			file: "../fixtures/nodestats/6.8.8.json",
			want: `# HELP elasticsearch_breakers_estimated_size_bytes Estimated size in bytes of breaker
						 # TYPE elasticsearch_breakers_estimated_size_bytes gauge
             elasticsearch_breakers_estimated_size_bytes{breaker="accounting",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 8490
             elasticsearch_breakers_estimated_size_bytes{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             elasticsearch_breakers_estimated_size_bytes{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             elasticsearch_breakers_estimated_size_bytes{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 8490
             elasticsearch_breakers_estimated_size_bytes{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_breakers_limit_size_bytes Limit size in bytes for breaker
             # TYPE elasticsearch_breakers_limit_size_bytes gauge
             elasticsearch_breakers_limit_size_bytes{breaker="accounting",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1.073741824e+09
             elasticsearch_breakers_limit_size_bytes{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 6.44245094e+08
             elasticsearch_breakers_limit_size_bytes{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1.073741824e+09
             elasticsearch_breakers_limit_size_bytes{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 7.51619276e+08
             elasticsearch_breakers_limit_size_bytes{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 6.44245094e+08
             # HELP elasticsearch_breakers_overhead Overhead of circuit breakers
             # TYPE elasticsearch_breakers_overhead counter
             elasticsearch_breakers_overhead{breaker="accounting",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1
             elasticsearch_breakers_overhead{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1.03
             elasticsearch_breakers_overhead{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1
             elasticsearch_breakers_overhead{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1
             elasticsearch_breakers_overhead{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1
             # HELP elasticsearch_breakers_tripped tripped for breaker
             # TYPE elasticsearch_breakers_tripped counter
             elasticsearch_breakers_tripped{breaker="accounting",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             elasticsearch_breakers_tripped{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             elasticsearch_breakers_tripped{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             elasticsearch_breakers_tripped{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             elasticsearch_breakers_tripped{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_filesystem_data_available_bytes Available space on block device in bytes
             # TYPE elasticsearch_filesystem_data_available_bytes gauge
             elasticsearch_filesystem_data_available_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",mount="/ (overlay)",name="9_P7yui",path="/usr/share/elasticsearch/data/nodes/0"} 7.753281536e+10
             # HELP elasticsearch_filesystem_data_free_bytes Free space on block device in bytes
             # TYPE elasticsearch_filesystem_data_free_bytes gauge
             elasticsearch_filesystem_data_free_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",mount="/ (overlay)",name="9_P7yui",path="/usr/share/elasticsearch/data/nodes/0"} 7.753281536e+10
             # HELP elasticsearch_filesystem_data_size_bytes Size of block device in bytes
             # TYPE elasticsearch_filesystem_data_size_bytes gauge
             elasticsearch_filesystem_data_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",mount="/ (overlay)",name="9_P7yui",path="/usr/share/elasticsearch/data/nodes/0"} 4.76630163456e+11
             # HELP elasticsearch_indices_completion_size_in_bytes Completion in bytes
             # TYPE elasticsearch_indices_completion_size_in_bytes counter
             elasticsearch_indices_completion_size_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_docs Count of documents on this node
             # TYPE elasticsearch_indices_docs gauge
             elasticsearch_indices_docs{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 5
             # HELP elasticsearch_indices_docs_deleted Count of deleted documents on this node
             # TYPE elasticsearch_indices_docs_deleted gauge
             elasticsearch_indices_docs_deleted{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_fielddata_evictions Evictions from field data
             # TYPE elasticsearch_indices_fielddata_evictions counter
             elasticsearch_indices_fielddata_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_fielddata_memory_size_bytes Field data cache memory usage in bytes
             # TYPE elasticsearch_indices_fielddata_memory_size_bytes gauge
             elasticsearch_indices_fielddata_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_filter_cache_evictions Evictions from filter cache
             # TYPE elasticsearch_indices_filter_cache_evictions counter
             elasticsearch_indices_filter_cache_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_filter_cache_memory_size_bytes Filter cache memory usage in bytes
             # TYPE elasticsearch_indices_filter_cache_memory_size_bytes gauge
             elasticsearch_indices_filter_cache_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_flush_time_seconds Cumulative flush time in seconds
             # TYPE elasticsearch_indices_flush_time_seconds counter
             elasticsearch_indices_flush_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_flush_total Total flushes
             # TYPE elasticsearch_indices_flush_total counter
             elasticsearch_indices_flush_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_get_exists_time_seconds Total time get exists in seconds
             # TYPE elasticsearch_indices_get_exists_time_seconds counter
             elasticsearch_indices_get_exists_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_get_exists_total Total get exists operations
             # TYPE elasticsearch_indices_get_exists_total counter
             elasticsearch_indices_get_exists_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_get_missing_time_seconds Total time of get missing in seconds
             # TYPE elasticsearch_indices_get_missing_time_seconds counter
             elasticsearch_indices_get_missing_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_get_missing_total Total get missing
             # TYPE elasticsearch_indices_get_missing_total counter
             elasticsearch_indices_get_missing_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_get_time_seconds Total get time in seconds
             # TYPE elasticsearch_indices_get_time_seconds counter
             elasticsearch_indices_get_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_get_total Total get
             # TYPE elasticsearch_indices_get_total counter
             elasticsearch_indices_get_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_indexing_delete_time_seconds_total Total time indexing delete in seconds
             # TYPE elasticsearch_indices_indexing_delete_time_seconds_total counter
             elasticsearch_indices_indexing_delete_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_indexing_delete_total Total indexing deletes
             # TYPE elasticsearch_indices_indexing_delete_total counter
             elasticsearch_indices_indexing_delete_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_indexing_index_time_seconds_total Cumulative index time in seconds
             # TYPE elasticsearch_indices_indexing_index_time_seconds_total counter
             elasticsearch_indices_indexing_index_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0.038
             # HELP elasticsearch_indices_indexing_index_total Total index calls
             # TYPE elasticsearch_indices_indexing_index_total counter
             elasticsearch_indices_indexing_index_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 5
             # HELP elasticsearch_indices_indexing_is_throttled Indexing throttling
             # TYPE elasticsearch_indices_indexing_is_throttled gauge
             elasticsearch_indices_indexing_is_throttled{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_indexing_throttle_time_seconds_total Cumulative indexing throttling time
             # TYPE elasticsearch_indices_indexing_throttle_time_seconds_total counter
             elasticsearch_indices_indexing_throttle_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_merges_current Current merges
             # TYPE elasticsearch_indices_merges_current gauge
             elasticsearch_indices_merges_current{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_merges_current_size_in_bytes Size of a current merges in bytes
             # TYPE elasticsearch_indices_merges_current_size_in_bytes gauge
             elasticsearch_indices_merges_current_size_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_merges_docs_total Cumulative docs merged
             # TYPE elasticsearch_indices_merges_docs_total counter
             elasticsearch_indices_merges_docs_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_merges_total Total merges
             # TYPE elasticsearch_indices_merges_total counter
             elasticsearch_indices_merges_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_merges_total_size_bytes_total Total merge size in bytes
             # TYPE elasticsearch_indices_merges_total_size_bytes_total counter
             elasticsearch_indices_merges_total_size_bytes_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_merges_total_throttled_time_seconds_total Total throttled time of merges in seconds
             # TYPE elasticsearch_indices_merges_total_throttled_time_seconds_total counter
             elasticsearch_indices_merges_total_throttled_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_merges_total_time_seconds_total Total time spent merging in seconds
             # TYPE elasticsearch_indices_merges_total_time_seconds_total counter
             elasticsearch_indices_merges_total_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_query_cache_cache_size Query cache cache size
             # TYPE elasticsearch_indices_query_cache_cache_size gauge
             elasticsearch_indices_query_cache_cache_size{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_query_cache_cache_total Query cache cache count
             # TYPE elasticsearch_indices_query_cache_cache_total counter
             elasticsearch_indices_query_cache_cache_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_query_cache_count Query cache count
             # TYPE elasticsearch_indices_query_cache_count counter
             elasticsearch_indices_query_cache_count{cache="hit",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_query_cache_evictions Evictions from query cache
             # TYPE elasticsearch_indices_query_cache_evictions counter
             elasticsearch_indices_query_cache_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_query_cache_memory_size_bytes Query cache memory usage in bytes
             # TYPE elasticsearch_indices_query_cache_memory_size_bytes gauge
             elasticsearch_indices_query_cache_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_query_cache_total Query cache total count
             # TYPE elasticsearch_indices_query_cache_total counter
             elasticsearch_indices_query_cache_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_query_miss_count Query miss count
             # TYPE elasticsearch_indices_query_miss_count counter
             elasticsearch_indices_query_miss_count{cache="miss",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_refresh_external_time_seconds_total Total time spent external refreshing in seconds
             # TYPE elasticsearch_indices_refresh_external_time_seconds_total counter
             elasticsearch_indices_refresh_external_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_refresh_external_total Total external refreshes
             # TYPE elasticsearch_indices_refresh_external_total counter
             elasticsearch_indices_refresh_external_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_refresh_time_seconds_total Total time spent refreshing in seconds
             # TYPE elasticsearch_indices_refresh_time_seconds_total counter
             elasticsearch_indices_refresh_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0.103
             # HELP elasticsearch_indices_refresh_total Total refreshes
             # TYPE elasticsearch_indices_refresh_total counter
             elasticsearch_indices_refresh_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 65
             # HELP elasticsearch_indices_request_cache_count Request cache count
             # TYPE elasticsearch_indices_request_cache_count counter
             elasticsearch_indices_request_cache_count{cache="hit",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_request_cache_evictions Evictions from request cache
             # TYPE elasticsearch_indices_request_cache_evictions counter
             elasticsearch_indices_request_cache_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_request_cache_memory_size_bytes Request cache memory usage in bytes
             # TYPE elasticsearch_indices_request_cache_memory_size_bytes gauge
             elasticsearch_indices_request_cache_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_request_miss_count Request miss count
             # TYPE elasticsearch_indices_request_miss_count counter
             elasticsearch_indices_request_miss_count{cache="miss",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_search_fetch_time_seconds Total search fetch time in seconds
             # TYPE elasticsearch_indices_search_fetch_time_seconds counter
             elasticsearch_indices_search_fetch_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_search_fetch_total Total number of fetches
             # TYPE elasticsearch_indices_search_fetch_total counter
             elasticsearch_indices_search_fetch_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_search_query_time_seconds Total search query time in seconds
             # TYPE elasticsearch_indices_search_query_time_seconds counter
             elasticsearch_indices_search_query_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_search_query_total Total number of queries
             # TYPE elasticsearch_indices_search_query_total counter
             elasticsearch_indices_search_query_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_search_scroll_time_seconds Total scroll time in seconds
             # TYPE elasticsearch_indices_search_scroll_time_seconds counter
             elasticsearch_indices_search_scroll_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_search_scroll_total Total number of scrolls
             # TYPE elasticsearch_indices_search_scroll_total counter
             elasticsearch_indices_search_scroll_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_search_suggest_time_seconds Total suggest time in seconds
             # TYPE elasticsearch_indices_search_suggest_time_seconds counter
             elasticsearch_indices_search_suggest_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_search_suggest_total Total number of suggests
             # TYPE elasticsearch_indices_search_suggest_total counter
             elasticsearch_indices_search_suggest_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_segments_count Count of index segments on this node
             # TYPE elasticsearch_indices_segments_count gauge
             elasticsearch_indices_segments_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 5
             # HELP elasticsearch_indices_segments_doc_values_memory_in_bytes Count of doc values memory
             # TYPE elasticsearch_indices_segments_doc_values_memory_in_bytes gauge
             elasticsearch_indices_segments_doc_values_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 340
             # HELP elasticsearch_indices_segments_fixed_bit_set_memory_in_bytes Count of fixed bit set
             # TYPE elasticsearch_indices_segments_fixed_bit_set_memory_in_bytes gauge
             elasticsearch_indices_segments_fixed_bit_set_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_segments_index_writer_memory_in_bytes Count of memory for index writer on this node
             # TYPE elasticsearch_indices_segments_index_writer_memory_in_bytes gauge
             elasticsearch_indices_segments_index_writer_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_segments_memory_bytes Current memory size of segments in bytes
             # TYPE elasticsearch_indices_segments_memory_bytes gauge
             elasticsearch_indices_segments_memory_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 8490
             # HELP elasticsearch_indices_segments_norms_memory_in_bytes Count of memory used by norms
             # TYPE elasticsearch_indices_segments_norms_memory_in_bytes gauge
             elasticsearch_indices_segments_norms_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 640
             # HELP elasticsearch_indices_segments_points_memory_in_bytes Point values memory usage in bytes
             # TYPE elasticsearch_indices_segments_points_memory_in_bytes gauge
             elasticsearch_indices_segments_points_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 5
             # HELP elasticsearch_indices_segments_stored_fields_memory_in_bytes Count of stored fields memory
             # TYPE elasticsearch_indices_segments_stored_fields_memory_in_bytes gauge
             elasticsearch_indices_segments_stored_fields_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1560
             # HELP elasticsearch_indices_segments_term_vectors_memory_in_bytes Term vectors memory usage in bytes
             # TYPE elasticsearch_indices_segments_term_vectors_memory_in_bytes gauge
             elasticsearch_indices_segments_term_vectors_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_segments_terms_memory_in_bytes Count of terms in memory for this node
             # TYPE elasticsearch_indices_segments_terms_memory_in_bytes gauge
             elasticsearch_indices_segments_terms_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 5945
             # HELP elasticsearch_indices_segments_version_map_memory_in_bytes Version map memory usage in bytes
             # TYPE elasticsearch_indices_segments_version_map_memory_in_bytes gauge
             elasticsearch_indices_segments_version_map_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_store_size_bytes Current size of stored index data in bytes
             # TYPE elasticsearch_indices_store_size_bytes gauge
             elasticsearch_indices_store_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 7261
             # HELP elasticsearch_indices_store_throttle_time_seconds_total Throttle time for index store in seconds
             # TYPE elasticsearch_indices_store_throttle_time_seconds_total counter
             elasticsearch_indices_store_throttle_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_translog_operations Total translog operations
             # TYPE elasticsearch_indices_translog_operations counter
             elasticsearch_indices_translog_operations{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 5
             # HELP elasticsearch_indices_translog_size_in_bytes Total translog size in bytes
             # TYPE elasticsearch_indices_translog_size_in_bytes gauge
             elasticsearch_indices_translog_size_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 3753
             # HELP elasticsearch_indices_warmer_time_seconds_total Total warmer time in seconds
             # TYPE elasticsearch_indices_warmer_time_seconds_total counter
             elasticsearch_indices_warmer_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_indices_warmer_total Total warmer count
             # TYPE elasticsearch_indices_warmer_total counter
             elasticsearch_indices_warmer_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 35
             # HELP elasticsearch_jvm_buffer_pool_used_bytes JVM buffer currently used
             # TYPE elasticsearch_jvm_buffer_pool_used_bytes gauge
             elasticsearch_jvm_buffer_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="direct"} 1.68849056e+08
             elasticsearch_jvm_buffer_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="mapped"} 15179
             # HELP elasticsearch_jvm_gc_collection_seconds_count Count of JVM GC runs
             # TYPE elasticsearch_jvm_gc_collection_seconds_count counter
             elasticsearch_jvm_gc_collection_seconds_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="old",host="172.17.0.2",name="9_P7yui"} 0
             elasticsearch_jvm_gc_collection_seconds_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="young",host="172.17.0.2",name="9_P7yui"} 5
             # HELP elasticsearch_jvm_gc_collection_seconds_sum GC run time in seconds
             # TYPE elasticsearch_jvm_gc_collection_seconds_sum counter
             elasticsearch_jvm_gc_collection_seconds_sum{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="old",host="172.17.0.2",name="9_P7yui"} 0
             elasticsearch_jvm_gc_collection_seconds_sum{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="young",host="172.17.0.2",name="9_P7yui"} 0.08
             # HELP elasticsearch_jvm_memory_committed_bytes JVM memory currently committed by area
             # TYPE elasticsearch_jvm_memory_committed_bytes gauge
             elasticsearch_jvm_memory_committed_bytes{area="heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1.073741824e+09
             elasticsearch_jvm_memory_committed_bytes{area="non-heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1.179648e+08
             # HELP elasticsearch_jvm_memory_max_bytes JVM memory max
             # TYPE elasticsearch_jvm_memory_max_bytes gauge
             elasticsearch_jvm_memory_max_bytes{area="heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1.073741824e+09
             # HELP elasticsearch_jvm_memory_pool_max_bytes JVM memory max by pool
             # TYPE elasticsearch_jvm_memory_pool_max_bytes counter
             elasticsearch_jvm_memory_pool_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="old"} 1.073741824e+09
             elasticsearch_jvm_memory_pool_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="survivor"} 0
             elasticsearch_jvm_memory_pool_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="young"} 0
             # HELP elasticsearch_jvm_memory_pool_peak_max_bytes JVM memory peak max by pool
             # TYPE elasticsearch_jvm_memory_pool_peak_max_bytes counter
             elasticsearch_jvm_memory_pool_peak_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="old"} 1.073741824e+09
             elasticsearch_jvm_memory_pool_peak_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="survivor"} 0
             elasticsearch_jvm_memory_pool_peak_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="young"} 0
             # HELP elasticsearch_jvm_memory_pool_peak_used_bytes JVM memory peak used by pool
             # TYPE elasticsearch_jvm_memory_pool_peak_used_bytes counter
             elasticsearch_jvm_memory_pool_peak_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="old"} 2.55827968e+08
             elasticsearch_jvm_memory_pool_peak_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="survivor"} 4.766816e+07
             elasticsearch_jvm_memory_pool_peak_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="young"} 5.59417344e+08
             # HELP elasticsearch_jvm_memory_pool_used_bytes JVM memory currently used by pool
             # TYPE elasticsearch_jvm_memory_pool_used_bytes gauge
             elasticsearch_jvm_memory_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="old"} 2.55827968e+08
             elasticsearch_jvm_memory_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="survivor"} 1.1010048e+07
             elasticsearch_jvm_memory_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",pool="young"} 3.93216e+08
             # HELP elasticsearch_jvm_memory_used_bytes JVM memory currently used by area
             # TYPE elasticsearch_jvm_memory_used_bytes gauge
             elasticsearch_jvm_memory_used_bytes{area="heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 6.60054016e+08
             elasticsearch_jvm_memory_used_bytes{area="non-heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1.08594112e+08
             # HELP elasticsearch_jvm_uptime_seconds JVM process uptime in seconds
             # TYPE elasticsearch_jvm_uptime_seconds gauge
             elasticsearch_jvm_uptime_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="mapped"} 16.456
             # HELP elasticsearch_nodes_roles Node roles
             # TYPE elasticsearch_nodes_roles gauge
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="client"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="data"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="data_cold"} 0
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="data_content"} 0
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="data_frozen"} 0
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="data_hot"} 0
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="data_warm"} 0
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="ingest"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="master"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="ml"} 0
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="remote_cluster_client"} 0
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="9_P7yui",role="transform"} 0
             # HELP elasticsearch_os_cpu_percent Percent CPU used by OS
             # TYPE elasticsearch_os_cpu_percent gauge
             elasticsearch_os_cpu_percent{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 30
             # HELP elasticsearch_os_load1 Shortterm load average
             # TYPE elasticsearch_os_load1 gauge
             elasticsearch_os_load1{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 27.55
             # HELP elasticsearch_os_load15 Longterm load average
             # TYPE elasticsearch_os_load15 gauge
             elasticsearch_os_load15{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 92.65
             # HELP elasticsearch_os_load5 Midterm load average
             # TYPE elasticsearch_os_load5 gauge
             elasticsearch_os_load5{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 134.28
             # HELP elasticsearch_os_mem_actual_free_bytes Amount of free physical memory in bytes
             # TYPE elasticsearch_os_mem_actual_free_bytes gauge
             elasticsearch_os_mem_actual_free_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_os_mem_actual_used_bytes Amount of used physical memory in bytes
             # TYPE elasticsearch_os_mem_actual_used_bytes gauge
             elasticsearch_os_mem_actual_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_os_mem_free_bytes Amount of free physical memory in bytes
             # TYPE elasticsearch_os_mem_free_bytes gauge
             elasticsearch_os_mem_free_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 7.651008512e+09
             # HELP elasticsearch_os_mem_used_bytes Amount of used physical memory in bytes
             # TYPE elasticsearch_os_mem_used_bytes gauge
             elasticsearch_os_mem_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 2.5972228096e+10
             # HELP elasticsearch_process_cpu_percent Percent CPU used by process
             # TYPE elasticsearch_process_cpu_percent gauge
             elasticsearch_process_cpu_percent{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 17
             # HELP elasticsearch_process_cpu_seconds_total Process CPU time in seconds
             # TYPE elasticsearch_process_cpu_seconds_total counter
             elasticsearch_process_cpu_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 32.81
             # HELP elasticsearch_process_max_files_descriptors Max file descriptors
             # TYPE elasticsearch_process_max_files_descriptors gauge
             elasticsearch_process_max_files_descriptors{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 1.048576e+06
             # HELP elasticsearch_process_mem_resident_size_bytes Resident memory in use by process in bytes
             # TYPE elasticsearch_process_mem_resident_size_bytes gauge
             elasticsearch_process_mem_resident_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_process_mem_share_size_bytes Shared memory in use by process in bytes
             # TYPE elasticsearch_process_mem_share_size_bytes gauge
             elasticsearch_process_mem_share_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_process_mem_virtual_size_bytes Total virtual memory used in bytes
             # TYPE elasticsearch_process_mem_virtual_size_bytes gauge
             elasticsearch_process_mem_virtual_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 7.269961728e+09
             # HELP elasticsearch_process_open_files_count Open file descriptors
             # TYPE elasticsearch_process_open_files_count gauge
             elasticsearch_process_open_files_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 355
             # HELP elasticsearch_thread_pool_active_count Thread Pool threads active
             # TYPE elasticsearch_thread_pool_active_count gauge
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="analyze"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ccr"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="flush"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="force_merge"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="generic"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="get"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="index"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="listener"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="management"} 1
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_autodetect"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_datafeed"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_utility"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="refresh"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="rollup_indexing"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search_throttled"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="security-token-key"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="snapshot"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="warmer"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="watcher"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="write"} 0
             # HELP elasticsearch_thread_pool_completed_count Thread Pool operations completed
             # TYPE elasticsearch_thread_pool_completed_count counter
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="analyze"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ccr"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="flush"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="force_merge"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="generic"} 87
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="get"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="index"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="listener"} 5
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="management"} 5
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_autodetect"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_datafeed"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_utility"} 1
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="refresh"} 32
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="rollup_indexing"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search_throttled"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="security-token-key"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="snapshot"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="warmer"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="watcher"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="write"} 5
             # HELP elasticsearch_thread_pool_largest_count Thread Pool largest threads count
             # TYPE elasticsearch_thread_pool_largest_count gauge
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="analyze"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ccr"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="flush"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="force_merge"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="generic"} 5
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="get"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="index"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="listener"} 4
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="management"} 2
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_autodetect"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_datafeed"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_utility"} 1
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="refresh"} 1
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="rollup_indexing"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search_throttled"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="security-token-key"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="snapshot"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="warmer"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="watcher"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="write"} 5
             # HELP elasticsearch_thread_pool_queue_count Thread Pool operations queued
             # TYPE elasticsearch_thread_pool_queue_count gauge
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="analyze"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ccr"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="flush"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="force_merge"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="generic"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="get"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="index"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="listener"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="management"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_autodetect"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_datafeed"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_utility"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="refresh"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="rollup_indexing"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search_throttled"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="security-token-key"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="snapshot"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="warmer"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="watcher"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="write"} 0
             # HELP elasticsearch_thread_pool_rejected_count Thread Pool operations rejected
             # TYPE elasticsearch_thread_pool_rejected_count counter
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="analyze"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ccr"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="flush"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="force_merge"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="generic"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="get"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="index"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="listener"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="management"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_autodetect"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_datafeed"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_utility"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="refresh"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="rollup_indexing"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search_throttled"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="security-token-key"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="snapshot"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="warmer"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="watcher"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="write"} 0
             # HELP elasticsearch_thread_pool_threads_count Thread Pool current threads count
             # TYPE elasticsearch_thread_pool_threads_count gauge
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="analyze"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ccr"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="flush"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="force_merge"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="generic"} 5
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="get"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="index"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="listener"} 4
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="management"} 2
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_autodetect"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_datafeed"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="ml_utility"} 1
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="refresh"} 1
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="rollup_indexing"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="search_throttled"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="security-token-key"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="snapshot"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="warmer"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="watcher"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui",type="write"} 5
             # HELP elasticsearch_transport_rx_packets_total Count of packets received
             # TYPE elasticsearch_transport_rx_packets_total counter
             elasticsearch_transport_rx_packets_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_transport_rx_size_bytes_total Total number of bytes received
             # TYPE elasticsearch_transport_rx_size_bytes_total counter
             elasticsearch_transport_rx_size_bytes_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_transport_tx_packets_total Count of packets sent
             # TYPE elasticsearch_transport_tx_packets_total counter
             elasticsearch_transport_tx_packets_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             # HELP elasticsearch_transport_tx_size_bytes_total Total number of bytes sent
             # TYPE elasticsearch_transport_tx_size_bytes_total counter
             elasticsearch_transport_tx_size_bytes_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="9_P7yui"} 0
             `,
		},
		{
			name: "7.13.1",
			file: "../fixtures/nodestats/7.13.1.json",
			want: `# HELP elasticsearch_breakers_estimated_size_bytes Estimated size in bytes of breaker
             # TYPE elasticsearch_breakers_estimated_size_bytes gauge
             elasticsearch_breakers_estimated_size_bytes{breaker="accounting",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 9380
             elasticsearch_breakers_estimated_size_bytes{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             elasticsearch_breakers_estimated_size_bytes{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             elasticsearch_breakers_estimated_size_bytes{breaker="model_inference",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             elasticsearch_breakers_estimated_size_bytes{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1.56194432e+08
             elasticsearch_breakers_estimated_size_bytes{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_breakers_limit_size_bytes Limit size in bytes for breaker
             # TYPE elasticsearch_breakers_limit_size_bytes gauge
             elasticsearch_breakers_limit_size_bytes{breaker="accounting",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 7.88529152e+08
             elasticsearch_breakers_limit_size_bytes{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 3.1541166e+08
             elasticsearch_breakers_limit_size_bytes{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 7.88529152e+08
             elasticsearch_breakers_limit_size_bytes{breaker="model_inference",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 3.94264576e+08
             elasticsearch_breakers_limit_size_bytes{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 7.49102694e+08
             elasticsearch_breakers_limit_size_bytes{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 4.73117491e+08
             # HELP elasticsearch_breakers_overhead Overhead of circuit breakers
             # TYPE elasticsearch_breakers_overhead counter
             elasticsearch_breakers_overhead{breaker="accounting",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1
             elasticsearch_breakers_overhead{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1.03
             elasticsearch_breakers_overhead{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 2
             elasticsearch_breakers_overhead{breaker="model_inference",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1
             elasticsearch_breakers_overhead{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1
             elasticsearch_breakers_overhead{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1
             # HELP elasticsearch_breakers_tripped tripped for breaker
             # TYPE elasticsearch_breakers_tripped counter
             elasticsearch_breakers_tripped{breaker="accounting",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             elasticsearch_breakers_tripped{breaker="fielddata",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             elasticsearch_breakers_tripped{breaker="in_flight_requests",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             elasticsearch_breakers_tripped{breaker="model_inference",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             elasticsearch_breakers_tripped{breaker="parent",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             elasticsearch_breakers_tripped{breaker="request",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_filesystem_data_available_bytes Available space on block device in bytes
             # TYPE elasticsearch_filesystem_data_available_bytes gauge
             elasticsearch_filesystem_data_available_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",mount="/ (overlay)",name="aaf5a8a0bceb",path="/usr/share/elasticsearch/data/nodes/0"} 6.3425642496e+10
             # HELP elasticsearch_filesystem_data_free_bytes Free space on block device in bytes
             # TYPE elasticsearch_filesystem_data_free_bytes gauge
             elasticsearch_filesystem_data_free_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",mount="/ (overlay)",name="aaf5a8a0bceb",path="/usr/share/elasticsearch/data/nodes/0"} 6.3425642496e+10
             # HELP elasticsearch_filesystem_data_size_bytes Size of block device in bytes
             # TYPE elasticsearch_filesystem_data_size_bytes gauge
             elasticsearch_filesystem_data_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",mount="/ (overlay)",name="aaf5a8a0bceb",path="/usr/share/elasticsearch/data/nodes/0"} 4.76630163456e+11
             # HELP elasticsearch_indexing_pressure_current_all_in_bytes Memory consumed, in bytes, by indexing requests in the coordinating, primary, or replica stage.
             # TYPE elasticsearch_indexing_pressure_current_all_in_bytes gauge
             elasticsearch_indexing_pressure_current_all_in_bytes{cluster="elasticsearch",host="172.17.0.2",indexing_pressure="memory",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indexing_pressure_limit_in_bytes Configured memory limit, in bytes, for the indexing requests
             # TYPE elasticsearch_indexing_pressure_limit_in_bytes gauge
             elasticsearch_indexing_pressure_limit_in_bytes{cluster="elasticsearch",host="172.17.0.2",indexing_pressure="memory",name="aaf5a8a0bceb"} 7.8852915e+07
             # HELP elasticsearch_indices_completion_size_in_bytes Completion in bytes
             # TYPE elasticsearch_indices_completion_size_in_bytes counter
             elasticsearch_indices_completion_size_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_docs Count of documents on this node
             # TYPE elasticsearch_indices_docs gauge
             elasticsearch_indices_docs{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 5
             # HELP elasticsearch_indices_docs_deleted Count of deleted documents on this node
             # TYPE elasticsearch_indices_docs_deleted gauge
             elasticsearch_indices_docs_deleted{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_fielddata_evictions Evictions from field data
             # TYPE elasticsearch_indices_fielddata_evictions counter
             elasticsearch_indices_fielddata_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_fielddata_memory_size_bytes Field data cache memory usage in bytes
             # TYPE elasticsearch_indices_fielddata_memory_size_bytes gauge
             elasticsearch_indices_fielddata_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_filter_cache_evictions Evictions from filter cache
             # TYPE elasticsearch_indices_filter_cache_evictions counter
             elasticsearch_indices_filter_cache_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_filter_cache_memory_size_bytes Filter cache memory usage in bytes
             # TYPE elasticsearch_indices_filter_cache_memory_size_bytes gauge
             elasticsearch_indices_filter_cache_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_flush_time_seconds Cumulative flush time in seconds
             # TYPE elasticsearch_indices_flush_time_seconds counter
             elasticsearch_indices_flush_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_flush_total Total flushes
             # TYPE elasticsearch_indices_flush_total counter
             elasticsearch_indices_flush_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_get_exists_time_seconds Total time get exists in seconds
             # TYPE elasticsearch_indices_get_exists_time_seconds counter
             elasticsearch_indices_get_exists_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_get_exists_total Total get exists operations
             # TYPE elasticsearch_indices_get_exists_total counter
             elasticsearch_indices_get_exists_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_get_missing_time_seconds Total time of get missing in seconds
             # TYPE elasticsearch_indices_get_missing_time_seconds counter
             elasticsearch_indices_get_missing_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_get_missing_total Total get missing
             # TYPE elasticsearch_indices_get_missing_total counter
             elasticsearch_indices_get_missing_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_get_time_seconds Total get time in seconds
             # TYPE elasticsearch_indices_get_time_seconds counter
             elasticsearch_indices_get_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_get_total Total get
             # TYPE elasticsearch_indices_get_total counter
             elasticsearch_indices_get_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_indexing_delete_time_seconds_total Total time indexing delete in seconds
             # TYPE elasticsearch_indices_indexing_delete_time_seconds_total counter
             elasticsearch_indices_indexing_delete_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_indexing_delete_total Total indexing deletes
             # TYPE elasticsearch_indices_indexing_delete_total counter
             elasticsearch_indices_indexing_delete_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_indexing_index_time_seconds_total Cumulative index time in seconds
             # TYPE elasticsearch_indices_indexing_index_time_seconds_total counter
             elasticsearch_indices_indexing_index_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0.014
             # HELP elasticsearch_indices_indexing_index_total Total index calls
             # TYPE elasticsearch_indices_indexing_index_total counter
             elasticsearch_indices_indexing_index_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 5
             # HELP elasticsearch_indices_indexing_is_throttled Indexing throttling
             # TYPE elasticsearch_indices_indexing_is_throttled gauge
             elasticsearch_indices_indexing_is_throttled{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_indexing_throttle_time_seconds_total Cumulative indexing throttling time
             # TYPE elasticsearch_indices_indexing_throttle_time_seconds_total counter
             elasticsearch_indices_indexing_throttle_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_merges_current Current merges
             # TYPE elasticsearch_indices_merges_current gauge
             elasticsearch_indices_merges_current{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_merges_current_size_in_bytes Size of a current merges in bytes
             # TYPE elasticsearch_indices_merges_current_size_in_bytes gauge
             elasticsearch_indices_merges_current_size_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_merges_docs_total Cumulative docs merged
             # TYPE elasticsearch_indices_merges_docs_total counter
             elasticsearch_indices_merges_docs_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_merges_total Total merges
             # TYPE elasticsearch_indices_merges_total counter
             elasticsearch_indices_merges_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_merges_total_size_bytes_total Total merge size in bytes
             # TYPE elasticsearch_indices_merges_total_size_bytes_total counter
             elasticsearch_indices_merges_total_size_bytes_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_merges_total_throttled_time_seconds_total Total throttled time of merges in seconds
             # TYPE elasticsearch_indices_merges_total_throttled_time_seconds_total counter
             elasticsearch_indices_merges_total_throttled_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_merges_total_time_seconds_total Total time spent merging in seconds
             # TYPE elasticsearch_indices_merges_total_time_seconds_total counter
             elasticsearch_indices_merges_total_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_query_cache_cache_size Query cache cache size
             # TYPE elasticsearch_indices_query_cache_cache_size gauge
             elasticsearch_indices_query_cache_cache_size{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_query_cache_cache_total Query cache cache count
             # TYPE elasticsearch_indices_query_cache_cache_total counter
             elasticsearch_indices_query_cache_cache_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_query_cache_count Query cache count
             # TYPE elasticsearch_indices_query_cache_count counter
             elasticsearch_indices_query_cache_count{cache="hit",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_query_cache_evictions Evictions from query cache
             # TYPE elasticsearch_indices_query_cache_evictions counter
             elasticsearch_indices_query_cache_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_query_cache_memory_size_bytes Query cache memory usage in bytes
             # TYPE elasticsearch_indices_query_cache_memory_size_bytes gauge
             elasticsearch_indices_query_cache_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_query_cache_total Query cache total count
             # TYPE elasticsearch_indices_query_cache_total counter
             elasticsearch_indices_query_cache_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_query_miss_count Query miss count
             # TYPE elasticsearch_indices_query_miss_count counter
             elasticsearch_indices_query_miss_count{cache="miss",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_refresh_external_time_seconds_total Total time spent external refreshing in seconds
             # TYPE elasticsearch_indices_refresh_external_time_seconds_total counter
             elasticsearch_indices_refresh_external_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0.15
             # HELP elasticsearch_indices_refresh_external_total Total external refreshes
             # TYPE elasticsearch_indices_refresh_external_total counter
             elasticsearch_indices_refresh_external_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 17
             # HELP elasticsearch_indices_refresh_time_seconds_total Total time spent refreshing in seconds
             # TYPE elasticsearch_indices_refresh_time_seconds_total counter
             elasticsearch_indices_refresh_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0.148
             # HELP elasticsearch_indices_refresh_total Total refreshes
             # TYPE elasticsearch_indices_refresh_total counter
             elasticsearch_indices_refresh_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 17
             # HELP elasticsearch_indices_request_cache_count Request cache count
             # TYPE elasticsearch_indices_request_cache_count counter
             elasticsearch_indices_request_cache_count{cache="hit",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_request_cache_evictions Evictions from request cache
             # TYPE elasticsearch_indices_request_cache_evictions counter
             elasticsearch_indices_request_cache_evictions{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_request_cache_memory_size_bytes Request cache memory usage in bytes
             # TYPE elasticsearch_indices_request_cache_memory_size_bytes gauge
             elasticsearch_indices_request_cache_memory_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_request_miss_count Request miss count
             # TYPE elasticsearch_indices_request_miss_count counter
             elasticsearch_indices_request_miss_count{cache="miss",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_search_fetch_time_seconds Total search fetch time in seconds
             # TYPE elasticsearch_indices_search_fetch_time_seconds counter
             elasticsearch_indices_search_fetch_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_search_fetch_total Total number of fetches
             # TYPE elasticsearch_indices_search_fetch_total counter
             elasticsearch_indices_search_fetch_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_search_query_time_seconds Total search query time in seconds
             # TYPE elasticsearch_indices_search_query_time_seconds counter
             elasticsearch_indices_search_query_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_search_query_total Total number of queries
             # TYPE elasticsearch_indices_search_query_total counter
             elasticsearch_indices_search_query_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_search_scroll_time_seconds Total scroll time in seconds
             # TYPE elasticsearch_indices_search_scroll_time_seconds counter
             elasticsearch_indices_search_scroll_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_search_scroll_total Total number of scrolls
             # TYPE elasticsearch_indices_search_scroll_total counter
             elasticsearch_indices_search_scroll_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_search_suggest_time_seconds Total suggest time in seconds
             # TYPE elasticsearch_indices_search_suggest_time_seconds counter
             elasticsearch_indices_search_suggest_time_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_search_suggest_total Total number of suggests
             # TYPE elasticsearch_indices_search_suggest_total counter
             elasticsearch_indices_search_suggest_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_segments_count Count of index segments on this node
             # TYPE elasticsearch_indices_segments_count gauge
             elasticsearch_indices_segments_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 5
             # HELP elasticsearch_indices_segments_doc_values_memory_in_bytes Count of doc values memory
             # TYPE elasticsearch_indices_segments_doc_values_memory_in_bytes gauge
             elasticsearch_indices_segments_doc_values_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 380
             # HELP elasticsearch_indices_segments_fixed_bit_set_memory_in_bytes Count of fixed bit set
             # TYPE elasticsearch_indices_segments_fixed_bit_set_memory_in_bytes gauge
             elasticsearch_indices_segments_fixed_bit_set_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_segments_index_writer_memory_in_bytes Count of memory for index writer on this node
             # TYPE elasticsearch_indices_segments_index_writer_memory_in_bytes gauge
             elasticsearch_indices_segments_index_writer_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_segments_memory_bytes Current memory size of segments in bytes
             # TYPE elasticsearch_indices_segments_memory_bytes gauge
             elasticsearch_indices_segments_memory_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 9380
             # HELP elasticsearch_indices_segments_norms_memory_in_bytes Count of memory used by norms
             # TYPE elasticsearch_indices_segments_norms_memory_in_bytes gauge
             elasticsearch_indices_segments_norms_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 640
             # HELP elasticsearch_indices_segments_points_memory_in_bytes Point values memory usage in bytes
             # TYPE elasticsearch_indices_segments_points_memory_in_bytes gauge
             elasticsearch_indices_segments_points_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_segments_stored_fields_memory_in_bytes Count of stored fields memory
             # TYPE elasticsearch_indices_segments_stored_fields_memory_in_bytes gauge
             elasticsearch_indices_segments_stored_fields_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 2440
             # HELP elasticsearch_indices_segments_term_vectors_memory_in_bytes Term vectors memory usage in bytes
             # TYPE elasticsearch_indices_segments_term_vectors_memory_in_bytes gauge
             elasticsearch_indices_segments_term_vectors_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_segments_terms_memory_in_bytes Count of terms in memory for this node
             # TYPE elasticsearch_indices_segments_terms_memory_in_bytes gauge
             elasticsearch_indices_segments_terms_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 5920
             # HELP elasticsearch_indices_segments_version_map_memory_in_bytes Version map memory usage in bytes
             # TYPE elasticsearch_indices_segments_version_map_memory_in_bytes gauge
             elasticsearch_indices_segments_version_map_memory_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_store_size_bytes Current size of stored index data in bytes
             # TYPE elasticsearch_indices_store_size_bytes gauge
             elasticsearch_indices_store_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 22296
             # HELP elasticsearch_indices_store_throttle_time_seconds_total Throttle time for index store in seconds
             # TYPE elasticsearch_indices_store_throttle_time_seconds_total counter
             elasticsearch_indices_store_throttle_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_translog_operations Total translog operations
             # TYPE elasticsearch_indices_translog_operations counter
             elasticsearch_indices_translog_operations{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 5
             # HELP elasticsearch_indices_translog_size_in_bytes Total translog size in bytes
             # TYPE elasticsearch_indices_translog_size_in_bytes gauge
             elasticsearch_indices_translog_size_in_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 773
             # HELP elasticsearch_indices_warmer_time_seconds_total Total warmer time in seconds
             # TYPE elasticsearch_indices_warmer_time_seconds_total counter
             elasticsearch_indices_warmer_time_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_indices_warmer_total Total warmer count
             # TYPE elasticsearch_indices_warmer_total counter
             elasticsearch_indices_warmer_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 11
             # HELP elasticsearch_jvm_buffer_pool_used_bytes JVM buffer currently used
             # TYPE elasticsearch_jvm_buffer_pool_used_bytes gauge
             elasticsearch_jvm_buffer_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="direct"} 8.811046e+06
             elasticsearch_jvm_buffer_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="mapped"} 16868
             # HELP elasticsearch_jvm_gc_collection_seconds_count Count of JVM GC runs
             # TYPE elasticsearch_jvm_gc_collection_seconds_count counter
             elasticsearch_jvm_gc_collection_seconds_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="old",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             elasticsearch_jvm_gc_collection_seconds_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="young",host="172.17.0.2",name="aaf5a8a0bceb"} 11
             # HELP elasticsearch_jvm_gc_collection_seconds_sum GC run time in seconds
             # TYPE elasticsearch_jvm_gc_collection_seconds_sum counter
             elasticsearch_jvm_gc_collection_seconds_sum{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="old",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             elasticsearch_jvm_gc_collection_seconds_sum{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",gc="young",host="172.17.0.2",name="aaf5a8a0bceb"} 0.113
             # HELP elasticsearch_jvm_memory_committed_bytes JVM memory currently committed by area
             # TYPE elasticsearch_jvm_memory_committed_bytes gauge
             elasticsearch_jvm_memory_committed_bytes{area="heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 7.88529152e+08
             elasticsearch_jvm_memory_committed_bytes{area="non-heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1.42606336e+08
             # HELP elasticsearch_jvm_memory_max_bytes JVM memory max
             # TYPE elasticsearch_jvm_memory_max_bytes gauge
             elasticsearch_jvm_memory_max_bytes{area="heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 7.88529152e+08
             # HELP elasticsearch_jvm_memory_pool_max_bytes JVM memory max by pool
             # TYPE elasticsearch_jvm_memory_pool_max_bytes counter
             elasticsearch_jvm_memory_pool_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="old"} 7.88529152e+08
             elasticsearch_jvm_memory_pool_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="survivor"} 0
             elasticsearch_jvm_memory_pool_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="young"} 0
             # HELP elasticsearch_jvm_memory_pool_peak_max_bytes JVM memory peak max by pool
             # TYPE elasticsearch_jvm_memory_pool_peak_max_bytes counter
             elasticsearch_jvm_memory_pool_peak_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="old"} 7.88529152e+08
             elasticsearch_jvm_memory_pool_peak_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="survivor"} 0
             elasticsearch_jvm_memory_pool_peak_max_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="young"} 0
             # HELP elasticsearch_jvm_memory_pool_peak_used_bytes JVM memory peak used by pool
             # TYPE elasticsearch_jvm_memory_pool_peak_used_bytes counter
             elasticsearch_jvm_memory_pool_peak_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="old"} 7.1059968e+07
             elasticsearch_jvm_memory_pool_peak_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="survivor"} 5.4525952e+07
             elasticsearch_jvm_memory_pool_peak_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="young"} 4.65567744e+08
             # HELP elasticsearch_jvm_memory_pool_used_bytes JVM memory currently used by pool
             # TYPE elasticsearch_jvm_memory_pool_used_bytes gauge
             elasticsearch_jvm_memory_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="old"} 7.1059968e+07
             elasticsearch_jvm_memory_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="survivor"} 3.0608512e+07
             elasticsearch_jvm_memory_pool_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",pool="young"} 5.4525952e+07
             # HELP elasticsearch_jvm_memory_used_bytes JVM memory currently used by area
             # TYPE elasticsearch_jvm_memory_used_bytes gauge
             elasticsearch_jvm_memory_used_bytes{area="heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1.56194432e+08
             elasticsearch_jvm_memory_used_bytes{area="non-heap",cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1.39526472e+08
             # HELP elasticsearch_jvm_uptime_seconds JVM process uptime in seconds
             # TYPE elasticsearch_jvm_uptime_seconds gauge
             elasticsearch_jvm_uptime_seconds{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="mapped"} 21.844
             # HELP elasticsearch_nodes_roles Node roles
             # TYPE elasticsearch_nodes_roles gauge
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="client"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="data"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="data_cold"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="data_content"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="data_frozen"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="data_hot"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="data_warm"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="ingest"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="master"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="ml"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="remote_cluster_client"} 1
             elasticsearch_nodes_roles{cluster="elasticsearch",host="172.17.0.2",name="aaf5a8a0bceb",role="transform"} 1
             # HELP elasticsearch_os_cpu_percent Percent CPU used by OS
             # TYPE elasticsearch_os_cpu_percent gauge
             elasticsearch_os_cpu_percent{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 37
             # HELP elasticsearch_os_load1 Shortterm load average
             # TYPE elasticsearch_os_load1 gauge
             elasticsearch_os_load1{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 2.74
             # HELP elasticsearch_os_load15 Longterm load average
             # TYPE elasticsearch_os_load15 gauge
             elasticsearch_os_load15{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 70.55
             # HELP elasticsearch_os_load5 Midterm load average
             # TYPE elasticsearch_os_load5 gauge
             elasticsearch_os_load5{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 44.07
             # HELP elasticsearch_os_mem_actual_free_bytes Amount of free physical memory in bytes
             # TYPE elasticsearch_os_mem_actual_free_bytes gauge
             elasticsearch_os_mem_actual_free_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_os_mem_actual_used_bytes Amount of used physical memory in bytes
             # TYPE elasticsearch_os_mem_actual_used_bytes gauge
             elasticsearch_os_mem_actual_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_os_mem_free_bytes Amount of free physical memory in bytes
             # TYPE elasticsearch_os_mem_free_bytes gauge
             elasticsearch_os_mem_free_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1.24007424e+10
             # HELP elasticsearch_os_mem_used_bytes Amount of used physical memory in bytes
             # TYPE elasticsearch_os_mem_used_bytes gauge
             elasticsearch_os_mem_used_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 2.1222490112e+10
             # HELP elasticsearch_process_cpu_percent Percent CPU used by process
             # TYPE elasticsearch_process_cpu_percent gauge
             elasticsearch_process_cpu_percent{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 19
             # HELP elasticsearch_process_cpu_seconds_total Process CPU time in seconds
             # TYPE elasticsearch_process_cpu_seconds_total counter
             elasticsearch_process_cpu_seconds_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 50.16
             # HELP elasticsearch_process_max_files_descriptors Max file descriptors
             # TYPE elasticsearch_process_max_files_descriptors gauge
             elasticsearch_process_max_files_descriptors{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 1.048576e+06
             # HELP elasticsearch_process_mem_resident_size_bytes Resident memory in use by process in bytes
             # TYPE elasticsearch_process_mem_resident_size_bytes gauge
             elasticsearch_process_mem_resident_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_process_mem_share_size_bytes Shared memory in use by process in bytes
             # TYPE elasticsearch_process_mem_share_size_bytes gauge
             elasticsearch_process_mem_share_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_process_mem_virtual_size_bytes Total virtual memory used in bytes
             # TYPE elasticsearch_process_mem_virtual_size_bytes gauge
             elasticsearch_process_mem_virtual_size_bytes{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 6.819479552e+09
             # HELP elasticsearch_process_open_files_count Open file descriptors
             # TYPE elasticsearch_process_open_files_count gauge
             elasticsearch_process_open_files_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 314
             # HELP elasticsearch_thread_pool_active_count Thread Pool threads active
             # TYPE elasticsearch_thread_pool_active_count gauge
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="analyze"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ccr"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="flush"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="force_merge"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="generic"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="get"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="listener"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="management"} 1
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_datafeed"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_job_comms"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_utility"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="refresh"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="rollup_indexing"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search_throttled"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_fetch_async"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_prewarming"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-crypto"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-token-key"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="snapshot"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_read"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_write"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="transform_indexing"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="warmer"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="watcher"} 0
             elasticsearch_thread_pool_active_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="write"} 0
             # HELP elasticsearch_thread_pool_completed_count Thread Pool operations completed
             # TYPE elasticsearch_thread_pool_completed_count counter
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="analyze"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ccr"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="flush"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="force_merge"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="generic"} 406
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="get"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="listener"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="management"} 9
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_datafeed"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_job_comms"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_utility"} 12
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="refresh"} 36
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="rollup_indexing"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search_throttled"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_fetch_async"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_prewarming"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-crypto"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-token-key"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="snapshot"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_read"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_write"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="transform_indexing"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="warmer"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="watcher"} 0
             elasticsearch_thread_pool_completed_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="write"} 9
             # HELP elasticsearch_thread_pool_largest_count Thread Pool largest threads count
             # TYPE elasticsearch_thread_pool_largest_count gauge
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="analyze"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ccr"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="flush"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="force_merge"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="generic"} 7
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="get"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="listener"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="management"} 2
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_datafeed"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_job_comms"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_utility"} 1
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="refresh"} 1
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="rollup_indexing"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search_throttled"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_fetch_async"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_prewarming"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-crypto"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-token-key"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="snapshot"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_read"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_write"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="transform_indexing"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="warmer"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="watcher"} 0
             elasticsearch_thread_pool_largest_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="write"} 8
             # HELP elasticsearch_thread_pool_queue_count Thread Pool operations queued
             # TYPE elasticsearch_thread_pool_queue_count gauge
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="analyze"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ccr"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="flush"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="force_merge"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="generic"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="get"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="listener"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="management"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_datafeed"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_job_comms"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_utility"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="refresh"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="rollup_indexing"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search_throttled"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_fetch_async"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_prewarming"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-crypto"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-token-key"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="snapshot"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_read"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_write"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="transform_indexing"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="warmer"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="watcher"} 0
             elasticsearch_thread_pool_queue_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="write"} 0
             # HELP elasticsearch_thread_pool_rejected_count Thread Pool operations rejected
             # TYPE elasticsearch_thread_pool_rejected_count counter
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="analyze"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ccr"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="flush"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="force_merge"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="generic"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="get"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="listener"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="management"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_datafeed"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_job_comms"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_utility"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="refresh"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="rollup_indexing"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search_throttled"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_fetch_async"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_prewarming"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-crypto"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-token-key"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="snapshot"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_read"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_write"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="transform_indexing"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="warmer"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="watcher"} 0
             elasticsearch_thread_pool_rejected_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="write"} 0
             # HELP elasticsearch_thread_pool_threads_count Thread Pool current threads count
             # TYPE elasticsearch_thread_pool_threads_count gauge
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="analyze"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ccr"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_started"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="fetch_shard_store"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="flush"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="force_merge"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="generic"} 7
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="get"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="listener"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="management"} 2
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_datafeed"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_job_comms"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="ml_utility"} 1
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="refresh"} 1
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="rollup_indexing"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="search_throttled"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_fetch_async"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="searchable_snapshots_cache_prewarming"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-crypto"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="security-token-key"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="snapshot"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_read"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="system_write"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="transform_indexing"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="warmer"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="watcher"} 0
             elasticsearch_thread_pool_threads_count{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb",type="write"} 8
             # HELP elasticsearch_transport_rx_packets_total Count of packets received
             # TYPE elasticsearch_transport_rx_packets_total counter
             elasticsearch_transport_rx_packets_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_transport_rx_size_bytes_total Total number of bytes received
             # TYPE elasticsearch_transport_rx_size_bytes_total counter
             elasticsearch_transport_rx_size_bytes_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_transport_tx_packets_total Count of packets sent
             # TYPE elasticsearch_transport_tx_packets_total counter
             elasticsearch_transport_tx_packets_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
             # HELP elasticsearch_transport_tx_size_bytes_total Total number of bytes sent
             # TYPE elasticsearch_transport_tx_size_bytes_total counter
             elasticsearch_transport_tx_size_bytes_total{cluster="elasticsearch",es_client_node="true",es_data_node="true",es_ingest_node="true",es_master_node="true",host="172.17.0.2",name="aaf5a8a0bceb"} 0
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

			c := NewNodes(promslog.NewNopLogger(), http.DefaultClient, u, true, "_local")
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(c, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
