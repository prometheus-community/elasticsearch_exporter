package collector

import (
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultIndexLabels      = []string{"cluster", "index"}
	defaultIndexLabelValues = func(clusterName string, indexName string) []string {
		return []string{clusterName, indexName}
	}
)

type indexMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(indexStats IndexStatsIndexResponse) float64
	Labels func(clusterName string, indexName string) []string
}

type Indices struct {
	logger        log.Logger
	client        *http.Client
	url           *url.URL
	all           bool
	exportIndices bool

	up                prometheus.Gauge
	totalScrapes      prometheus.Counter
	jsonParseFailures prometheus.Counter

	nodeMetrics  []*nodeMetric
	indexMetrics []*indexMetric
}

func NewIndices(logger log.Logger, client *http.Client, url *url.URL, all bool, exportIndices bool) *Indices {
	return &Indices{
		logger:        logger,
		client:        client,
		url:           url,
		all:           all,
		exportIndices: exportIndices,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "index_stats", "up"),
			Help: "Was the last scrape of the ElasticSearch index endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "index_stats", "total_scrapes"),
			Help: "Current total ElasticSearch index scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "index_stats", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),

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
		},
		indexMetrics: []*indexMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "docs_primary"),
					"Count of documents which only primary shards",
					defaultIndexLabels, nil,
				),
				Value: func(indexStats IndexStatsIndexResponse) float64 {
					return float64(indexStats.Primaries.Docs.Count)
				},
				Labels: defaultIndexLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "store_size_bytes_primary"),
					"Current total size of stored index data in bytes which only primary shards on all nodes",
					defaultIndexLabels, nil,
				),
				Value: func(indexStats IndexStatsIndexResponse) float64 {
					return float64(indexStats.Primaries.Store.SizeInBytes)
				},
				Labels: defaultIndexLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "store_size_bytes_total"),
					"Current total size of stored index data in bytes which all shards on all nodes",
					defaultIndexLabels, nil,
				),
				Value: func(indexStats IndexStatsIndexResponse) float64 {
					return float64(indexStats.Total.Store.SizeInBytes)
				},
				Labels: defaultIndexLabelValues,
			},
		},
	}
}

func (i *Indices) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range i.indexMetrics {
		ch <- metric.Desc
	}
	ch <- i.up.Desc()
	ch <- i.totalScrapes.Desc()
	ch <- i.jsonParseFailures.Desc()
}

func (i *Indices) Collect(ch chan<- prometheus.Metric, clusterHealthResponse clusterHealthResponse,
	nodeStatsResponse nodeStatsResponse, indexStatsResponse indexStatsResponse) {
	i.totalScrapes.Inc()
	defer func() {
		ch <- i.up
		ch <- i.totalScrapes
		ch <- i.jsonParseFailures
	}()

	// Node stats
	for _, node := range nodeStatsResponse.Nodes {
		for _, metric := range i.nodeMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(node),
				metric.Labels(nodeStatsResponse.ClusterName, node)...,
			)
		}
	}

	// Index stats
	for indexName, indexStats := range indexStatsResponse.Indices {
		for _, metric := range i.indexMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(indexStats),
				metric.Labels(clusterHealthResponse.ClusterName, indexName)...,
			)
		}
	}
}
