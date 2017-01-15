package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Exporter collects Elasticsearch stats from the given server and exports
// them using the prometheus metrics package.
type Exporter struct {
	URI   string
	mutex sync.RWMutex

	up prometheus.Gauge

	gauges      map[string]*prometheus.GaugeVec
	gaugeVecs   map[string]*prometheus.GaugeVec
	counters    map[string]*prometheus.CounterVec
	counterVecs map[string]*prometheus.CounterVec

	allNodes bool

	client *http.Client
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, timeout time.Duration, allNodes bool) *Exporter {
	counters := make(map[string]*prometheus.CounterVec, len(counterMetrics))
	counterVecs := make(map[string]*prometheus.CounterVec, len(counterVecMetrics))
	gauges := make(map[string]*prometheus.GaugeVec, len(gaugeMetrics))
	gaugeVecs := make(map[string]*prometheus.GaugeVec, len(gaugeVecMetrics))

	for name, info := range counterVecMetrics {
		counterVecs[name] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      name,
			Help:      info.help,
		}, append([]string{"cluster", "node"}, info.labels...))
	}

	for name, info := range gaugeVecMetrics {
		gaugeVecs[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      name,
			Help:      info.help,
		}, append([]string{"cluster", "node"}, info.labels...))
	}

	for name, help := range counterMetrics {
		counters[name] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      name,
			Help:      help,
		}, []string{"cluster", "node"})
	}

	for name, help := range gaugeMetrics {
		gauges[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      name,
			Help:      help,
		}, []string{"cluster", "node"})
	}

	// Init our exporter.
	return &Exporter{
		URI: uri,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the Elasticsearch instance query successful?",
		}),

		counters:    counters,
		counterVecs: counterVecs,
		gauges:      gauges,
		gaugeVecs:   gaugeVecs,

		allNodes: allNodes,

		client: &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, timeout)
					if err != nil {
						return nil, err
					}
					if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
						return nil, err
					}
					return c, nil
				},
			},
		},
	}
}

// Describe describes all the metrics ever exported by the elasticsearch
// exporter. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up.Desc()

	for _, vec := range e.counters {
		vec.Describe(ch)
	}

	for _, vec := range e.gauges {
		vec.Describe(ch)
	}

	for _, vec := range e.counterVecs {
		vec.Describe(ch)
	}

	for _, vec := range e.gaugeVecs {
		vec.Describe(ch)
	}

}

