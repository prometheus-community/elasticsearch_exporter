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
	defaultNodeLabels       = []string{"cluster", "host", "name"}
	defaultThreadPoolLabels = append(defaultNodeLabels, "type")
	defaultBreakerLabels    = append(defaultNodeLabels, "breaker")
	defaultFilesystemLabels = append(defaultNodeLabels, "mount", "path")

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

		nodeMetrics: []*nodeMetric{
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
