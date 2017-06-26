package collector

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type Nodes struct {
	logger log.Logger
	client *http.Client
	url    url.URL
	all    bool

	IndicesFielddataMemory        *prometheus.Desc
	IndicesFilterCacheMemorySize  *prometheus.Desc
	IndicesQueryCacheMemorySize   *prometheus.Desc
	IndicesRequestCacheMemorySize *prometheus.Desc
	IndicesDocs                   *prometheus.Desc
	IndicesDocsDeleted            *prometheus.Desc
	IndicesStoreSize              *prometheus.Desc
	IndicesSegmentsMemory         *prometheus.Desc
	IndicesSegmentsCount          *prometheus.Desc
}

func NewNodes(logger log.Logger, client *http.Client, url url.URL, all bool) *Nodes {
	defaultLabels := []string{"cluster", "host", "name"}

	return &Nodes{
		logger: logger,
		client: client,
		url:    url,
		all:    all,

		IndicesFielddataMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "indices", "fielddata_memory_size_bytes"),
			"Field data cache memory usage in bytes",
			defaultLabels, nil,
		),
		IndicesFilterCacheMemorySize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "indices", "filter_cache_memory_size_bytes"),
			"Filter cache memory usage in bytes",
			defaultLabels, nil,
		),
		IndicesQueryCacheMemorySize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "indices", "query_cache_memory_size_bytes"),
			"Query cache memory usage in bytes",
			defaultLabels, nil,
		),
		IndicesRequestCacheMemorySize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "indices", "request_cache_memory_size_bytes"),
			"Request cache memory usage in bytes",
			defaultLabels, nil,
		),
		IndicesDocs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "indices", "docs"),
			"Count of documents on this node",
			defaultLabels, nil,
		),
		IndicesDocsDeleted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "indices", "docs_deleted"),
			"Count of deleted documents on this node",
			defaultLabels, nil,
		),
		IndicesStoreSize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "indices", "store_size_bytes"),
			"Current size of stored index data in bytes",
			defaultLabels, nil,
		),
		IndicesSegmentsMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "indices", "segments_memory_bytes"),
			"Current memory size of segments in bytes",
			defaultLabels, nil,
		),
		IndicesSegmentsCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "indices", "segments_count"),
			"Count of index segments on this node",
			defaultLabels, nil,
		),
	}
}

func (c *Nodes) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.IndicesFielddataMemory
	ch <- c.IndicesFilterCacheMemorySize
	ch <- c.IndicesQueryCacheMemorySize
	ch <- c.IndicesRequestCacheMemorySize
	ch <- c.IndicesDocs
	ch <- c.IndicesDocsDeleted
	ch <- c.IndicesStoreSize
	ch <- c.IndicesSegmentsMemory
	ch <- c.IndicesSegmentsCount
}

func (c *Nodes) Collect(ch chan<- prometheus.Metric) {
	path := "/_nodes/_local/stats"
	if c.all {
		path = "/_nodes/stats"
	}
	c.url.Path = path

	res, err := c.client.Get(c.url.String())
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "failed to get nodes",
			"url", c.url.String(),
			"err", err,
		)
		return
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)

	var nodeStatsResponse NodeStatsResponse
	if err := dec.Decode(&nodeStatsResponse); err != nil {
		level.Warn(c.logger).Log(
			"msg", "failed to decode nodes",
			"err", err,
		)
		return
	}

	for _, node := range nodeStatsResponse.Nodes {
		ch <- prometheus.MustNewConstMetric(
			c.IndicesFielddataMemory,
			prometheus.GaugeValue,
			float64(node.Indices.FieldData.MemorySize),
			nodeStatsResponse.ClusterName, node.Host, node.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IndicesFilterCacheMemorySize,
			prometheus.GaugeValue,
			float64(node.Indices.FilterCache.MemorySize),
			nodeStatsResponse.ClusterName, node.Host, node.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IndicesQueryCacheMemorySize,
			prometheus.GaugeValue,
			float64(node.Indices.QueryCache.MemorySize),
			nodeStatsResponse.ClusterName, node.Host, node.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IndicesRequestCacheMemorySize,
			prometheus.GaugeValue,
			float64(node.Indices.RequestCache.MemorySize),
			nodeStatsResponse.ClusterName, node.Host, node.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IndicesDocs,
			prometheus.GaugeValue,
			float64(node.Indices.Docs.Count),
			nodeStatsResponse.ClusterName, node.Host, node.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IndicesDocsDeleted,
			prometheus.GaugeValue,
			float64(node.Indices.Docs.Deleted),
			nodeStatsResponse.ClusterName, node.Host, node.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IndicesStoreSize,
			prometheus.GaugeValue,
			float64(node.Indices.Store.Size),
			nodeStatsResponse.ClusterName, node.Host, node.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IndicesSegmentsMemory,
			prometheus.GaugeValue,
			float64(node.Indices.Segments.Memory),
			nodeStatsResponse.ClusterName, node.Host, node.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IndicesSegmentsCount,
			prometheus.GaugeValue,
			float64(node.Indices.Segments.Count),
			nodeStatsResponse.ClusterName, node.Host, node.Name,
		)
	}
}
