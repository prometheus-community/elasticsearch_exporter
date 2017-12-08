package collector

import "encoding/json"

// nodeStatsResponse is a representation of a Elasticsearch Node Stats
type nodeStatsResponse struct {
	ClusterName string `json:"cluster_name"`
	Nodes       map[string]NodeStatsNodeResponse
}

type NodeStatsNodeResponse struct {
	Name             string                                     `json:"name"`
	Host             string                                     `json:"host"`
	Timestamp        int64                                      `json:"timestamp"`
	TransportAddress string                                     `json:"transport_address"`
	Hostname         string                                     `json:"hostname"`
	Roles            []string                                   `json:"roles"`
	Attributes       map[string]string                          `json:"attributes"`
	Indices          NodeStatsIndicesResponse                   `json:"indices"`
	OS               NodeStatsOSResponse                        `json:"os"`
	Network          NodeStatsNetworkResponse                   `json:"network"`
	FS               NodeStatsFSResponse                        `json:"fs"`
	ThreadPool       map[string]NodeStatsThreadPoolPoolResponse `json:"thread_pool"`
	JVM              NodeStatsJVMResponse                       `json:"jvm"`
	Breakers         map[string]NodeStatsBreakersResponse       `json:"breakers"`
	Http             map[string]int                             `json:"http"`
	Transport        NodeStatsTransportResponse                 `json:"transport"`
	Process          NodeStatsProcessResponse                   `json:"process"`
}

// NodeStatsBreakersResponse is a representation of a statistics about the field data circuit breaker
type NodeStatsBreakersResponse struct {
	EstimatedSize int64   `json:"estimated_size_in_bytes"`
	LimitSize     int64   `json:"limit_size_in_bytes"`
	Overhead      float64 `json:"overhead"`
	Tripped       int64   `json:"tripped"`
}

// NodeStatsJVMResponse is a representation of a JVM stats, memory pool information, garbage collection, buffer pools, number of loaded/unloaded classes
type NodeStatsJVMResponse struct {
	BufferPools map[string]NodeStatsJVMBufferPoolResponse `json:"buffer_pools"`
	GC          NodeStatsJVMGCResponse                    `json:"gc"`
	Mem         NodeStatsJVMMemResponse                   `json:"mem"`
}

type NodeStatsJVMGCResponse struct {
	Collectors map[string]NodeStatsJVMGCCollectorResponse `json:"collectors"`
}

type NodeStatsJVMGCCollectorResponse struct {
	CollectionCount int64 `json:"collection_count"`
	CollectionTime  int64 `json:"collection_time_in_millis"`
}

type NodeStatsJVMBufferPoolResponse struct {
	Count         int64 `json:"count"`
	TotalCapacity int64 `json:"total_capacity_in_bytes"`
	Used          int64 `json:"used_in_bytes"`
}

type NodeStatsJVMMemResponse struct {
	HeapCommitted    int64 `json:"heap_committed_in_bytes"`
	HeapUsed         int64 `json:"heap_used_in_bytes"`
	HeapMax          int64 `json:"heap_max_in_bytes"`
	NonHeapCommitted int64 `json:"non_heap_committed_in_bytes"`
	NonHeapUsed      int64 `json:"non_heap_used_in_bytes"`
}

type NodeStatsNetworkResponse struct {
	TCP NodeStatsTCPResponse `json:"tcp"`
}

// NodeStatsTransportResponse is a representation of a transport statistics about sent and received bytes in cluster communication
type NodeStatsTransportResponse struct {
	ServerOpen int64 `json:"server_open"`
	RxCount    int64 `json:"rx_count"`
	RxSize     int64 `json:"rx_size_in_bytes"`
	TxCount    int64 `json:"tx_count"`
	TxSize     int64 `json:"tx_size_in_bytes"`
}

// NodeStatsThreadPoolPoolResponse is a representation of a statistics about each thread pool, including current size, queue and rejected tasks
type NodeStatsThreadPoolPoolResponse struct {
	Threads   int64 `json:"threads"`
	Queue     int64 `json:"queue"`
	Active    int64 `json:"active"`
	Rejected  int64 `json:"rejected"`
	Largest   int64 `json:"largest"`
	Completed int64 `json:"completed"`
}

