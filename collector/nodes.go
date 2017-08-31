package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultNodeLabels          = []string{"cluster", "host", "name"}
	defaultThreadPoolLabels    = append(defaultNodeLabels, "type")
	defaultBreakerLabels       = append(defaultNodeLabels, "breaker")
	defaultFilesystemLabels    = append(defaultNodeLabels, "mount", "path")
	defaultCacheCategoryLabels = append(defaultNodeLabels, "category")
	defaultCacheHitLabels      = append(defaultNodeLabels, "hit")
	defaultCacheMissLabels     = append(defaultNodeLabels, "miss")

	defaultNodeLabelValues = func(cluster string, node NodeStatsNodeResponse) []string {
		return []string{cluster, node.Host, node.Name}
	}
	defaultThreadPoolLabelValues = func(cluster string, node NodeStatsNodeResponse, pool string) []string {
		return append(defaultNodeLabelValues(cluster, node), pool)
	}
	defaultFilesystemLabelValues = func(cluster string, node NodeStatsNodeResponse, mount string, path string) []string {
		return append(defaultNodeLabelValues(cluster, node), mount, path)
	}
)

type nodeMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(node NodeStatsNodeResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse) []string
}

type nodeMetricVec struct {
	Subsystem   string
	Name        string
	Help        string
	Desc        func(subsystem string, name string, help string) *prometheus.Desc
	Labels      []string
	LabelValues func(cluster string, node NodeStatsNodeResponse) [][]string
	Values      func(node NodeStatsNodeResponse) []float64
}

type gcCollectionMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(gcStats NodeStatsJVMGCCollectorResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse, collector string) []string
}

type breakerMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(breakerStats NodeStatsBreakersResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse, breaker string) []string
}

type threadPoolMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse, breaker string) []string
}

type filesystemMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(fsStats NodeStatsFSDataResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse, mount string, path string) []string
}

type Nodes struct {
	logger log.Logger
	client *http.Client
	url    *url.URL
	all    bool

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter

	nodeMetrics         []*nodeMetric
	nodeMetricsVec      []*nodeMetricVec
	gcCollectionMetrics []*gcCollectionMetric
	breakerMetrics      []*breakerMetric
	threadPoolMetrics   []*threadPoolMetric
	filesystemMetrics   []*filesystemMetric
}

