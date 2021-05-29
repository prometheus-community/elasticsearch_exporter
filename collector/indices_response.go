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

// indexStatsResponse is a representation of a Elasticsearch Index Stats
type indexStatsResponse struct {
	Shards  IndexStatsShardsResponse           `json:"_shards"`
	All     IndexStatsIndexResponse            `json:"_all"`
	Indices map[string]IndexStatsIndexResponse `json:"indices"`
}

// IndexStatsShardsResponse defines index stats shards information structure
type IndexStatsShardsResponse struct {
	Total      int64 `json:"total"`
	Successful int64 `json:"successful"`
	Failed     int64 `json:"failed"`
}

// IndexStatsIndexResponse defines index stats index information structure
type IndexStatsIndexResponse struct {
	Primaries IndexStatsIndexDetailResponse                    `json:"primaries"`
	Total     IndexStatsIndexDetailResponse                    `json:"total"`
	Shards    map[string][]IndexStatsIndexShardsDetailResponse `json:"shards"`
}

// IndexStatsIndexDetailResponse defines index stats index details information structure
type IndexStatsIndexDetailResponse struct {
	Docs         IndexStatsIndexDocsResponse         `json:"docs"`
	Store        IndexStatsIndexStoreResponse        `json:"store"`
	Indexing     IndexStatsIndexIndexingResponse     `json:"indexing"`
	Get          IndexStatsIndexGetResponse          `json:"get"`
	Search       IndexStatsIndexSearchResponse       `json:"search"`
	Merges       IndexStatsIndexMergesResponse       `json:"merges"`
	Refresh      IndexStatsIndexRefreshResponse      `json:"refresh"`
	Flush        IndexStatsIndexFlushResponse        `json:"flush"`
	Warmer       IndexStatsIndexWarmerResponse       `json:"warmer"`
	QueryCache   IndexStatsIndexQueryCacheResponse   `json:"query_cache"`
	Fielddata    IndexStatsIndexFielddataResponse    `json:"fielddata"`
	Completion   IndexStatsIndexCompletionResponse   `json:"completion"`
	Segments     IndexStatsIndexSegmentsResponse     `json:"segments"`
	Translog     IndexStatsIndexTranslogResponse     `json:"translog"`
	RequestCache IndexStatsIndexRequestCacheResponse `json:"request_cache"`
	Recovery     IndexStatsIndexRecoveryResponse     `json:"recovery"`
}

// IndexStatsIndexShardsDetailResponse defines index stats index shard details information structure
type IndexStatsIndexShardsDetailResponse struct {
	*IndexStatsIndexDetailResponse
	Routing IndexStatsIndexRoutingResponse `json:"routing"`
}

// IndexStatsIndexRoutingResponse defines index stats index routing information structure
type IndexStatsIndexRoutingResponse struct {
	Node    string `json:"node"`
	Primary bool   `json:"primary"`
}

// IndexStatsIndexDocsResponse defines index stats index documents information structure
type IndexStatsIndexDocsResponse struct {
	Count   int64 `json:"count"`
	Deleted int64 `json:"deleted"`
}

// IndexStatsIndexStoreResponse defines index stats index store information structure
type IndexStatsIndexStoreResponse struct {
	SizeInBytes          int64 `json:"size_in_bytes"`
	ThrottleTimeInMillis int64 `json:"throttle_time_in_millis"`
}

// IndexStatsIndexIndexingResponse defines index stats index indexing information structure
type IndexStatsIndexIndexingResponse struct {
	IndexTotal           int64 `json:"index_total"`
	IndexTimeInMillis    int64 `json:"index_time_in_millis"`
	IndexCurrent         int64 `json:"index_current"`
	IndexFailed          int64 `json:"index_failed"`
	DeleteTotal          int64 `json:"delete_total"`
	DeleteTimeInMillis   int64 `json:"delete_time_in_millis"`
	DeleteCurrent        int64 `json:"delete_current"`
	NoopUpdateTotal      int64 `json:"noop_update_total"`
	IsThrottled          bool  `json:"is_throttled"`
	ThrottleTimeInMillis int64 `json:"throttle_time_in_millis"`
}

// IndexStatsIndexGetResponse defines index stats index get information structure
type IndexStatsIndexGetResponse struct {
	Total               int64 `json:"total"`
	TimeInMillis        int64 `json:"time_in_millis"`
	ExistsTotal         int64 `json:"exists_total"`
	ExistsTimeInMillis  int64 `json:"exists_time_in_millis"`
	MissingTotal        int64 `json:"missing_total"`
	MissingTimeInMillis int64 `json:"missing_time_in_millis"`
	Current             int64 `json:"current"`
}

// IndexStatsIndexSearchResponse defines index stats index search information structure
type IndexStatsIndexSearchResponse struct {
	OpenContexts        int64 `json:"open_contexts"`
	QueryTotal          int64 `json:"query_total"`
	QueryTimeInMillis   int64 `json:"query_time_in_millis"`
	QueryCurrent        int64 `json:"query_current"`
	FetchTotal          int64 `json:"fetch_total"`
	FetchTimeInMillis   int64 `json:"fetch_time_in_millis"`
	FetchCurrent        int64 `json:"fetch_current"`
	ScrollTotal         int64 `json:"scroll_total"`
	ScrollTimeInMillis  int64 `json:"scroll_time_in_millis"`
	ScrollCurrent       int64 `json:"scroll_current"`
	SuggestTotal        int64 `json:"suggest_total"`
	SuggestTimeInMillis int64 `json:"suggest_time_in_millis"`
	SuggestCurrent      int64 `json:"suggest_current"`
}

