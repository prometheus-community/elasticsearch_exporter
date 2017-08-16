package collector

// indexStatsResponse is a representation of a Elasticsearch Index Stats
type indexStatsResponse struct {
	Shards  IndexStatsShardsResponse `json:"_shards"`
	All     IndexStatsIndexResponse `json:"_all"`
	Indices map[string]IndexStatsIndexResponse `json:"indices"`
}

type IndexStatsShardsResponse struct {
	Total      int64  `json:"total"`
	Successful int64  `json:"successful"`
	Failed     int64  `json:"failed"`
}

type IndexStatsIndexResponse struct {
	Primaries IndexStatsIndexDetailResponse `json:"primaries"`
	Total     IndexStatsIndexDetailResponse `json:"total"`
}

type IndexStatsIndexDetailResponse struct {
	Docs         IndexStatsIndexDocsResponse  `json:"docs"`
	Store        IndexStatsIndexStoreResponse  `json:"store"`
	Indexing     IndexStatsIndexIndexingResponse  `json:"indexing"`
	Get          IndexStatsIndexGetResponse  `json:"get"`
	Search       IndexStatsIndexSearchResponse  `json:"search"`
	Merges       IndexStatsIndexMergesResponse  `json:"merges"`
	Refresh      IndexStatsIndexRefreshResponse  `json:"refresh"`
	Flush        IndexStatsIndexFlushResponse  `json:"flush"`
	Warmer       IndexStatsIndexWarmerResponse  `json:"warmer"`
	QueryCache   IndexStatsIndexQueryCacheResponse  `json:"query_cache"`
	Fielddata    IndexStatsIndexFielddataResponse  `json:"fielddata"`
	Completion   IndexStatsIndexCompletionResponse  `json:"completion"`
	Segments     IndexStatsIndexSegmentsResponse  `json:"segments"`
	Translog     IndexStatsIndexTranslogResponse  `json:"translog"`
	RequestCache IndexStatsIndexRequestCacheResponse  `json:"request_cache"`
	Recovery     IndexStatsIndexRecoveryResponse  `json:"recovery"`
}

type IndexStatsIndexDocsResponse struct {
	Count   int64  `json:"count"`
	Deleted int64  `json:"deleted"`
}

type IndexStatsIndexStoreResponse struct {
	SizeInBytes          int64  `json:"size_in_bytes"`
	ThrottleTimeInMillis int64  `json:"throttle_time_in_millis"`
}

type IndexStatsIndexIndexingResponse struct {
	IndexTotal           int64  `json:"index_total"`
	IndexTimeInMillis    int64  `json:"index_time_in_millis"`
	IndexCurrent         int64  `json:"index_current"`
	IndexFailed          int64  `json:"index_failed"`
	DeleteTotal          int64  `json:"delete_total"`
	DeleteTimeInMillis   int64  `json:"delete_time_in_millis"`
	DeleteCurrent        int64  `json:"delete_current"`
	NoopUpdateTotal      int64  `json:"noop_update_total"`
	IsThrottled          bool   `json:"is_throttled"`
	ThrottleTimeInMillis int64  `json:"throttle_time_in_millis"`
}

type IndexStatsIndexGetResponse struct {
	Total               int64  `json:"total"`
	TimeInMillis        int64  `json:"time_in_millis"`
	ExistsTotal         int64  `json:"exists_total"`
	ExistsTimeInMillis  int64  `json:"exists_time_in_millis"`
	MissingTotal        int64  `json:"missing_total"`
	MissingTimeInMillis int64  `json:"missing_time_in_millis"`
	Current             int64  `json:"current"`
}

type IndexStatsIndexSearchResponse struct {
	OpenContexts        int64  `json:"open_contexts"`
	QueryTotal          int64  `json:"query_total"`
	QueryTimeInMillis   int64  `json:"query_time_in_millis"`
	QueryCurrent        int64  `json:"query_current"`
	FetchTotal          int64  `json:"fetch_total"`
	FetchTimeInMillis   int64  `json:"fetch_time_in_millis"`
	FetchCurrent        int64  `json:"fetch_current"`
	ScrollTotal         int64  `json:"scroll_total"`
	ScrollTimeInMillis  int64  `json:"scroll_time_in_millis"`
	ScrollCurrent       int64  `json:"scroll_current"`
	SuggestTotal        int64  `json:"suggest_total"`
	SuggestTimeInMillis int64  `json:"suggest_time_in_millis"`
	SuggestCurrent      int64  `json:"suggest_current"`
}