func NewNodes(logger log.Logger, client *http.Client, url *url.URL, all bool) *Nodes {
	return &Nodes{
		logger: logger,
		client: client,
		url:    url,
		all:    all,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "node_stats", "up"),
			Help: "Was the last scrape of the ElasticSearch nodes endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "node_stats", "total_scrapes"),
			Help: "Current total ElasticSearch node scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "node_stats", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),

		nodeMetricsVec: []*nodeMetricVec{
			{
				Subsystem: "indices",
				Name:      "query_cache_count",
				Help:      "Query cache count",
				Desc: func(subsystem string, name string, help string) *prometheus.Desc {
					return prometheus.NewDesc(
						prometheus.BuildFQName(namespace, subsystem, name),
						help, defaultCacheCategoryLabels, nil)
				},
				Labels: defaultCacheCategoryLabels,
				LabelValues: func(cluster string, node NodeStatsNodeResponse) [][]string {
					return [][]string{defaultCacheHitLabels, defaultCacheMissLabels}
				},
				Values: func(node NodeStatsNodeResponse) []float64 {
					return []float64{float64(node.Indices.QueryCache.HitCount), float64(node.Indices.QueryCache.MissCount)}
				},
			},
			{
				Subsystem: "indices",
				Name:      "request_cache_count",
				Help:      "Request cache count",
				Desc: func(subsystem string, name string, help string) *prometheus.Desc {
					return prometheus.NewDesc(
						prometheus.BuildFQName(namespace, subsystem, name),
						help, defaultCacheCategoryLabels, nil)
				},
				Labels: defaultCacheCategoryLabels,
				LabelValues: func(cluster string, node NodeStatsNodeResponse) [][]string {
					return [][]string{defaultCacheHitLabels, defaultCacheMissLabels}
				},
				Values: func(node NodeStatsNodeResponse) []float64 {
					return []float64{float64(node.Indices.RequestCache.HitCount), float64(node.Indices.RequestCache.MissCount)}
				},
			},
		},

		nodeMetrics: []*nodeMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "fielddata_memory_size_bytes"),
					"Field data cache memory usage in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.FieldData.MemorySize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "fielddata_evictions"),
					"Evictions from field data",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.FieldData.Evictions)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "filter_cache_memory_size_bytes"),
					"Filter cache memory usage in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.FilterCache.MemorySize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "filter_cache_evictions"),
					"Evictions from filter cache",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.FilterCache.Evictions)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "query_cache_memory_size_bytes"),
					"Query cache memory usage in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.QueryCache.MemorySize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "query_cache_evictions"),
					"Evictions from query cache",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.QueryCache.Evictions)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "query_cache_total"),
					"Query cache total count",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.QueryCache.TotalCount)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "query_cache_cache_size"),
					"Query cache cache size",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.QueryCache.CacheSize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "query_cache_cache_count"),
					"Query cache cache count",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.QueryCache.CacheCount)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "request_cache_memory_size_bytes"),
					"Request cache memory usage in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.RequestCache.MemorySize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "request_cache_evictions"),
					"Evictions from request cache",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.RequestCache.Evictions)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "translog_operations"),
					"Total translog operations",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Translog.Operations)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "translog_size_in_bytes"),
					"Total translog size in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Translog.Size)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "get_time_seconds"),
					"Total get time in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Get.Time / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "get_total"),
					"Total get",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Get.Total)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "get_missing_time_seconds"),
					"Total time of get missing in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Get.MissingTime / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "get_missing_total"),
					"Total get missing",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Get.MissingTotal)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "get_exists_time_seconds"),
					"Total time get exists in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Get.ExistsTime / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "get_exists_total"),
					"Total get exists operations",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Get.ExistsTotal)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_refresh", "time_seconds_total"),
					"Total refreshes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Refresh.TotalTime / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_refresh", "total"),
					"Total time spent refreshing in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Refresh.Total)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "search_query_time_seconds"),
					"Total search query time in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Search.QueryTime / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "search_query_total"),
					"Total number of queries",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Search.QueryTotal)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "search_fetch_time_seconds"),
					"Total search fetch time in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Search.FetchTime / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "search_fetch_total"),
					"Total number of fetches",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Search.FetchTotal)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "docs"),
					"Count of documents on this node",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Docs.Count)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "docs_deleted"),
					"Count of deleted documents on this node",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Docs.Deleted)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "store_size_bytes"),
					"Current size of stored index data in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Store.Size)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "store_throttle_time_seconds_total"),
					"Throttle time for index store in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Store.ThrottleTime / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "segments_memory_bytes"),
					"Current memory size of segments in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Segments.Memory)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "segments_count"),
					"Count of index segments on this node",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Segments.Count)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "flush_total"),
					"Total flushes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Flush.Total)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "flush_time_seconds"),
					"Cumulative flush time in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Flush.Time / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_indexing", "index_time_seconds_total"),
					"Cumulative index time in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Indexing.IndexTime / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_indexing", "index_total"),
					"Total index calls",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Indexing.IndexTotal)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_indexing", "delete_time_seconds_total"),
					"Total time indexing delete in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Indexing.DeleteTime / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_indexing", "delete_total"),
					"Total indexing deletes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Indexing.DeleteTotal)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_merges", "total"),
					"Total merges",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Merges.Total)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_merges", "docs_total"),
					"Cumulative docs merged",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Merges.TotalDocs)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_merges", "total_size_bytes_total"),
					"Total merge size in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Merges.TotalSize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_merges", "total_time_seconds_total"),
					"Total time spent merging in seconds",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Merges.TotalTime / 1000)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_memory", "used_bytes"),
					"JVM memory currently used by area",
					append(defaultNodeLabels, "area"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.JVM.Mem.HeapUsed)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "heap")
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_memory", "used_bytes"),
					"JVM memory currently used by area",
					append(defaultNodeLabels, "area"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.JVM.Mem.NonHeapUsed)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "non-heap")
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_memory", "max_bytes"),
					"JVM memory max",
					append(defaultNodeLabels, "area"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.JVM.Mem.HeapMax)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "heap")
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_memory", "committed_bytes"),
					"JVM memory currently committed by area",
					append(defaultNodeLabels, "area"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.JVM.Mem.HeapCommitted)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "heap")
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_memory", "committed_bytes"),
					"JVM memory currently committed by area",
					append(defaultNodeLabels, "area"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.JVM.Mem.NonHeapCommitted)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "non-heap")
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "process", "cpu_percent"),
					"Percent CPU used by process",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Process.CPU.Percent)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "process", "mem_resident_size_bytes"),
					"Resident memory in use by process in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Process.Memory.Resident)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "process", "mem_share_size_bytes"),
					"Shared memory in use by process in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Process.Memory.Share)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "process", "mem_virtual_size_bytes"),
					"Total virtual memory used in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Process.Memory.TotalVirtual)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "process", "open_files_count"),
					"Open file descriptors",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Process.OpenFD)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "process", "cpu_time_seconds_sum"),
					"Process CPU time in seconds",
					append(defaultNodeLabels, "type"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Process.CPU.Total / 1000)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "total")
				},
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "process", "cpu_time_seconds_sum"),
					"Process CPU time in seconds",
					append(defaultNodeLabels, "type"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Process.CPU.Sys / 1000)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "sys")
				},
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "process", "cpu_time_seconds_sum"),
					"Process CPU time in seconds",
					append(defaultNodeLabels, "type"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Process.CPU.User / 1000)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "user")
				},
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "transport", "rx_packets_total"),
					"Count of packets received",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Transport.RxCount)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "transport", "rx_size_bytes_total"),
					"Total number of bytes received",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Transport.RxSize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "transport", "tx_packets_total"),
					"Count of packets sent",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Transport.TxCount)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "transport", "tx_size_bytes_total"),
					"Total number of bytes sent",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Transport.TxSize)
				},
				Labels: defaultNodeLabelValues,
			},
		},
		gcCollectionMetrics: []*gcCollectionMetric{
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_gc", "collection_seconds_count"),
					"Count of JVM GC runs",
					append(defaultNodeLabels, "gc"), nil,
				),
				Value: func(gcStats NodeStatsJVMGCCollectorResponse) float64 {
					return float64(gcStats.CollectionCount)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse, collector string) []string {
					return append(defaultNodeLabelValues(cluster, node), collector)
				},
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_gc", "collection_seconds_sum"),
					"GC run time in seconds",
					append(defaultNodeLabels, "gc"), nil,
				),
				Value: func(gcStats NodeStatsJVMGCCollectorResponse) float64 {
					return float64(gcStats.CollectionTime / 1000)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse, collector string) []string {
					return append(defaultNodeLabelValues(cluster, node), collector)
				},
			},
		},
		breakerMetrics: []*breakerMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "breakers", "estimated_size_bytes"),
					"Estimated size in bytes of breaker",
					defaultBreakerLabels, nil,
				),
				Value: func(breakerStats NodeStatsBreakersResponse) float64 {
					return float64(breakerStats.EstimatedSize)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse, breaker string) []string {
					return append(defaultNodeLabelValues(cluster, node), breaker)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "breakers", "limit_size_bytes"),
					"Limit size in bytes for breaker",
					defaultBreakerLabels, nil,
				),
				Value: func(breakerStats NodeStatsBreakersResponse) float64 {
					return float64(breakerStats.LimitSize)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse, breaker string) []string {
					return append(defaultNodeLabelValues(cluster, node), breaker)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "breakers", "tripped"),
					"tripped for breaker",
					defaultBreakerLabels, nil,
				),
				Value: func(breakerStats NodeStatsBreakersResponse) float64 {
					return float64(breakerStats.Tripped)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse, breaker string) []string {
					return append(defaultNodeLabelValues(cluster, node), breaker)
				},
			},
		},
		threadPoolMetrics: []*threadPoolMetric{
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "completed_count"),
					"Thread Pool operations completed",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Completed)
				},
				Labels: defaultThreadPoolLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "rejected_count"),
					"Thread Pool operations rejected",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Rejected)
				},
				Labels: defaultThreadPoolLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "active_count"),
					"Thread Pool threads active",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Active)
				},
				Labels: defaultThreadPoolLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "largest_count"),
					"Thread Pool largest threads count",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Largest)
				},
				Labels: defaultThreadPoolLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "queue_count"),
					"Thread Pool operations queued",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Queue)
				},
				Labels: defaultThreadPoolLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "threads_count"),
					"Thread Pool current threads count",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Threads)
				},
				Labels: defaultThreadPoolLabelValues,
			},
		},
		filesystemMetrics: []*filesystemMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "filesystem_data", "available_bytes"),
					"Available space on block device in bytes",
					defaultFilesystemLabels, nil,
				),
				Value: func(fsStats NodeStatsFSDataResponse) float64 {
					return float64(fsStats.Available)
				},
				Labels: defaultFilesystemLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "filesystem_data", "free_bytes"),
					"Free space on block device in bytes",
					defaultFilesystemLabels, nil,
				),
				Value: func(fsStats NodeStatsFSDataResponse) float64 {
					return float64(fsStats.Free)
				},
				Labels: defaultFilesystemLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "filesystem_data", "size_bytes"),
					"Size of block device in bytes",
					defaultFilesystemLabels, nil,
				),
				Value: func(fsStats NodeStatsFSDataResponse) float64 {
					return float64(fsStats.Total)
				},
				Labels: defaultFilesystemLabelValues,
			},
		},
	}
}