type NodeStatsTCPResponse struct {
	ActiveOpens  int64 `json:"active_opens"`
	PassiveOpens int64 `json:"passive_opens"`
	CurrEstab    int64 `json:"curr_estab"`
	InSegs       int64 `json:"in_segs"`
	OutSegs      int64 `json:"out_segs"`
	RetransSegs  int64 `json:"retrans_segs"`
	EstabResets  int64 `json:"estab_resets"`
	AttemptFails int64 `json:"attempt_fails"`
	InErrs       int64 `json:"in_errs"`
	OutRsts      int64 `json:"out_rsts"`
}

// NodeStatsIndicesResponse is a representation of a indices stats (size, document count, indexing and deletion times, search times, field cache size, merges and flushes)
type NodeStatsIndicesResponse struct {
	Docs         NodeStatsIndicesDocsResponse
	Store        NodeStatsIndicesStoreResponse
	Indexing     NodeStatsIndicesIndexingResponse
	Merges       NodeStatsIndicesMergesResponse
	Get          NodeStatsIndicesGetResponse
	Search       NodeStatsIndicesSearchResponse
	FieldData    NodeStatsIndicesCacheResponse `json:"fielddata"`
	FilterCache  NodeStatsIndicesCacheResponse `json:"filter_cache"`
	QueryCache   NodeStatsIndicesCacheResponse `json:"query_cache"`
	RequestCache NodeStatsIndicesCacheResponse `json:"request_cache"`
	Flush        NodeStatsIndicesFlushResponse
	Segments     NodeStatsIndicesSegmentsResponse
	Refresh      NodeStatsIndicesRefreshResponse
	Translog     NodeStatsIndicesTranslogResponse
}

type NodeStatsIndicesDocsResponse struct {
	Count   int64 `json:"count"`
	Deleted int64 `json:"deleted"`
}

type NodeStatsIndicesRefreshResponse struct {
	Total     int64 `json:"total"`
	TotalTime int64 `json:"total_time_in_millis"`
}

type NodeStatsIndicesTranslogResponse struct {
	Operations int64 `json:"operations"`
	Size       int64 `json:"size_in_bytes"`
}

type NodeStatsIndicesSegmentsResponse struct {
	Count  int64 `json:"count"`
	Memory int64 `json:"memory_in_bytes"`
}

type NodeStatsIndicesStoreResponse struct {
	Size         int64 `json:"size_in_bytes"`
	ThrottleTime int64 `json:"throttle_time_in_millis"`
}

type NodeStatsIndicesIndexingResponse struct {
	IndexTotal    int64 `json:"index_total"`
	IndexTime     int64 `json:"index_time_in_millis"`
	IndexCurrent  int64 `json:"index_current"`
	DeleteTotal   int64 `json:"delete_total"`
	DeleteTime    int64 `json:"delete_time_in_millis"`
	DeleteCurrent int64 `json:"delete_current"`
}

type NodeStatsIndicesMergesResponse struct {
	Current     int64 `json:"current"`
	CurrentDocs int64 `json:"current_docs"`
	CurrentSize int64 `json:"current_size_in_bytes"`
	Total       int64 `json:"total"`
	TotalDocs   int64 `json:"total_docs"`
	TotalSize   int64 `json:"total_size_in_bytes"`
	TotalTime   int64 `json:"total_time_in_millis"`
}

type NodeStatsIndicesGetResponse struct {
	Total        int64 `json:"total"`
	Time         int64 `json:"time_in_millis"`
	ExistsTotal  int64 `json:"exists_total"`
	ExistsTime   int64 `json:"exists_time_in_millis"`
	MissingTotal int64 `json:"missing_total"`
	MissingTime  int64 `json:"missing_time_in_millis"`
	Current      int64 `json:"current"`
}

type NodeStatsIndicesSearchResponse struct {
	OpenContext  int64 `json:"open_contexts"`
	QueryTotal   int64 `json:"query_total"`
	QueryTime    int64 `json:"query_time_in_millis"`
	QueryCurrent int64 `json:"query_current"`
	FetchTotal   int64 `json:"fetch_total"`
	FetchTime    int64 `json:"fetch_time_in_millis"`
	FetchCurrent int64 `json:"fetch_current"`
}

type NodeStatsIndicesFlushResponse struct {
	Total int64 `json:"total"`
	Time  int64 `json:"total_time_in_millis"`
}

type NodeStatsIndicesCacheResponse struct {
	Evictions  int64 `json:"evictions"`
	MemorySize int64 `json:"memory_size_in_bytes"`
	CacheCount int64 `json:"cache_count"`
	CacheSize  int64 `json:"cache_size"`
	HitCount   int64 `json:"hit_count"`
	MissCount  int64 `json:"miss_count"`
	TotalCount int64 `json:"total_count"`
}

