package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "elasticsearch"
)

// Elasticsearch Node Stats Structs

type NodeStatsResponse struct {
	ClusterName string `json:"cluster_name"`
	Nodes       map[string]NodeStatsNodeResponse
}

type NodeStatsNodeResponse struct {
	Name             string                                     `json:"name"`
	Timestamp        int64                                      `json:"timestamp"`
	TransportAddress string                                     `json:"transport_address"`
	Hostname         string                                     `json:"hostname"`
	Indices          NodeStatsIndicesResponse                   `json:"indices"`
	OS               NodeStatsOSResponse                        `json:"os"`
	Network          NodeStatsNetworkResponse                   `json:"network"`
	FS               NodeStatsFSResponse                        `json:"fs"`
	ThreadPool       map[string]NodeStatsThreadPoolPoolResponse `json:"thread_pool"`
	JVM              NodeStatsJVMResponse                       `json:"jvm"`
	Breakers         map[string]NodeStatsBreakersResponse       `json:"breakers"`
	Transport        NodeStatsTransportResponse                 `json:"transport"`
}

type NodeStatsBreakersResponse struct {
	EstimatedSize int64   `json:"estimated_size_in_bytes"`
	LimitSize     int64   `json:"limit_size_in_bytes"`
	Overhead      float64 `json:"overhead"`
	Tripped       int64   `json:"tripped"`
}

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

type NodeStatsTransportResponse struct {
	ServerOpen int64 `json:"server_open"`
	RxCount    int64 `json:"rx_count"`
	RxSize     int64 `json:"rx_size_in_bytes"`
	TxCount    int64 `json:"tx_count"`
	TxSize     int64 `json:"tx_size_in_bytes"`
}

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

type NodeStatsIndicesResponse struct {
	Docs        NodeStatsIndicesDocsResponse
	Store       NodeStatsIndicesStoreResponse
	Indexing    NodeStatsIndicesIndexingResponse
	Merges      NodeStatsIndicesMergesResponse
	Get         NodeStatsIndicesGetResponse
	Search      NodeStatsIndicesSearchResponse
	FieldData   NodeStatsIndicesFieldDataResponse
	FilterCache NodeStatsIndicesFieldDataResponse
	Flush       NodeStatsIndicesFlushResponse
	Segments    NodeStatsIndicesSegmentsResponse
}