// Collect fetches the stats from configured elasticsearch location and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	// Reset metrics.
	for _, vec := range e.gaugeVecs {
		vec.Reset()
	}

	for _, vec := range e.counterVecs {
		vec.Reset()
	}

	for _, vec := range e.gauges {
		vec.Reset()
	}

	for _, vec := range e.counters {
		vec.Reset()
	}

	defer func() { ch <- e.up }()

	resp, err := e.client.Get(e.URI)
	if err != nil {
		e.up.Set(0)
		log.Println("Error while querying Elasticsearch:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read ES response body:", err)
		e.up.Set(0)
		return
	}

	e.up.Set(1)

	var allStats NodeStatsResponse
	err = json.Unmarshal(body, &allStats)
	if err != nil {
		log.Println("Failed to unmarshal JSON into struct:", err)
		return
	}

	// If we aren't polling all nodes, make sure we only got one response.
	if !e.allNodes && len(allStats.Nodes) != 1 {
		log.Println("Unexpected number of nodes returned.")
	}

	for _, stats := range allStats.Nodes {
		// GC Stats
		for collector, gcstats := range stats.JVM.GC.Collectors {
			e.counterVecs["jvm_gc_collection_seconds_count"].WithLabelValues(allStats.ClusterName, stats.Host, collector).Set(float64(gcstats.CollectionCount))
			e.counterVecs["jvm_gc_collection_seconds_sum"].WithLabelValues(allStats.ClusterName, stats.Host, collector).Set(float64(gcstats.CollectionTime / 1000))
		}

		// Breaker stats
		for breaker, bstats := range stats.Breakers {
			e.gaugeVecs["breakers_estimated_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Host, breaker).Set(float64(bstats.EstimatedSize))
			e.gaugeVecs["breakers_limit_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Host, breaker).Set(float64(bstats.LimitSize))
		}

		// Thread Pool stats
		for pool, pstats := range stats.ThreadPool {
			e.counterVecs["thread_pool_completed_count"].WithLabelValues(allStats.ClusterName, stats.Host, pool).Set(float64(pstats.Completed))
			e.counterVecs["thread_pool_rejected_count"].WithLabelValues(allStats.ClusterName, stats.Host, pool).Set(float64(pstats.Rejected))

			e.gaugeVecs["thread_pool_active_count"].WithLabelValues(allStats.ClusterName, stats.Host, pool).Set(float64(pstats.Active))
			e.gaugeVecs["thread_pool_threads_count"].WithLabelValues(allStats.ClusterName, stats.Host, pool).Set(float64(pstats.Active))
			e.gaugeVecs["thread_pool_largest_count"].WithLabelValues(allStats.ClusterName, stats.Host, pool).Set(float64(pstats.Active))
			e.gaugeVecs["thread_pool_queue_count"].WithLabelValues(allStats.ClusterName, stats.Host, pool).Set(float64(pstats.Active))
		}

		// JVM Memory Stats
		e.gaugeVecs["jvm_memory_committed_bytes"].WithLabelValues(allStats.ClusterName, stats.Host, "heap").Set(float64(stats.JVM.Mem.HeapCommitted))
		e.gaugeVecs["jvm_memory_used_bytes"].WithLabelValues(allStats.ClusterName, stats.Host, "heap").Set(float64(stats.JVM.Mem.HeapUsed))
		e.gaugeVecs["jvm_memory_max_bytes"].WithLabelValues(allStats.ClusterName, stats.Host, "heap").Set(float64(stats.JVM.Mem.HeapMax))
		e.gaugeVecs["jvm_memory_committed_bytes"].WithLabelValues(allStats.ClusterName, stats.Host, "non-heap").Set(float64(stats.JVM.Mem.NonHeapCommitted))
		e.gaugeVecs["jvm_memory_used_bytes"].WithLabelValues(allStats.ClusterName, stats.Host, "non-heap").Set(float64(stats.JVM.Mem.NonHeapUsed))

		// Indices Stats
		e.gauges["indices_fielddata_memory_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.FieldData.MemorySize))
		e.counters["indices_fielddata_evictions"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.FieldData.Evictions))

		e.gauges["indices_filter_cache_memory_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.FilterCache.MemorySize))
		e.counters["indices_filter_cache_evictions"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.FilterCache.Evictions))

		e.gauges["indices_query_cache_memory_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.QueryCache.MemorySize))
		e.counters["indices_query_cache_evictions"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.QueryCache.Evictions))

		e.gauges["indices_request_cache_memory_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.QueryCache.MemorySize))
		e.counters["indices_request_cache_evictions"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.QueryCache.Evictions))

		e.gauges["indices_docs"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Docs.Count))
		e.gauges["indices_docs_deleted"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Docs.Deleted))

		e.gauges["indices_segments_memory_bytes"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Segments.Memory))
		e.gauges["indices_segments_count"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Segments.Count))

		e.gauges["indices_store_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Store.Size))
		e.counters["indices_store_throttle_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Store.ThrottleTime))

		e.counters["indices_flush_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Flush.Total))
		e.counters["indices_flush_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Flush.Time))

		e.counters["indices_indexing_index_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Indexing.IndexTime))
		e.counters["indices_indexing_index_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Indexing.IndexTotal))

		e.counters["indices_merges_total_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Merges.TotalTime))
		e.counters["indices_merges_total_size_bytes_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Merges.TotalSize))
		e.counters["indices_merges_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Merges.Total))

		e.counters["indices_refresh_total_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Refresh.TotalTime))
		e.counters["indices_refresh_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Indices.Refresh.Total))

		// Transport Stats
		e.counters["transport_rx_packets_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Transport.RxCount))
		e.counters["transport_rx_size_bytes_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Transport.RxSize))
		e.counters["transport_tx_packets_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Transport.TxCount))
		e.counters["transport_tx_size_bytes_total"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Transport.TxSize))

		// Process Stats
		e.gauges["process_cpu_percent"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Process.CPU.Percent))
		e.gauges["process_mem_resident_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Process.Memory.Resident))
		e.gauges["process_mem_share_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Process.Memory.Share))
		e.gauges["process_mem_virtual_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Process.Memory.TotalVirtual))
		e.gauges["process_open_files_count"].WithLabelValues(allStats.ClusterName, stats.Host).Set(float64(stats.Process.OpenFD))

		e.counterVecs["process_cpu_time_seconds_sum"].WithLabelValues(allStats.ClusterName, stats.Host, "total").Set(float64(stats.Process.CPU.Total / 1000))
		e.counterVecs["process_cpu_time_seconds_sum"].WithLabelValues(allStats.ClusterName, stats.Host, "sys").Set(float64(stats.Process.CPU.Sys / 1000))
		e.counterVecs["process_cpu_time_seconds_sum"].WithLabelValues(allStats.ClusterName, stats.Host, "user").Set(float64(stats.Process.CPU.User / 1000))
	}

	// Report metrics.

	for _, vec := range e.counterVecs {
		vec.Collect(ch)
	}

	for _, vec := range e.gaugeVecs {
		vec.Collect(ch)
	}

	for _, vec := range e.counters {
		vec.Collect(ch)
	}

	for _, vec := range e.gauges {
		vec.Collect(ch)
	}
}