// IndexStatsIndexMergesResponse defines index stats index merges information structure
type IndexStatsIndexMergesResponse struct {
	Current                    int64 `json:"current"`
	CurrentDocs                int64 `json:"current_docs"`
	CurrentSizeInBytes         int64 `json:"current_size_in_bytes"`
	Total                      int64 `json:"total"`
	TotalTimeInMillis          int64 `json:"total_time_in_millis"`
	TotalDocs                  int64 `json:"total_docs"`
	TotalSizeInBytes           int64 `json:"total_size_in_bytes"`
	TotalStoppedTimeInMillis   int64 `json:"total_stopped_time_in_millis"`
	TotalThrottledTimeInMillis int64 `json:"total_throttled_time_in_millis"`
	TotalAutoThrottleInBytes   int64 `json:"total_auto_throttle_in_bytes"`
}

// IndexStatsIndexRefreshResponse defines index stats index refresh information structure
type IndexStatsIndexRefreshResponse struct {
	Total             int64 `json:"total"`
	TotalTimeInMillis int64 `json:"total_time_in_millis"`
	Listeners         int64 `json:"listeners"`
}

// IndexStatsIndexFlushResponse defines index stats index flush information structure
type IndexStatsIndexFlushResponse struct {
	Total             int64 `json:"total"`
	TotalTimeInMillis int64 `json:"total_time_in_millis"`
}

// IndexStatsIndexWarmerResponse defines index stats index warmer information structure
type IndexStatsIndexWarmerResponse struct {
	Current           int64 `json:"current"`
	Total             int64 `json:"total"`
	TotalTimeInMillis int64 `json:"total_time_in_millis"`
}

// IndexStatsIndexQueryCacheResponse defines index stats index query cache information structure
type IndexStatsIndexQueryCacheResponse struct {
	MemorySizeInBytes int64 `json:"memory_size_in_bytes"`
	TotalCount        int64 `json:"total_count"`
	HitCount          int64 `json:"hit_count"`
	MissCount         int64 `json:"miss_count"`
	CacheSize         int64 `json:"cache_size"`
	CacheCount        int64 `json:"cache_count"`
	Evictions         int64 `json:"evictions"`
}

// IndexStatsIndexFielddataResponse defines index stats index fielddata information structure
type IndexStatsIndexFielddataResponse struct {
	MemorySizeInBytes int64 `json:"memory_size_in_bytes"`
	Evictions         int64 `json:"evictions"`
}

// IndexStatsIndexCompletionResponse defines index stats index completion information structure
type IndexStatsIndexCompletionResponse struct {
	SizeInBytes int64 `json:"size_in_bytes"`
}

// IndexStatsIndexSegmentsResponse defines index stats index segments information structure
type IndexStatsIndexSegmentsResponse struct {
	Count                     int64 `json:"count"`
	MemoryInBytes             int64 `json:"memory_in_bytes"`
	TermsMemoryInBytes        int64 `json:"terms_memory_in_bytes"`
	StoredFieldsMemoryInBytes int64 `json:"stored_fields_memory_in_bytes"`
	TermVectorsMemoryInBytes  int64 `json:"term_vectors_memory_in_bytes"`
	NormsMemoryInBytes        int64 `json:"norms_memory_in_bytes"`
	PointsMemoryInBytes       int64 `json:"points_memory_in_bytes"`
	DocValuesMemoryInBytes    int64 `json:"doc_values_memory_in_bytes"`
	IndexWriterMemoryInBytes  int64 `json:"index_writer_memory_in_bytes"`
	VersionMapMemoryInBytes   int64 `json:"version_map_memory_in_bytes"`
	FixedBitSetMemoryInBytes  int64 `json:"fixed_bit_set_memory_in_bytes"`
	MaxUnsafeAutoIDTimestamp  int64 `json:"max_unsafe_auto_id_timestamp"`
}

// IndexStatsIndexTranslogResponse defines index stats index translog information structure
type IndexStatsIndexTranslogResponse struct {
	Operations  int64 `json:"operations"`
	SizeInBytes int64 `json:"size_in_bytes"`
}

// IndexStatsIndexRequestCacheResponse defines index stats index request cache information structure
type IndexStatsIndexRequestCacheResponse struct {
	MemorySizeInBytes int64 `json:"memory_size_in_bytes"`
	Evictions         int64 `json:"evictions"`
	HitCount          int64 `json:"hit_count"`
	MissCount         int64 `json:"miss_count"`
}

// IndexStatsIndexRecoveryResponse defines index stats index recovery information structure
type IndexStatsIndexRecoveryResponse struct {
	CurrentAsSource      int64 `json:"current_as_source"`
	CurrentAsTarget      int64 `json:"current_as_target"`
	ThrottleTimeInMillis int64 `json:"throttle_time_in_millis"`
}