type IndexStatsIndexMergesResponse struct {
	Current                    int64  `json:"current"`
	CurrentDocs                int64  `json:"current_docs"`
	CurrentSizeInBytes         int64  `json:"current_size_in_bytes"`
	Total                      int64  `json:"total"`
	TotalTimeInMillis          int64  `json:"total_time_in_millis"`
	TotalDocs                  int64  `json:"total_docs"`
	TotalSizeInBytes           int64  `json:"total_size_in_bytes"`
	TotalStoppedTimeInMillis   int64  `json:"total_stopped_time_in_millis"`
	TotalThrottledTimeInMillis int64  `json:"total_throttled_time_in_millis"`
	TotalAutoThrottleInBytes   int64  `json:"total_auto_throttle_in_bytes"`
}

type IndexStatsIndexRefreshResponse struct {
	Total             int64  `json:"total"`
	TotalTimeInMillis int64  `json:"total_time_in_millis"`
	Listeners         int64  `json:"listeners"`
}

type IndexStatsIndexFlushResponse struct {
	Total             int64  `json:"total"`
	TotalTimeInMillis int64  `json:"total_time_in_millis"`
}

type IndexStatsIndexWarmerResponse struct {
	Current           int64  `json:"current"`
	Total             int64  `json:"total"`
	TotalTimeInMillis int64  `json:"total_time_in_millis"`
}

type IndexStatsIndexQueryCacheResponse struct {
	MemorySizeInBytes int64  `json:"memory_size_in_bytes"`
	TotalCount        int64  `json:"total_count"`
	HitCount          int64  `json:"hit_count"`
	MissCount         int64  `json:"miss_count"`
	CacheSize         int64  `json:"cache_size"`
	CacheCount        int64  `json:"cache_count"`
	Evictions         int64  `json:"evictions"`
}

type IndexStatsIndexFielddataResponse struct {
	MemorySizeInBytes int64  `json:"memory_size_in_bytes"`
	Evictions         int64  `json:"evictions"`
}

type IndexStatsIndexCompletionResponse struct {
	SizeInBytes int64  `json:"size_in_bytes"`
}

type IndexStatsIndexSegmentsResponse struct {
	Count                     int64  `json:"count"`
	MemoryInBytes             int64  `json:"memory_in_bytes"`
	TermsMemoryInBytes        int64  `json:"terms_memory_in_bytes"`
	StoredFieldsMemoryInBytes int64  `json:"stored_fields_memory_in_bytes"`
	TermVectorsMemoryInBytes  int64  `json:"term_vectors_memory_in_bytes"`
	NormsMemoryInBytes        int64  `json:"norms_memory_in_bytes"`
	PointsMemoryInBytes       int64  `json:"points_memory_in_bytes"`
	DocValuesMemoryInBytes    int64  `json:"doc_values_memory_in_bytes"`
	IndexWriterMemoryInBytes  int64  `json:"index_writer_memory_in_bytes"`
	VersionMapMemoryInBytes   int64  `json:"version_map_memory_in_bytes"`
	FixedBitSetMemoryInBytes  int64  `json:"fixed_bit_set_memory_in_bytes"`
	MaxUnsafeAutoIdTimestamp  int64  `json:"max_unsafe_auto_id_timestamp"`
}

type IndexStatsIndexTranslogResponse struct {
	Operations  int64  `json:"operations"`
	SizeInBytes int64  `json:"size_in_bytes"`
}

type IndexStatsIndexRequestCacheResponse struct {
	MemorySizeInBytes int64  `json:"memory_size_in_bytes"`
	Evictions         int64  `json:"evictions"`
	HitCount          int64  `json:"hit_count"`
	MissCount         int64  `json:"miss_count"`
}

type IndexStatsIndexRecoveryResponse struct {
	CurrentAsSource      int64  `json:"current_as_source"`
	CurrentAsTarget      int64  `json:"current_as_target"`
	ThrottleTimeInMillis int64  `json:"throttle_time_in_millis"`
}
