package collector

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultNodeLabels       = []string{"cluster", "host", "name"}
	defaultNodeLabelsValues = func(cluster string, node NodeStatsNodeResponse) []string {
		return []string{cluster, node.Host, node.Name}
	}
)

type Nodes struct {
	logger log.Logger
	client *http.Client
	url    url.URL
	all    bool

	metrics []*nodeMetric
}

type nodeMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(node NodeStatsNodeResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse) []string
}

func NewNodes(logger log.Logger, client *http.Client, url url.URL, all bool) *Nodes {
	return &Nodes{
		logger: logger,
		client: client,
		url:    url,
		all:    all,

		metrics: []*nodeMetric{
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
				Labels: defaultNodeLabelsValues,
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
				Labels: defaultNodeLabelsValues,
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
				Labels: defaultNodeLabelsValues,
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
				Labels: defaultNodeLabelsValues,
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
				Labels: defaultNodeLabelsValues,
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
				Labels: defaultNodeLabelsValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "store_size_bytes"),
					"Current size of stored index data in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return 0
				},
				Labels: defaultNodeLabelsValues,
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
				Labels: defaultNodeLabelsValues,
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
				Labels: defaultNodeLabelsValues,
			},
		},
	}
}

func (c *Nodes) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Desc
	}
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
		for _, metric := range c.metrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(node),
				metric.Labels(nodeStatsResponse.ClusterName, node)...,
			)
		}
	}
}