// NodeStatsOSResponse is a representation of a  operating system stats, load average, mem, swap
type NodeStatsOSResponse struct {
	Timestamp int64 `json:"timestamp"`
	Uptime    int64 `json:"uptime_in_millis"`
	// LoadAvg was an array of per-cpu values pre-2.0, and is a string in 2.0
	// Leaving this here in case we want to implement parsing logic later
	LoadAvg json.RawMessage         `json:"load_average"`
	CPU     NodeStatsOSCPUResponse  `json:"cpu"`
	Mem     NodeStatsOSMemResponse  `json:"mem"`
	Swap    NodeStatsOSSwapResponse `json:"swap"`
}

type NodeStatsOSMemResponse struct {
	Free       int64 `json:"free_in_bytes"`
	Used       int64 `json:"used_in_bytes"`
	ActualFree int64 `json:"actual_free_in_bytes"`
	ActualUsed int64 `json:"actual_used_in_bytes"`
}

type NodeStatsOSSwapResponse struct {
	Used int64 `json:"used_in_bytes"`
	Free int64 `json:"free_in_bytes"`
}

type NodeStatsOSCPUResponse struct {
	Sys     int64                      `json:"sys"`
	User    int64                      `json:"user"`
	Idle    int64                      `json:"idle"`
	Steal   int64                      `json:"stolen"`
        LoadAvg NodeStatsOSCPULoadResponse `json:"load_average"`
}

type NodeStatsOSCPULoadResponse struct {
        Load1  float64 `json:"1m"`
        Load5  float64 `json:"5m"`
        Load15 float64 `json:"15m"`
}

// NodeStatsProcessResponse is a representation of a process statistics, memory consumption, cpu usage, open file descriptors
type NodeStatsProcessResponse struct {
	Timestamp int64                       `json:"timestamp"`
	OpenFD    int64                       `json:"open_file_descriptors"`
	MaxFD     int64                       `json:"max_file_descriptors"`
	CPU       NodeStatsProcessCPUResponse `json:"cpu"`
	Memory    NodeStatsProcessMemResponse `json:"mem"`
}

type NodeStatsProcessMemResponse struct {
	Resident     int64 `json:"resident_in_bytes"`
	Share        int64 `json:"share_in_bytes"`
	TotalVirtual int64 `json:"total_virtual_in_bytes"`
}

type NodeStatsProcessCPUResponse struct {
	Percent int64 `json:"percent"`
	Sys     int64 `json:"sys_in_millis"`
	User    int64 `json:"user_in_millis"`
	Total   int64 `json:"total_in_millis"`
}

type NodeStatsHTTPResponse struct {
	CurrentOpen int64 `json:"current_open"`
	TotalOpen   int64 `json:"total_open"`
}

// NodeStatsFSResponse is a representation of a file system information, data path, free disk space, read/write stats
type NodeStatsFSResponse struct {
	Timestamp int64                     `json:"timestamp"`
	Data      []NodeStatsFSDataResponse `json:"data"`
}

type NodeStatsFSDataResponse struct {
	Path          string `json:"path"`
	Mount         string `json:"mount"`
	Device        string `json:"dev"`
	Total         int64  `json:"total_in_bytes"`
	Free          int64  `json:"free_in_bytes"`
	Available     int64  `json:"available_in_bytes"`
	DiskReads     int64  `json:"disk_reads"`
	DiskWrites    int64  `json:"disk_writes"`
	DiskReadSize  int64  `json:"disk_read_size_in_bytes"`
	DiskWriteSize int64  `json:"disk_write_size_in_bytes"`
}

// ClusterHealthResponse is a representation of a Elasticsearch Cluster Health
type ClusterHealthResponse struct {
	ActivePrimaryShards     int64  `json:"active_primary_shards"`
	ActiveShards            int64  `json:"active_shards"`
	ClusterName             string `json:"cluster_name"`
	DelayedUnassignedShards int64  `json:"delayed_unassigned_shards"`
	InitializingShards      int64  `json:"initializing_shards"`
	NumberOfDataNodes       int64  `json:"number_of_data_nodes"`
	NumberOfInFlightFetch   int64  `json:"number_of_in_flight_fetch"`
	NumberOfNodes           int64  `json:"number_of_nodes"`
	NumberOfPendingTasks    int64  `json:"number_of_pending_tasks"`
	RelocatingShards        int64  `json:"relocating_shards"`
	Status                  string `json:"status"`
	TimedOut                bool   `json:"timed_out"`
	UnassignedShards        int64  `json:"unassigned_shards"`
}