func (c *Nodes) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.nodeMetrics {
		ch <- metric.Desc
	}
	for _, metric := range c.nodeMetricsVec {
		ch <- metric.Desc(metric.Subsystem, metric.Name, metric.Help)
	}
	for _, metric := range c.gcCollectionMetrics {
		ch <- metric.Desc
	}
	for _, metric := range c.threadPoolMetrics {
		ch <- metric.Desc
	}
	for _, metric := range c.filesystemMetrics {
		ch <- metric.Desc
	}
	ch <- c.up.Desc()
	ch <- c.totalScrapes.Desc()
	ch <- c.jsonParseFailures.Desc()
}

func (c *Nodes) fetchAndDecodeNodeStats() (nodeStatsResponse, error) {
	var nsr nodeStatsResponse

	u := *c.url
	u.Path = "/_nodes/_local/stats"
	if c.all {
		u.Path = "/_nodes/stats"
	}

	res, err := c.client.Get(u.String())
	if err != nil {
		return nsr, fmt.Errorf("failed to get cluster health from %s://%s:%s/%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nsr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&nsr); err != nil {
		c.jsonParseFailures.Inc()
		return nsr, err
	}
	return nsr, nil
}

func (c *Nodes) Collect(ch chan<- prometheus.Metric) {
	c.totalScrapes.Inc()
	defer func() {
		ch <- c.up
		ch <- c.totalScrapes
		ch <- c.jsonParseFailures
	}()

	nodeStatsResponse, err := c.fetchAndDecodeNodeStats()
	if err != nil {
		c.up.Set(0)
		level.Warn(c.logger).Log(
			"msg", "failed to fetch and decode node stats",
			"err", err,
		)
		return
	}
	c.up.Set(1)

	for _, node := range nodeStatsResponse.Nodes {
		for _, metric := range c.nodeMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(node),
				metric.Labels(nodeStatsResponse.ClusterName, node)...,
			)
		}
		for _, metric := range c.nodeMetricsVec {
			labelValues := metric.LabelValues(nodeStatsResponse.ClusterName, node)
			values := metric.Values(node)

			for i, labelValue := range labelValues {
				counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
					Namespace: namespace,
					Subsystem: metric.Subsystem,
					Name:      metric.Name,
					Help:      metric.Help},
					metric.Labels)
				counterVec.WithLabelValues(labelValue...).Set(values[i])
				counterVec.Collect(ch)
			}

		}

		// GC Stats
		for collector, gcStats := range node.JVM.GC.Collectors {
			for _, metric := range c.gcCollectionMetrics {
				ch <- prometheus.MustNewConstMetric(
					metric.Desc,
					metric.Type,
					metric.Value(gcStats),
					metric.Labels(nodeStatsResponse.ClusterName, node, collector)...,
				)
			}
		}

		// Breaker stats
		for breaker, bstats := range node.Breakers {
			for _, metric := range c.breakerMetrics {
				ch <- prometheus.MustNewConstMetric(
					metric.Desc,
					metric.Type,
					metric.Value(bstats),
					metric.Labels(nodeStatsResponse.ClusterName, node, breaker)...,
				)
			}
		}

		// Thread Pool stats
		for pool, pstats := range node.ThreadPool {
			for _, metric := range c.threadPoolMetrics {
				ch <- prometheus.MustNewConstMetric(
					metric.Desc,
					metric.Type,
					metric.Value(pstats),
					metric.Labels(nodeStatsResponse.ClusterName, node, pool)...,
				)
			}
		}

		// File System Stats
		for _, fsStats := range node.FS.Data {
			for _, metric := range c.filesystemMetrics {
				ch <- prometheus.MustNewConstMetric(
					metric.Desc,
					metric.Type,
					metric.Value(fsStats),
					metric.Labels(nodeStatsResponse.ClusterName, node, fsStats.Mount, fsStats.Path)...,
				)
			}
		}
	}
}
