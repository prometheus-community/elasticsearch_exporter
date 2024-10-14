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
	"path"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func TestIndices(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine
	//  curl -XPUT http://localhost:9200/foo_1/type1/1 -d '{"title":"abc","content":"hello"}'
	//  curl -XPUT http://localhost:9200/foo_1/type1/2 -d '{"title":"def","content":"world"}'
	//  curl -XPUT http://localhost:9200/foo_2/type1/1 -d '{"title":"abc001","content":"hello001"}'
	//  curl -XPUT http://localhost:9200/foo_2/type1/2 -d '{"title":"def002","content":"world002"}'
	//  curl -XPUT http://localhost:9200/foo_2/type1/3 -d '{"title":"def003","content":"world003"}'
	//  curl -XPOST -H "Content-Type: application/json" http://localhost:9200/_aliases -d '{"actions": [{"add": {"index": "foo_2","alias": "foo_alias_2_1"}}]}'
	//  curl -XPOST -H "Content-Type: application/json" http://localhost:9200/_aliases -d '{"actions": [{"add": {"index": "foo_3","alias": "foo_alias_3_2"}}]}'
	//  curl -XPOST -H "Content-Type: application/json" http://localhost:9200/_aliases -d '{"actions": [{"add": {"index": "foo_3","alias": "foo_alias_3_1", "is_write_index": true, "routing": "title"}}]}'
	//  curl http://localhost:9200/_all/_stats
	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "1.7.6",
			file: "1.7.6.json",
			want: `# HELP elasticsearch_index_stats_fielddata_evictions_total Total fielddata evictions count
             # TYPE elasticsearch_index_stats_fielddata_evictions_total counter
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_fielddata_memory_bytes_total Total fielddata memory bytes
             # TYPE elasticsearch_index_stats_fielddata_memory_bytes_total counter
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_flush_time_seconds_total Total flush time in seconds
             # TYPE elasticsearch_index_stats_flush_time_seconds_total counter
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_flush_total Total flush count
             # TYPE elasticsearch_index_stats_flush_total counter
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_get_time_seconds_total Total get time in seconds
             # TYPE elasticsearch_index_stats_get_time_seconds_total counter
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_get_total Total get count
             # TYPE elasticsearch_index_stats_get_total counter
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_index_current The number of documents currently being indexed to an index
             # TYPE elasticsearch_index_stats_index_current gauge
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_delete_time_seconds_total Total indexing delete time in seconds
             # TYPE elasticsearch_index_stats_indexing_delete_time_seconds_total counter
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_delete_total Total indexing delete count
             # TYPE elasticsearch_index_stats_indexing_delete_total counter
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_index_time_seconds_total Total indexing index time in seconds
             # TYPE elasticsearch_index_stats_indexing_index_time_seconds_total counter
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.046
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0.006
             # HELP elasticsearch_index_stats_indexing_index_total Total indexing index count
             # TYPE elasticsearch_index_stats_indexing_index_total counter
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_index_stats_indexing_noop_update_total Total indexing no-op update count
             # TYPE elasticsearch_index_stats_indexing_noop_update_total counter
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_throttle_time_seconds_total Total indexing throttle time in seconds
             # TYPE elasticsearch_index_stats_indexing_throttle_time_seconds_total counter
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_auto_throttle_bytes_total Total bytes that were auto-throttled during merging
             # TYPE elasticsearch_index_stats_merge_auto_throttle_bytes_total counter
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_stopped_time_seconds_total Total large merge stopped time in seconds, allowing smaller merges to complete
             # TYPE elasticsearch_index_stats_merge_stopped_time_seconds_total counter
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_throttle_time_seconds_total Total merge I/O throttle time in seconds
             # TYPE elasticsearch_index_stats_merge_throttle_time_seconds_total counter
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_time_seconds_total Total merge time in seconds
             # TYPE elasticsearch_index_stats_merge_time_seconds_total counter
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_total Total merge count
             # TYPE elasticsearch_index_stats_merge_total counter
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_caches_total Total query cache caches count
             # TYPE elasticsearch_index_stats_query_cache_caches_total counter
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_evictions_total Total query cache evictions count
             # TYPE elasticsearch_index_stats_query_cache_evictions_total counter
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_hits_total Total query cache hits count
             # TYPE elasticsearch_index_stats_query_cache_hits_total counter
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_memory_bytes_total Total query cache memory bytes
             # TYPE elasticsearch_index_stats_query_cache_memory_bytes_total counter
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_misses_total Total query cache misses count
             # TYPE elasticsearch_index_stats_query_cache_misses_total counter
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_size Total query cache size
             # TYPE elasticsearch_index_stats_query_cache_size gauge
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_refresh_external_time_seconds_total Total external refresh time in seconds
             # TYPE elasticsearch_index_stats_refresh_external_time_seconds_total counter
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_refresh_external_total Total external refresh count
             # TYPE elasticsearch_index_stats_refresh_external_total counter
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_refresh_time_seconds_total Total refresh time in seconds
             # TYPE elasticsearch_index_stats_refresh_time_seconds_total counter
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.125
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0.038
             # HELP elasticsearch_index_stats_refresh_total Total refresh count
             # TYPE elasticsearch_index_stats_refresh_total counter
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_index_stats_request_cache_evictions_total Total request cache evictions count
             # TYPE elasticsearch_index_stats_request_cache_evictions_total counter
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_request_cache_hits_total Total request cache hits count
             # TYPE elasticsearch_index_stats_request_cache_hits_total counter
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_request_cache_memory_bytes_total Total request cache memory bytes
             # TYPE elasticsearch_index_stats_request_cache_memory_bytes_total counter
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_request_cache_misses_total Total request cache misses count
             # TYPE elasticsearch_index_stats_request_cache_misses_total counter
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_fetch_time_seconds_total Total search fetch time in seconds
             # TYPE elasticsearch_index_stats_search_fetch_time_seconds_total counter
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_fetch_total Total search fetch count
             # TYPE elasticsearch_index_stats_search_fetch_total counter
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_query_time_seconds_total Total search query time in seconds
             # TYPE elasticsearch_index_stats_search_query_time_seconds_total counter
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_query_total Total number of queries
             # TYPE elasticsearch_index_stats_search_query_total counter
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_scroll_current Current search scroll count
             # TYPE elasticsearch_index_stats_search_scroll_current gauge
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_scroll_time_seconds_total Total search scroll time in seconds
             # TYPE elasticsearch_index_stats_search_scroll_time_seconds_total counter
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_scroll_total Total search scroll count
             # TYPE elasticsearch_index_stats_search_scroll_total counter
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_suggest_time_seconds_total Total search suggest time in seconds
             # TYPE elasticsearch_index_stats_search_suggest_time_seconds_total counter
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_suggest_total Total search suggest count
             # TYPE elasticsearch_index_stats_search_suggest_total counter
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_warmer_time_seconds_total Total warmer time in seconds
             # TYPE elasticsearch_index_stats_warmer_time_seconds_total counter
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.042
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_warmer_total Total warmer count
             # TYPE elasticsearch_index_stats_warmer_total counter
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index="foo_1"} 14
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index="foo_2"} 16
						 # HELP elasticsearch_indices_aliases Record aliases associated with an index
             # TYPE elasticsearch_indices_aliases gauge
             elasticsearch_indices_aliases{alias="foo_alias_2_1",cluster="unknown_cluster",index="foo_2"} 1
             elasticsearch_indices_aliases{alias="foo_alias_3_1",cluster="unknown_cluster",index="foo_3"} 1
             elasticsearch_indices_aliases{alias="foo_alias_3_2",cluster="unknown_cluster",index="foo_3"} 1
             # HELP elasticsearch_indices_completion_bytes_primary Current size of completion with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_completion_bytes_primary gauge
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_completion_bytes_total Current size of completion with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_completion_bytes_total gauge
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_deleted_docs_primary Count of deleted documents with only primary shards
             # TYPE elasticsearch_indices_deleted_docs_primary gauge
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_deleted_docs_total Total count of deleted documents
             # TYPE elasticsearch_indices_deleted_docs_total gauge
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_docs_primary Count of documents with only primary shards
             # TYPE elasticsearch_indices_docs_primary gauge
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_docs_total Total count of documents
             # TYPE elasticsearch_indices_docs_total gauge
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_segment_count_primary Current number of segments with only primary shards on all nodes
             # TYPE elasticsearch_indices_segment_count_primary gauge
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_segment_count_total Current number of segments with all shards on all nodes
             # TYPE elasticsearch_indices_segment_count_total gauge
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_segment_doc_values_memory_bytes_primary Current size of doc values with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_doc_values_memory_bytes_primary gauge
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_doc_values_memory_bytes_total Current size of doc values with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_doc_values_memory_bytes_total gauge
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_fields_memory_bytes_primary Current size of fields with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fields_memory_bytes_primary gauge
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_fields_memory_bytes_total Current size of fields with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fields_memory_bytes_total gauge
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary Current size of fixed bit with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary gauge
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total Current size of fixed bit with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total gauge
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_index_writer_memory_bytes_primary Current size of index writer with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_index_writer_memory_bytes_primary gauge
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_index_writer_memory_bytes_total Current size of index writer with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_index_writer_memory_bytes_total gauge
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_memory_bytes_primary Current size of segments with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_memory_bytes_primary gauge
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 7364
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 11046
             # HELP elasticsearch_indices_segment_memory_bytes_total Current size of segments with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_memory_bytes_total gauge
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 7364
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 11046
             # HELP elasticsearch_indices_segment_norms_memory_bytes_primary Current size of norms with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_norms_memory_bytes_primary gauge
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_norms_memory_bytes_total Current size of norms with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_norms_memory_bytes_total gauge
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_points_memory_bytes_primary Current size of points with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_points_memory_bytes_primary gauge
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_points_memory_bytes_total Current size of points with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_points_memory_bytes_total gauge
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_term_vectors_memory_primary_bytes Current size of term vectors with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_term_vectors_memory_primary_bytes gauge
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_term_vectors_memory_total_bytes Current size of term vectors with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_term_vectors_memory_total_bytes gauge
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_terms_memory_primary Current size of terms with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_terms_memory_primary gauge
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_terms_memory_total Current number of terms with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_terms_memory_total gauge
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_version_map_memory_bytes_primary Current size of version map with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_version_map_memory_bytes_primary gauge
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_version_map_memory_bytes_total Current size of version map with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_version_map_memory_bytes_total gauge
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_store_size_bytes_primary Current total size of stored index data in bytes with only primary shards on all nodes
             # TYPE elasticsearch_indices_store_size_bytes_primary gauge
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index="foo_1"} 5591
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index="foo_2"} 8207
             # HELP elasticsearch_indices_store_size_bytes_total Current total size of stored index data in bytes with all shards on all nodes
             # TYPE elasticsearch_indices_store_size_bytes_total gauge
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index="foo_1"} 5591
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index="foo_2"} 8207
             # HELP elasticsearch_search_active_queries The number of currently active queries
             # TYPE elasticsearch_search_active_queries gauge
             elasticsearch_search_active_queries{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_search_active_queries{cluster="unknown_cluster",index="foo_2"} 0
`,
		},
		{
			name: "2.4.5",
			file: "2.4.5.json",
			want: `# HELP elasticsearch_index_stats_fielddata_evictions_total Total fielddata evictions count
             # TYPE elasticsearch_index_stats_fielddata_evictions_total counter
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_fielddata_memory_bytes_total Total fielddata memory bytes
             # TYPE elasticsearch_index_stats_fielddata_memory_bytes_total counter
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_flush_time_seconds_total Total flush time in seconds
             # TYPE elasticsearch_index_stats_flush_time_seconds_total counter
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_flush_total Total flush count
             # TYPE elasticsearch_index_stats_flush_total counter
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_get_time_seconds_total Total get time in seconds
             # TYPE elasticsearch_index_stats_get_time_seconds_total counter
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_get_total Total get count
             # TYPE elasticsearch_index_stats_get_total counter
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_index_current The number of documents currently being indexed to an index
             # TYPE elasticsearch_index_stats_index_current gauge
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_delete_time_seconds_total Total indexing delete time in seconds
             # TYPE elasticsearch_index_stats_indexing_delete_time_seconds_total counter
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_delete_total Total indexing delete count
             # TYPE elasticsearch_index_stats_indexing_delete_total counter
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_index_time_seconds_total Total indexing index time in seconds
             # TYPE elasticsearch_index_stats_indexing_index_time_seconds_total counter
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.034
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0.006
             # HELP elasticsearch_index_stats_indexing_index_total Total indexing index count
             # TYPE elasticsearch_index_stats_indexing_index_total counter
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_index_stats_indexing_noop_update_total Total indexing no-op update count
             # TYPE elasticsearch_index_stats_indexing_noop_update_total counter
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_throttle_time_seconds_total Total indexing throttle time in seconds
             # TYPE elasticsearch_index_stats_indexing_throttle_time_seconds_total counter
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_auto_throttle_bytes_total Total bytes that were auto-throttled during merging
             # TYPE elasticsearch_index_stats_merge_auto_throttle_bytes_total counter
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index="foo_1"} 1.048576e+08
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index="foo_2"} 1.048576e+08
             # HELP elasticsearch_index_stats_merge_stopped_time_seconds_total Total large merge stopped time in seconds, allowing smaller merges to complete
             # TYPE elasticsearch_index_stats_merge_stopped_time_seconds_total counter
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_throttle_time_seconds_total Total merge I/O throttle time in seconds
             # TYPE elasticsearch_index_stats_merge_throttle_time_seconds_total counter
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_time_seconds_total Total merge time in seconds
             # TYPE elasticsearch_index_stats_merge_time_seconds_total counter
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_total Total merge count
             # TYPE elasticsearch_index_stats_merge_total counter
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_caches_total Total query cache caches count
             # TYPE elasticsearch_index_stats_query_cache_caches_total counter
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_evictions_total Total query cache evictions count
             # TYPE elasticsearch_index_stats_query_cache_evictions_total counter
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_hits_total Total query cache hits count
             # TYPE elasticsearch_index_stats_query_cache_hits_total counter
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_memory_bytes_total Total query cache memory bytes
             # TYPE elasticsearch_index_stats_query_cache_memory_bytes_total counter
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_misses_total Total query cache misses count
             # TYPE elasticsearch_index_stats_query_cache_misses_total counter
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_size Total query cache size
             # TYPE elasticsearch_index_stats_query_cache_size gauge
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_refresh_external_time_seconds_total Total external refresh time in seconds
             # TYPE elasticsearch_index_stats_refresh_external_time_seconds_total counter
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_refresh_external_total Total external refresh count
             # TYPE elasticsearch_index_stats_refresh_external_total counter
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_refresh_time_seconds_total Total refresh time in seconds
             # TYPE elasticsearch_index_stats_refresh_time_seconds_total counter
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.137
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0.034
             # HELP elasticsearch_index_stats_refresh_total Total refresh count
             # TYPE elasticsearch_index_stats_refresh_total counter
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_index_stats_request_cache_evictions_total Total request cache evictions count
             # TYPE elasticsearch_index_stats_request_cache_evictions_total counter
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_request_cache_hits_total Total request cache hits count
             # TYPE elasticsearch_index_stats_request_cache_hits_total counter
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_request_cache_memory_bytes_total Total request cache memory bytes
             # TYPE elasticsearch_index_stats_request_cache_memory_bytes_total counter
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_request_cache_misses_total Total request cache misses count
             # TYPE elasticsearch_index_stats_request_cache_misses_total counter
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_fetch_time_seconds_total Total search fetch time in seconds
             # TYPE elasticsearch_index_stats_search_fetch_time_seconds_total counter
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_fetch_total Total search fetch count
             # TYPE elasticsearch_index_stats_search_fetch_total counter
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_query_time_seconds_total Total search query time in seconds
             # TYPE elasticsearch_index_stats_search_query_time_seconds_total counter
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_query_total Total number of queries
             # TYPE elasticsearch_index_stats_search_query_total counter
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_scroll_current Current search scroll count
             # TYPE elasticsearch_index_stats_search_scroll_current gauge
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_scroll_time_seconds_total Total search scroll time in seconds
             # TYPE elasticsearch_index_stats_search_scroll_time_seconds_total counter
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_scroll_total Total search scroll count
             # TYPE elasticsearch_index_stats_search_scroll_total counter
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_suggest_time_seconds_total Total search suggest time in seconds
             # TYPE elasticsearch_index_stats_search_suggest_time_seconds_total counter
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_suggest_total Total search suggest count
             # TYPE elasticsearch_index_stats_search_suggest_total counter
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_warmer_time_seconds_total Total warmer time in seconds
             # TYPE elasticsearch_index_stats_warmer_time_seconds_total counter
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.012
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_warmer_total Total warmer count
             # TYPE elasticsearch_index_stats_warmer_total counter
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index="foo_1"} 14
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index="foo_2"} 16
             # HELP elasticsearch_indices_aliases Record aliases associated with an index
             # TYPE elasticsearch_indices_aliases gauge
             elasticsearch_indices_aliases{alias="foo_alias_2_1",cluster="unknown_cluster",index="foo_2"} 1
             elasticsearch_indices_aliases{alias="foo_alias_3_1",cluster="unknown_cluster",index="foo_3"} 1
             elasticsearch_indices_aliases{alias="foo_alias_3_2",cluster="unknown_cluster",index="foo_3"} 1
             # HELP elasticsearch_indices_completion_bytes_primary Current size of completion with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_completion_bytes_primary gauge
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_completion_bytes_total Current size of completion with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_completion_bytes_total gauge
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_deleted_docs_primary Count of deleted documents with only primary shards
             # TYPE elasticsearch_indices_deleted_docs_primary gauge
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_deleted_docs_total Total count of deleted documents
             # TYPE elasticsearch_indices_deleted_docs_total gauge
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_docs_primary Count of documents with only primary shards
             # TYPE elasticsearch_indices_docs_primary gauge
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_docs_total Total count of documents
             # TYPE elasticsearch_indices_docs_total gauge
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_segment_count_primary Current number of segments with only primary shards on all nodes
             # TYPE elasticsearch_indices_segment_count_primary gauge
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_segment_count_total Current number of segments with all shards on all nodes
             # TYPE elasticsearch_indices_segment_count_total gauge
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_segment_doc_values_memory_bytes_primary Current size of doc values with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_doc_values_memory_bytes_primary gauge
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 184
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 276
             # HELP elasticsearch_indices_segment_doc_values_memory_bytes_total Current size of doc values with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_doc_values_memory_bytes_total gauge
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 184
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 276
             # HELP elasticsearch_indices_segment_fields_memory_bytes_primary Current size of fields with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fields_memory_bytes_primary gauge
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 624
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 936
             # HELP elasticsearch_indices_segment_fields_memory_bytes_total Current size of fields with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fields_memory_bytes_total gauge
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 624
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 936
             # HELP elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary Current size of fixed bit with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary gauge
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total Current size of fixed bit with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total gauge
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_index_writer_memory_bytes_primary Current size of index writer with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_index_writer_memory_bytes_primary gauge
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_index_writer_memory_bytes_total Current size of index writer with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_index_writer_memory_bytes_total gauge
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_memory_bytes_primary Current size of segments with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_memory_bytes_primary gauge
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 4212
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 6318
             # HELP elasticsearch_indices_segment_memory_bytes_total Current size of segments with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_memory_bytes_total gauge
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 4212
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 6318
             # HELP elasticsearch_indices_segment_norms_memory_bytes_primary Current size of norms with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_norms_memory_bytes_primary gauge
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 384
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 576
             # HELP elasticsearch_indices_segment_norms_memory_bytes_total Current size of norms with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_norms_memory_bytes_total gauge
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 384
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 576
             # HELP elasticsearch_indices_segment_points_memory_bytes_primary Current size of points with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_points_memory_bytes_primary gauge
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_points_memory_bytes_total Current size of points with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_points_memory_bytes_total gauge
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_term_vectors_memory_primary_bytes Current size of term vectors with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_term_vectors_memory_primary_bytes gauge
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_term_vectors_memory_total_bytes Current size of term vectors with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_term_vectors_memory_total_bytes gauge
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_terms_memory_primary Current size of terms with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_terms_memory_primary gauge
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index="foo_1"} 3020
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index="foo_2"} 4530
             # HELP elasticsearch_indices_segment_terms_memory_total Current number of terms with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_terms_memory_total gauge
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index="foo_1"} 3020
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index="foo_2"} 4530
             # HELP elasticsearch_indices_segment_version_map_memory_bytes_primary Current size of version map with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_version_map_memory_bytes_primary gauge
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_version_map_memory_bytes_total Current size of version map with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_version_map_memory_bytes_total gauge
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_store_size_bytes_primary Current total size of stored index data in bytes with only primary shards on all nodes
             # TYPE elasticsearch_indices_store_size_bytes_primary gauge
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index="foo_1"} 260
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index="foo_2"} 3350
             # HELP elasticsearch_indices_store_size_bytes_total Current total size of stored index data in bytes with all shards on all nodes
             # TYPE elasticsearch_indices_store_size_bytes_total gauge
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index="foo_1"} 260
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index="foo_2"} 3350
             # HELP elasticsearch_search_active_queries The number of currently active queries
             # TYPE elasticsearch_search_active_queries gauge
             elasticsearch_search_active_queries{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_search_active_queries{cluster="unknown_cluster",index="foo_2"} 0

`,
		},
		{
			name: "5.4.2",
			file: "5.4.2.json",
			want: `# HELP elasticsearch_index_stats_fielddata_evictions_total Total fielddata evictions count
             # TYPE elasticsearch_index_stats_fielddata_evictions_total counter
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_fielddata_memory_bytes_total Total fielddata memory bytes
             # TYPE elasticsearch_index_stats_fielddata_memory_bytes_total counter
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_flush_time_seconds_total Total flush time in seconds
             # TYPE elasticsearch_index_stats_flush_time_seconds_total counter
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_flush_total Total flush count
             # TYPE elasticsearch_index_stats_flush_total counter
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_get_time_seconds_total Total get time in seconds
             # TYPE elasticsearch_index_stats_get_time_seconds_total counter
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_get_total Total get count
             # TYPE elasticsearch_index_stats_get_total counter
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_index_current The number of documents currently being indexed to an index
             # TYPE elasticsearch_index_stats_index_current gauge
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_delete_time_seconds_total Total indexing delete time in seconds
             # TYPE elasticsearch_index_stats_indexing_delete_time_seconds_total counter
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_delete_total Total indexing delete count
             # TYPE elasticsearch_index_stats_indexing_delete_total counter
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_index_time_seconds_total Total indexing index time in seconds
             # TYPE elasticsearch_index_stats_indexing_index_time_seconds_total counter
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0.013
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0.106
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index=".watches"} 1.421
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.046
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0.012
             # HELP elasticsearch_index_stats_indexing_index_total Total indexing index count
             # TYPE elasticsearch_index_stats_indexing_index_total counter
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index=".monitoring-data-2"} 4
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 65
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index=".watches"} 4
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_index_stats_indexing_noop_update_total Total indexing no-op update count
             # TYPE elasticsearch_index_stats_indexing_noop_update_total counter
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_indexing_throttle_time_seconds_total Total indexing throttle time in seconds
             # TYPE elasticsearch_index_stats_indexing_throttle_time_seconds_total counter
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_auto_throttle_bytes_total Total bytes that were auto-throttled during merging
             # TYPE elasticsearch_index_stats_merge_auto_throttle_bytes_total counter
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 2.097152e+07
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 2.097152e+07
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index=".watches"} 2.097152e+07
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index="foo_1"} 1.048576e+08
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index="foo_2"} 1.048576e+08
             # HELP elasticsearch_index_stats_merge_stopped_time_seconds_total Total large merge stopped time in seconds, allowing smaller merges to complete
             # TYPE elasticsearch_index_stats_merge_stopped_time_seconds_total counter
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_throttle_time_seconds_total Total merge I/O throttle time in seconds
             # TYPE elasticsearch_index_stats_merge_throttle_time_seconds_total counter
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_time_seconds_total Total merge time in seconds
             # TYPE elasticsearch_index_stats_merge_time_seconds_total counter
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_merge_total Total merge count
             # TYPE elasticsearch_index_stats_merge_total counter
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_caches_total Total query cache caches count
             # TYPE elasticsearch_index_stats_query_cache_caches_total counter
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_evictions_total Total query cache evictions count
             # TYPE elasticsearch_index_stats_query_cache_evictions_total counter
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_hits_total Total query cache hits count
             # TYPE elasticsearch_index_stats_query_cache_hits_total counter
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_memory_bytes_total Total query cache memory bytes
             # TYPE elasticsearch_index_stats_query_cache_memory_bytes_total counter
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_misses_total Total query cache misses count
             # TYPE elasticsearch_index_stats_query_cache_misses_total counter
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_query_cache_size Total query cache size
             # TYPE elasticsearch_index_stats_query_cache_size gauge
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_refresh_external_time_seconds_total Total external refresh time in seconds
             # TYPE elasticsearch_index_stats_refresh_external_time_seconds_total counter
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_refresh_external_total Total external refresh count
             # TYPE elasticsearch_index_stats_refresh_external_total counter
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_refresh_time_seconds_total Total refresh time in seconds
             # TYPE elasticsearch_index_stats_refresh_time_seconds_total counter
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0.074
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0.39
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0.771
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.084
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0.042
             # HELP elasticsearch_index_stats_refresh_total Total refresh count
             # TYPE elasticsearch_index_stats_refresh_total counter
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index=".monitoring-data-2"} 2
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 3
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index=".watches"} 5
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_index_stats_request_cache_evictions_total Total request cache evictions count
             # TYPE elasticsearch_index_stats_request_cache_evictions_total counter
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_request_cache_hits_total Total request cache hits count
             # TYPE elasticsearch_index_stats_request_cache_hits_total counter
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_request_cache_memory_bytes_total Total request cache memory bytes
             # TYPE elasticsearch_index_stats_request_cache_memory_bytes_total counter
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_request_cache_misses_total Total request cache misses count
             # TYPE elasticsearch_index_stats_request_cache_misses_total counter
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_fetch_time_seconds_total Total search fetch time in seconds
             # TYPE elasticsearch_index_stats_search_fetch_time_seconds_total counter
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_fetch_total Total search fetch count
             # TYPE elasticsearch_index_stats_search_fetch_total counter
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_query_time_seconds_total Total search query time in seconds
             # TYPE elasticsearch_index_stats_search_query_time_seconds_total counter
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_query_total Total number of queries
             # TYPE elasticsearch_index_stats_search_query_total counter
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_scroll_current Current search scroll count
             # TYPE elasticsearch_index_stats_search_scroll_current gauge
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_scroll_time_seconds_total Total search scroll time in seconds
             # TYPE elasticsearch_index_stats_search_scroll_time_seconds_total counter
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_scroll_total Total search scroll count
             # TYPE elasticsearch_index_stats_search_scroll_total counter
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_suggest_time_seconds_total Total search suggest time in seconds
             # TYPE elasticsearch_index_stats_search_suggest_time_seconds_total counter
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_search_suggest_total Total search suggest count
             # TYPE elasticsearch_index_stats_search_suggest_total counter
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_index_stats_warmer_time_seconds_total Total warmer time in seconds
             # TYPE elasticsearch_index_stats_warmer_time_seconds_total counter
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0.002
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0.015
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index=".watches"} 0.009
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.094
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0.004
             # HELP elasticsearch_index_stats_warmer_total Total warmer count
             # TYPE elasticsearch_index_stats_warmer_total counter
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index=".monitoring-data-2"} 3
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 4
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index=".watches"} 4
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index="foo_1"} 7
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index="foo_2"} 8
             # HELP elasticsearch_indices_aliases Record aliases associated with an index
             # TYPE elasticsearch_indices_aliases gauge
             elasticsearch_indices_aliases{alias="foo_alias_2_1",cluster="unknown_cluster",index="foo_2"} 1
             elasticsearch_indices_aliases{alias="foo_alias_3_1",cluster="unknown_cluster",index="foo_3"} 1
             elasticsearch_indices_aliases{alias="foo_alias_3_2",cluster="unknown_cluster",index="foo_3"} 1
             # HELP elasticsearch_indices_completion_bytes_primary Current size of completion with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_completion_bytes_primary gauge
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_completion_bytes_total Current size of completion with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_completion_bytes_total gauge
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_deleted_docs_primary Count of deleted documents with only primary shards
             # TYPE elasticsearch_indices_deleted_docs_primary gauge
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_deleted_docs_total Total count of deleted documents
             # TYPE elasticsearch_indices_deleted_docs_total gauge
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_docs_primary Count of documents with only primary shards
             # TYPE elasticsearch_indices_docs_primary gauge
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 2
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 65
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index=".watches"} 4
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_docs_total Total count of documents
             # TYPE elasticsearch_indices_docs_total gauge
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index=".monitoring-data-2"} 2
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 65
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index=".watches"} 4
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_segment_count_primary Current number of segments with only primary shards on all nodes
             # TYPE elasticsearch_indices_segment_count_primary gauge
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 1
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 3
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index=".watches"} 4
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_segment_count_total Current number of segments with all shards on all nodes
             # TYPE elasticsearch_indices_segment_count_total gauge
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index=".monitoring-data-2"} 1
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 3
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index=".watches"} 4
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index="foo_2"} 3
             # HELP elasticsearch_indices_segment_doc_values_memory_bytes_primary Current size of doc values with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_doc_values_memory_bytes_primary gauge
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 236
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 3452
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index=".watches"} 368
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 184
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 276
             # HELP elasticsearch_indices_segment_doc_values_memory_bytes_total Current size of doc values with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_doc_values_memory_bytes_total gauge
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 236
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 3452
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 368
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 184
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 276
             # HELP elasticsearch_indices_segment_fields_memory_bytes_primary Current size of fields with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fields_memory_bytes_primary gauge
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 312
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 936
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index=".watches"} 1248
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 624
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 936
             # HELP elasticsearch_indices_segment_fields_memory_bytes_total Current size of fields with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fields_memory_bytes_total gauge
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 312
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 936
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 1248
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 624
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 936
             # HELP elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary Current size of fixed bit with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary gauge
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total Current size of fixed bit with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total gauge
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_index_writer_memory_bytes_primary Current size of index writer with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_index_writer_memory_bytes_primary gauge
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_index_writer_memory_bytes_total Current size of index writer with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_index_writer_memory_bytes_total gauge
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_memory_bytes_primary Current size of segments with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_memory_bytes_primary gauge
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 1335
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 23830
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index=".watches"} 18418
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 5176
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 7764
             # HELP elasticsearch_indices_segment_memory_bytes_total Current size of segments with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_memory_bytes_total gauge
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 1335
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 23830
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 18418
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 5176
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 7764
             # HELP elasticsearch_indices_segment_norms_memory_bytes_primary Current size of norms with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_norms_memory_bytes_primary gauge
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 320
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index=".watches"} 1600
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 384
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 576
             # HELP elasticsearch_indices_segment_norms_memory_bytes_total Current size of norms with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_norms_memory_bytes_total gauge
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 320
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 1600
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 384
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 576
             # HELP elasticsearch_indices_segment_points_memory_bytes_primary Current size of points with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_points_memory_bytes_primary gauge
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 648
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index=".watches"} 4
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_points_memory_bytes_total Current size of points with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_points_memory_bytes_total gauge
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 648
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 4
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_term_vectors_memory_primary_bytes Current size of term vectors with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_term_vectors_memory_primary_bytes gauge
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_term_vectors_memory_total_bytes Current size of term vectors with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_term_vectors_memory_total_bytes gauge
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_terms_memory_primary Current size of terms with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_terms_memory_primary gauge
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 787
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 18474
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index=".watches"} 15198
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index="foo_1"} 3984
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index="foo_2"} 5976
             # HELP elasticsearch_indices_segment_terms_memory_total Current number of terms with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_terms_memory_total gauge
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index=".monitoring-data-2"} 787
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 18474
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index=".watches"} 15198
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index="foo_1"} 3984
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index="foo_2"} 5976
             # HELP elasticsearch_indices_segment_version_map_memory_bytes_primary Current size of version map with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_version_map_memory_bytes_primary gauge
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_segment_version_map_memory_bytes_total Current size of version map with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_version_map_memory_bytes_total gauge
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             # HELP elasticsearch_indices_store_size_bytes_primary Current total size of stored index data in bytes with only primary shards on all nodes
             # TYPE elasticsearch_indices_store_size_bytes_primary gauge
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index=".monitoring-data-2"} 4226
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 68917
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index=".watches"} 35444
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index="foo_1"} 8038
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index="foo_2"} 11909
             # HELP elasticsearch_indices_store_size_bytes_total Current total size of stored index data in bytes with all shards on all nodes
             # TYPE elasticsearch_indices_store_size_bytes_total gauge
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index=".monitoring-data-2"} 4226
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 68917
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index=".watches"} 35444
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index="foo_1"} 8038
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index="foo_2"} 11909
             # HELP elasticsearch_search_active_queries The number of currently active queries
             # TYPE elasticsearch_search_active_queries gauge
             elasticsearch_search_active_queries{cluster="unknown_cluster",index=".monitoring-data-2"} 0
             elasticsearch_search_active_queries{cluster="unknown_cluster",index=".monitoring-es-2-2017.08.23"} 0
             elasticsearch_search_active_queries{cluster="unknown_cluster",index=".watches"} 0
             elasticsearch_search_active_queries{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_search_active_queries{cluster="unknown_cluster",index="foo_2"} 0

`,
		},
		{
			name: "7.17.3",
			file: "7.17.3.json",
			want: `# HELP elasticsearch_index_stats_fielddata_evictions_total Total fielddata evictions count
             # TYPE elasticsearch_index_stats_fielddata_evictions_total counter
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_fielddata_evictions_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_fielddata_memory_bytes_total Total fielddata memory bytes
             # TYPE elasticsearch_index_stats_fielddata_memory_bytes_total counter
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_fielddata_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_flush_time_seconds_total Total flush time in seconds
             # TYPE elasticsearch_index_stats_flush_time_seconds_total counter
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0.15
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_flush_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_flush_total Total flush count
             # TYPE elasticsearch_index_stats_flush_total counter
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index=".geoip_databases"} 4
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_flush_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_get_time_seconds_total Total get time in seconds
             # TYPE elasticsearch_index_stats_get_time_seconds_total counter
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_get_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_get_total Total get count
             # TYPE elasticsearch_index_stats_get_total counter
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_get_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_index_current The number of documents currently being indexed to an index
             # TYPE elasticsearch_index_stats_index_current gauge
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_index_current{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_indexing_delete_time_seconds_total Total indexing delete time in seconds
             # TYPE elasticsearch_index_stats_indexing_delete_time_seconds_total counter
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_indexing_delete_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_indexing_delete_total Total indexing delete count
             # TYPE elasticsearch_index_stats_indexing_delete_total counter
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_indexing_delete_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_indexing_index_time_seconds_total Total indexing index time in seconds
             # TYPE elasticsearch_index_stats_indexing_index_time_seconds_total counter
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0.738
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.001
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0.001
             elasticsearch_index_stats_indexing_index_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0.001
             # HELP elasticsearch_index_stats_indexing_index_total Total indexing index count
             # TYPE elasticsearch_index_stats_indexing_index_total counter
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index=".geoip_databases"} 40
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index="foo_1"} 1
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index="foo_2"} 1
             elasticsearch_index_stats_indexing_index_total{cluster="unknown_cluster",index="foo_3"} 1
             # HELP elasticsearch_index_stats_indexing_noop_update_total Total indexing no-op update count
             # TYPE elasticsearch_index_stats_indexing_noop_update_total counter
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_indexing_noop_update_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_indexing_throttle_time_seconds_total Total indexing throttle time in seconds
             # TYPE elasticsearch_index_stats_indexing_throttle_time_seconds_total counter
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_indexing_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_merge_auto_throttle_bytes_total Total bytes that were auto-throttled during merging
             # TYPE elasticsearch_index_stats_merge_auto_throttle_bytes_total counter
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 2.097152e+07
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index="foo_1"} 2.097152e+07
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index="foo_2"} 2.097152e+07
             elasticsearch_index_stats_merge_auto_throttle_bytes_total{cluster="unknown_cluster",index="foo_3"} 2.097152e+07
             # HELP elasticsearch_index_stats_merge_stopped_time_seconds_total Total large merge stopped time in seconds, allowing smaller merges to complete
             # TYPE elasticsearch_index_stats_merge_stopped_time_seconds_total counter
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_merge_stopped_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_merge_throttle_time_seconds_total Total merge I/O throttle time in seconds
             # TYPE elasticsearch_index_stats_merge_throttle_time_seconds_total counter
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_merge_throttle_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_merge_time_seconds_total Total merge time in seconds
             # TYPE elasticsearch_index_stats_merge_time_seconds_total counter
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_merge_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_merge_total Total merge count
             # TYPE elasticsearch_index_stats_merge_total counter
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_merge_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_query_cache_caches_total Total query cache caches count
             # TYPE elasticsearch_index_stats_query_cache_caches_total counter
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_query_cache_caches_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_query_cache_evictions_total Total query cache evictions count
             # TYPE elasticsearch_index_stats_query_cache_evictions_total counter
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_query_cache_evictions_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_query_cache_hits_total Total query cache hits count
             # TYPE elasticsearch_index_stats_query_cache_hits_total counter
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_query_cache_hits_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_query_cache_memory_bytes_total Total query cache memory bytes
             # TYPE elasticsearch_index_stats_query_cache_memory_bytes_total counter
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_query_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_query_cache_misses_total Total query cache misses count
             # TYPE elasticsearch_index_stats_query_cache_misses_total counter
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_query_cache_misses_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_query_cache_size Total query cache size
             # TYPE elasticsearch_index_stats_query_cache_size gauge
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_query_cache_size{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_refresh_external_time_seconds_total Total external refresh time in seconds
             # TYPE elasticsearch_index_stats_refresh_external_time_seconds_total counter
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0.045
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.008
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0.01
             elasticsearch_index_stats_refresh_external_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0.01
             # HELP elasticsearch_index_stats_refresh_external_total Total external refresh count
             # TYPE elasticsearch_index_stats_refresh_external_total counter
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index=".geoip_databases"} 6
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index="foo_1"} 3
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index="foo_2"} 3
             elasticsearch_index_stats_refresh_external_total{cluster="unknown_cluster",index="foo_3"} 3
             # HELP elasticsearch_index_stats_refresh_time_seconds_total Total refresh time in seconds
             # TYPE elasticsearch_index_stats_refresh_time_seconds_total counter
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0.05
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0.008
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0.009
             elasticsearch_index_stats_refresh_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0.009
             # HELP elasticsearch_index_stats_refresh_total Total refresh count
             # TYPE elasticsearch_index_stats_refresh_total counter
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index=".geoip_databases"} 9
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index="foo_1"} 3
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index="foo_2"} 3
             elasticsearch_index_stats_refresh_total{cluster="unknown_cluster",index="foo_3"} 3
             # HELP elasticsearch_index_stats_request_cache_evictions_total Total request cache evictions count
             # TYPE elasticsearch_index_stats_request_cache_evictions_total counter
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_request_cache_evictions_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_request_cache_hits_total Total request cache hits count
             # TYPE elasticsearch_index_stats_request_cache_hits_total counter
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_request_cache_hits_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_request_cache_memory_bytes_total Total request cache memory bytes
             # TYPE elasticsearch_index_stats_request_cache_memory_bytes_total counter
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_request_cache_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_request_cache_misses_total Total request cache misses count
             # TYPE elasticsearch_index_stats_request_cache_misses_total counter
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_request_cache_misses_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_search_fetch_time_seconds_total Total search fetch time in seconds
             # TYPE elasticsearch_index_stats_search_fetch_time_seconds_total counter
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0.096
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_search_fetch_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_search_fetch_total Total search fetch count
             # TYPE elasticsearch_index_stats_search_fetch_total counter
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index=".geoip_databases"} 43
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_search_fetch_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_search_query_time_seconds_total Total search query time in seconds
             # TYPE elasticsearch_index_stats_search_query_time_seconds_total counter
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0.071
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_search_query_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_search_query_total Total number of queries
             # TYPE elasticsearch_index_stats_search_query_total counter
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index=".geoip_databases"} 43
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_search_query_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_search_scroll_current Current search scroll count
             # TYPE elasticsearch_index_stats_search_scroll_current gauge
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_search_scroll_current{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_search_scroll_time_seconds_total Total search scroll time in seconds
             # TYPE elasticsearch_index_stats_search_scroll_time_seconds_total counter
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0.06
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_search_scroll_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_search_scroll_total Total search scroll count
             # TYPE elasticsearch_index_stats_search_scroll_total counter
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index=".geoip_databases"} 3
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_search_scroll_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_search_suggest_time_seconds_total Total search suggest time in seconds
             # TYPE elasticsearch_index_stats_search_suggest_time_seconds_total counter
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_search_suggest_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_search_suggest_total Total search suggest count
             # TYPE elasticsearch_index_stats_search_suggest_total counter
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_search_suggest_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_warmer_time_seconds_total Total warmer time in seconds
             # TYPE elasticsearch_index_stats_warmer_time_seconds_total counter
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_index_stats_warmer_time_seconds_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_index_stats_warmer_total Total warmer count
             # TYPE elasticsearch_index_stats_warmer_total counter
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index=".geoip_databases"} 5
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index="foo_1"} 2
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index="foo_2"} 2
             elasticsearch_index_stats_warmer_total{cluster="unknown_cluster",index="foo_3"} 2
             # HELP elasticsearch_indices_aliases Record aliases associated with an index
             # TYPE elasticsearch_indices_aliases gauge
             elasticsearch_indices_aliases{alias="foo_alias_2_1",cluster="unknown_cluster",index="foo_2"} 1
             elasticsearch_indices_aliases{alias="foo_alias_3_1",cluster="unknown_cluster",index="foo_3"} 1
             elasticsearch_indices_aliases{alias="foo_alias_3_2",cluster="unknown_cluster",index="foo_3"} 1
             # HELP elasticsearch_indices_completion_bytes_primary Current size of completion with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_completion_bytes_primary gauge
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_completion_bytes_primary{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_completion_bytes_total Current size of completion with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_completion_bytes_total gauge
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_completion_bytes_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_deleted_docs_primary Count of deleted documents with only primary shards
             # TYPE elasticsearch_indices_deleted_docs_primary gauge
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_deleted_docs_primary{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_deleted_docs_total Total count of deleted documents
             # TYPE elasticsearch_indices_deleted_docs_total gauge
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_deleted_docs_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_docs_primary Count of documents with only primary shards
             # TYPE elasticsearch_indices_docs_primary gauge
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index=".geoip_databases"} 40
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index="foo_1"} 1
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index="foo_2"} 1
             elasticsearch_indices_docs_primary{cluster="unknown_cluster",index="foo_3"} 1
             # HELP elasticsearch_indices_docs_total Total count of documents
             # TYPE elasticsearch_indices_docs_total gauge
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index=".geoip_databases"} 40
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index="foo_1"} 1
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index="foo_2"} 1
             elasticsearch_indices_docs_total{cluster="unknown_cluster",index="foo_3"} 1
             # HELP elasticsearch_indices_segment_count_primary Current number of segments with only primary shards on all nodes
             # TYPE elasticsearch_indices_segment_count_primary gauge
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index=".geoip_databases"} 4
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index="foo_1"} 1
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index="foo_2"} 1
             elasticsearch_indices_segment_count_primary{cluster="unknown_cluster",index="foo_3"} 1
             # HELP elasticsearch_indices_segment_count_total Current number of segments with all shards on all nodes
             # TYPE elasticsearch_indices_segment_count_total gauge
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index=".geoip_databases"} 4
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index="foo_1"} 1
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index="foo_2"} 1
             elasticsearch_indices_segment_count_total{cluster="unknown_cluster",index="foo_3"} 1
             # HELP elasticsearch_indices_segment_doc_values_memory_bytes_primary Current size of doc values with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_doc_values_memory_bytes_primary gauge
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index=".geoip_databases"} 304
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 76
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 76
             elasticsearch_indices_segment_doc_values_memory_bytes_primary{cluster="unknown_cluster",index="foo_3"} 76
             # HELP elasticsearch_indices_segment_doc_values_memory_bytes_total Current size of doc values with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_doc_values_memory_bytes_total gauge
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 304
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 76
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 76
             elasticsearch_indices_segment_doc_values_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 76
             # HELP elasticsearch_indices_segment_fields_memory_bytes_primary Current size of fields with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fields_memory_bytes_primary gauge
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index=".geoip_databases"} 2016
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 488
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 488
             elasticsearch_indices_segment_fields_memory_bytes_primary{cluster="unknown_cluster",index="foo_3"} 488
             # HELP elasticsearch_indices_segment_fields_memory_bytes_total Current size of fields with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fields_memory_bytes_total gauge
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 2016
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 488
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 488
             elasticsearch_indices_segment_fields_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 488
             # HELP elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary Current size of fixed bit with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary gauge
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_primary{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total Current size of fixed bit with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total gauge
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_segment_fixed_bit_set_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_segment_index_writer_memory_bytes_primary Current size of index writer with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_index_writer_memory_bytes_primary gauge
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_primary{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_segment_index_writer_memory_bytes_total Current size of index writer with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_index_writer_memory_bytes_total gauge
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_segment_index_writer_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_segment_memory_bytes_primary Current size of segments with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_memory_bytes_primary gauge
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index=".geoip_databases"} 4368
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 1876
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 1876
             elasticsearch_indices_segment_memory_bytes_primary{cluster="unknown_cluster",index="foo_3"} 1876
             # HELP elasticsearch_indices_segment_memory_bytes_total Current size of segments with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_memory_bytes_total gauge
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 4368
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 1876
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 1876
             elasticsearch_indices_segment_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 1876
             # HELP elasticsearch_indices_segment_norms_memory_bytes_primary Current size of norms with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_norms_memory_bytes_primary gauge
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 128
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 128
             elasticsearch_indices_segment_norms_memory_bytes_primary{cluster="unknown_cluster",index="foo_3"} 128
             # HELP elasticsearch_indices_segment_norms_memory_bytes_total Current size of norms with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_norms_memory_bytes_total gauge
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 128
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 128
             elasticsearch_indices_segment_norms_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 128
             # HELP elasticsearch_indices_segment_points_memory_bytes_primary Current size of points with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_points_memory_bytes_primary gauge
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_segment_points_memory_bytes_primary{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_segment_points_memory_bytes_total Current size of points with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_points_memory_bytes_total gauge
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_segment_points_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_segment_term_vectors_memory_primary_bytes Current size of term vectors with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_term_vectors_memory_primary_bytes gauge
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_segment_term_vectors_memory_primary_bytes{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_segment_term_vectors_memory_total_bytes Current size of term vectors with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_term_vectors_memory_total_bytes gauge
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_segment_term_vectors_memory_total_bytes{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_segment_terms_memory_primary Current size of terms with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_terms_memory_primary gauge
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index=".geoip_databases"} 2048
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index="foo_1"} 1184
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index="foo_2"} 1184
             elasticsearch_indices_segment_terms_memory_primary{cluster="unknown_cluster",index="foo_3"} 1184
             # HELP elasticsearch_indices_segment_terms_memory_total Current number of terms with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_terms_memory_total gauge
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index=".geoip_databases"} 2048
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index="foo_1"} 1184
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index="foo_2"} 1184
             elasticsearch_indices_segment_terms_memory_total{cluster="unknown_cluster",index="foo_3"} 1184
             # HELP elasticsearch_indices_segment_version_map_memory_bytes_primary Current size of version map with only primary shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_version_map_memory_bytes_primary gauge
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_primary{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_segment_version_map_memory_bytes_total Current size of version map with all shards on all nodes in bytes
             # TYPE elasticsearch_indices_segment_version_map_memory_bytes_total gauge
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_indices_segment_version_map_memory_bytes_total{cluster="unknown_cluster",index="foo_3"} 0
             # HELP elasticsearch_indices_store_size_bytes_primary Current total size of stored index data in bytes with only primary shards on all nodes
             # TYPE elasticsearch_indices_store_size_bytes_primary gauge
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index=".geoip_databases"} 3.9904033e+07
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index="foo_1"} 4413
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index="foo_2"} 4459
             elasticsearch_indices_store_size_bytes_primary{cluster="unknown_cluster",index="foo_3"} 4459
             # HELP elasticsearch_indices_store_size_bytes_total Current total size of stored index data in bytes with all shards on all nodes
             # TYPE elasticsearch_indices_store_size_bytes_total gauge
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index=".geoip_databases"} 3.9904033e+07
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index="foo_1"} 4413
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index="foo_2"} 4459
             elasticsearch_indices_store_size_bytes_total{cluster="unknown_cluster",index="foo_3"} 4459
             # HELP elasticsearch_search_active_queries The number of currently active queries
             # TYPE elasticsearch_search_active_queries gauge
             elasticsearch_search_active_queries{cluster="unknown_cluster",index=".geoip_databases"} 0
             elasticsearch_search_active_queries{cluster="unknown_cluster",index="foo_1"} 0
             elasticsearch_search_active_queries{cluster="unknown_cluster",index="foo_2"} 0
             elasticsearch_search_active_queries{cluster="unknown_cluster",index="foo_3"} 0
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fStats, err := os.Open(path.Join("../fixtures/indices/", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			defer fStats.Close()
			fAlias, err := os.Open(path.Join("../fixtures/indices/alias/", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			defer fAlias.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/_all/_stats":
					io.Copy(w, fStats)
				case "/_alias":
					io.Copy(w, fAlias)
				default:
					http.Error(w, "Not Found", http.StatusNotFound)
				}

			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			c := NewIndices(promslog.NewNopLogger(), http.DefaultClient, u, false, true)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(c, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