type NodeStatsIndicesDocsResponse struct {
	Count   int64 `json:"count"`
	Deleted int64 `json:"deleted"`
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

type NodeStatsIndicesFieldDataResponse struct {
	Evictions  int64 `json:"evictions"`
	MemorySize int64 `json:"memory_size_in_bytes"`
}

type NodeStatsOSResponse struct {
	Timestamp int64                   `json:"timestamp"`
	Uptime    int64                   `json:"uptime_in_millis"`
	LoadAvg   []float64               `json:"load_average"`
	CPU       NodeStatsOSCPUResponse  `json:"cpu"`
	Mem       NodeStatsOSMemResponse  `json:"mem"`
	Swap      NodeStatsOSSwapResponse `json:"swap"`
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
	Sys   int64 `json:"sys"`
	User  int64 `json:"user"`
	Idle  int64 `json:"idle"`
	Steal int64 `json:"stolen"`
}

type NodeStatsProcessResponse struct {
	Timestamp int64                       `json:"timestamp"`
	OpenFD    int64                       `json:"open_file_descriptors"`
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

//----------

type VecInfo struct {
	help   string
	labels []string
}

var (
	gaugeMetrics = map[string]string{
		"indices_fielddata_evictions":               "Evictions from field data",
		"indices_fielddata_memory_size_in_bytes":    "Field data cache memory usage in bytes",
		"indices_filter_cache_evictions":            "Evictions from field data",
		"indices_filter_cache_memory_size_in_bytes": "Field data cache memory usage in bytes",
		"indices_docs_count":                        "Count of documents on this node",
		"indices_docs_deleted":                      "Count of deleted documents on this node",
		"indices_store_size_in_bytes":               "Size of stored index data in bytes",
		"indices_segments_memory_in_bytes":          "Memory size of segments in bytes",
		"jvm_mem_heap_committed_in_bytes":           "JVM heap memory committed",
		"jvm_mem_heap_used_in_bytes":                "JVM heap memory used",
		"jvm_mem_heap_max_in_bytes":                 "JVM heap memory max",
		"jvm_mem_non_heap_committed_in_bytes":       "JVM non-heap memory committed",
		"jvm_mem_non_heap_used_in_bytes":            "JVM non-heap memory used",
	}
	counterMetrics = map[string]string{
		"indices_flush_total":                   "Total flushes",
		"indices_flush_time_in_millis":          "Cumulative flush time",
		"transport_rx_count":                    "Count of packets received",
		"transport_rx_size_in_bytes":            "Bytes received",
		"transport_tx_count":                    "Count of packets sent",
		"transport_tx_size_in_bytes":            "Bytes sent",
		"indices_store_throttle_time_in_millis": "Throttle time for index store",
		"indices_indexing_index_total":          "Total index calls",
		"indices_indexing_index_time_in_millis": "Cumulative index time",
		"indices_merges_total":                  "Total merges",
		"indices_merges_total_docs":             "Cumulative docs merged",
		"indices_merges_total_size_in_bytes":    "Total merge size in bytes",
		"indices_merges_total_time_in_millis":   "Total time spent merging",
	}
	counterVecMetrics = map[string]*VecInfo{
		"jvm_gc_collection_count": &VecInfo{
			help:   "Count of JVM GC runs",
			labels: []string{"collector"},
		},
		"jvm_gc_collection_time_in_millis": &VecInfo{
			help:   "GC run time",
			labels: []string{"collector"},
		},
	}

	gaugeVecMetrics = map[string]*VecInfo{
		"breakers_estimated_size_in_bytes": &VecInfo{
			help:   "Estimated size in bytes of breaker",
			labels: []string{"breaker"},
		},
		"breakers_limit_size_in_bytes": &VecInfo{
			help:   "Limit size in bytes for breaker",
			labels: []string{"breaker"},
		},
	}
)

// Exporter collects Elasticsearch stats from the given server and exports
// them using the prometheus metrics package.
type Exporter struct {
	URI   string
	mutex sync.RWMutex

	up prometheus.Gauge

	gauges   map[string]*prometheus.GaugeVec
	counters map[string]*prometheus.CounterVec

	client *http.Client
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, timeout time.Duration) *Exporter {
	counters := make(map[string]*prometheus.CounterVec)
	gauges := make(map[string]*prometheus.GaugeVec)

	for name, info := range counterVecMetrics {
		log.Printf("Registering %s", name)
		counters[name] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      name,
			Help:      info.help,
		}, append([]string{"cluster", "node"}, info.labels...))
	}

	for name, info := range gaugeVecMetrics {
		log.Printf("Registering %s", name)
		gauges[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
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

		counters: counters,
		gauges:   gauges,

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

// Describe describes all the metrics ever exported by the Consul exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up.Desc()

	for _, vec := range e.counters {
		vec.Describe(ch)
	}

	for _, vec := range e.gauges {
		vec.Describe(ch)
	}
}

// Collect fetches the stats from configured Consul location and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	// Reset metrics.
	for _, vec := range e.gauges {
		vec.Reset()
	}

	for _, vec := range e.counters {
		vec.Reset()
	}

	resp, err := e.client.Get(e.URI)
	if err != nil {
		e.up.Set(0)
		log.Printf("Error while querying Elasticsearch: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Failed to read ES response body: %v", err)
		e.up.Set(0)
		return
	}

	e.up.Set(1)

	var all_stats NodeStatsResponse
	err = json.Unmarshal(body, &all_stats)

	if err != nil {
		log.Printf("Failed to unmarshal JSON into struct: %v", err)
		return
	}

	// Regardless of whether we're querying the local host or the whole
	// cluster, here we can just iterate through all nodes found.

	for node, stats := range all_stats.Nodes {
		log.Printf("Processing node %v", node)
		// GC Stats
		for collector, gcstats := range stats.JVM.GC.Collectors {
			e.counters["jvm_gc_collection_count"].WithLabelValues(all_stats.ClusterName, stats.Name, collector).Set(float64(gcstats.CollectionCount))
			e.counters["jvm_gc_collection_time_in_millis"].WithLabelValues(all_stats.ClusterName, stats.Name, collector).Set(float64(gcstats.CollectionTime))
		}

		// Breaker stats
		for breaker, bstats := range stats.Breakers {
			e.gauges["breakers_estimated_size_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name, breaker).Set(float64(bstats.EstimatedSize))
			e.gauges["breakers_limit_size_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name, breaker).Set(float64(bstats.LimitSize))
		}

		// JVM Memory Stats
		e.gauges["jvm_mem_heap_committed_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.JVM.Mem.HeapCommitted))
		e.gauges["jvm_mem_heap_used_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.JVM.Mem.HeapUsed))
		e.gauges["jvm_mem_heap_max_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.JVM.Mem.HeapMax))
		e.gauges["jvm_mem_non_heap_committed_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.JVM.Mem.NonHeapCommitted))
		e.gauges["jvm_mem_non_heap_used_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.JVM.Mem.NonHeapUsed))

		// Indices Stats
		e.gauges["indices_fielddata_evictions"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.FieldData.Evictions))
		e.gauges["indices_fielddata_memory_size_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.FieldData.MemorySize))
		e.gauges["indices_filter_cache_evictions"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.FilterCache.Evictions))
		e.gauges["indices_filter_cache_memory_size_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.FilterCache.MemorySize))

		e.gauges["indices_docs_count"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Docs.Count))
		e.gauges["indices_docs_deleted"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Docs.Deleted))

		e.gauges["indices_segments_memory_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Segments.Memory))

		e.gauges["indices_store_size_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Store.Size))
		e.counters["indices_store_throttle_time_in_millis"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Store.ThrottleTime))

		e.counters["indices_flush_total"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Flush.Total))
		e.counters["indices_flush_time_in_millis"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Flush.Time))

		e.counters["indices_indexing_index_time_in_millis"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Indexing.IndexTime))
		e.counters["indices_indexing_index_total"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Indexing.IndexTotal))

		e.counters["indices_merges_total_time_in_millis"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Merges.TotalTime))
		e.counters["indices_merges_total_size_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Merges.TotalSize))
		e.counters["indices_merges_total"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Indices.Merges.Total))

		// Transport Stats
		e.counters["transport_rx_count"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Transport.RxCount))
		e.counters["transport_rx_size_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Transport.RxSize))
		e.counters["transport_tx_count"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Transport.TxCount))
		e.counters["transport_tx_size_in_bytes"].WithLabelValues(all_stats.ClusterName, stats.Name).Set(float64(stats.Transport.TxSize))
	}

	// Report metrics.
	ch <- e.up

	for _, vec := range e.counters {
		vec.Collect(ch)
	}

	for _, vec := range e.gauges {
		vec.Collect(ch)
	}
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9108", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		esUri         = flag.String("es.uri", "http://localhost:9200", "HTTP API address of a Elasticsearch node.")
		esTimeout     = flag.Duration("es.timeout", 5*time.Second, "Timeout for trying to get stats from Elasticsearch.")
	)
	flag.Parse()

	*esUri = *esUri + "/_nodes/_local/stats"

	exporter := NewExporter(*esUri, *esTimeout)
	prometheus.MustRegister(exporter)

	log.Printf("Starting Server: %s", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Elasticsearch Exporter</title></head>
             <body>
             <h1>Elasticsearch Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
