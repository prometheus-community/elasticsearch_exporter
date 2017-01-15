package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "elasticsearch"
	indexHTML = `
	<html>
		<head>
			<title>Elasticsearch Exporter</title>
		</head>
		<body>
			<h1>Elasticsearch Exporter</h1>
			<p>
			<a href='%s'>Metrics</a>
			</p>
		</body>
	</html>`
)

type VecInfo struct {
	help   string
	labels []string
}

var (
	gaugeMetrics = map[string]string{
		"indices_fielddata_memory_size_bytes":     "Field data cache memory usage in bytes",
		"indices_filter_cache_memory_size_bytes":  "Filter cache memory usage in bytes",
		"indices_query_cache_memory_size_bytes":   "Query cache memory usage in bytes",
		"indices_request_cache_memory_size_bytes": "Request cache memory usage in bytes",
		"indices_docs":                            "Count of documents on this node",
		"indices_docs_deleted":                    "Count of deleted documents on this node",
		"indices_store_size_bytes":                "Current size of stored index data in bytes",
		"indices_segments_memory_bytes":           "Current memory size of segments in bytes",
		"indices_segments_count":                  "Count of index segments on this node",
		"process_cpu_percent":                     "Percent CPU used by process",
		"process_mem_resident_size_bytes":         "Resident memory in use by process in bytes",
		"process_mem_share_size_bytes":            "Shared memory in use by process in bytes",
		"process_mem_virtual_size_bytes":          "Total virtual memory used in bytes",
		"process_open_files_count":                "Open file descriptors",
		"process_max_files_count":                 "Max file descriptors for process",
	}
	counterMetrics = map[string]string{
		"indices_fielddata_evictions":           "Evictions from field data",
		"indices_filter_cache_evictions":        "Evictions from filter cache",
		"indices_query_cache_evictions":         "Evictions from query cache",
		"indices_request_cache_evictions":       "Evictions from request cache",
		"indices_flush_total":                   "Total flushes",
		"indices_flush_time_ms_total":           "Cumulative flush time in milliseconds",
		"transport_rx_packets_total":            "Count of packets received",
		"transport_rx_size_bytes_total":         "Total number of bytes received",
		"transport_tx_packets_total":            "Count of packets sent",
		"transport_tx_size_bytes_total":         "Total number of bytes sent",
		"indices_store_throttle_time_ms_total":  "Throttle time for index store in milliseconds",
		"indices_indexing_index_total":          "Total index calls",
		"indices_indexing_index_time_ms_total":  "Cumulative index time in milliseconds",
		"indices_merges_total":                  "Total merges",
		"indices_merges_total_docs_total":       "Cumulative docs merged",
		"indices_merges_total_size_bytes_total": "Total merge size in bytes",
		"indices_merges_total_time_ms_total":    "Total time spent merging in milliseconds",
		"indices_refresh_total":                 "Total refreshes",
		"indices_refresh_total_time_ms_total":   "Total time spent refreshing",
	}
	counterVecMetrics = map[string]*VecInfo{
		"jvm_gc_collection_seconds_count": &VecInfo{
			help:   "Count of JVM GC runs",
			labels: []string{"gc"},
		},
		"jvm_gc_collection_seconds_sum": &VecInfo{
			help:   "GC run time in seconds",
			labels: []string{"gc"},
		},
		"process_cpu_time_seconds_sum": &VecInfo{
			help:   "Process CPU time in seconds",
			labels: []string{"type"},
		},
		"thread_pool_completed_count": &VecInfo{
			help:   "Thread Pool operations completed",
			labels: []string{"type"},
		},
		"thread_pool_rejected_count": &VecInfo{
			help:   "Thread Pool operations rejected",
			labels: []string{"type"},
		},
	}

	gaugeVecMetrics = map[string]*VecInfo{
		"breakers_estimated_size_bytes": &VecInfo{
			help:   "Estimated size in bytes of breaker",
			labels: []string{"breaker"},
		},
		"breakers_limit_size_bytes": &VecInfo{
			help:   "Limit size in bytes for breaker",
			labels: []string{"breaker"},
		},
		"jvm_memory_committed_bytes": &VecInfo{
			help:   "JVM memory currently committed by area",
			labels: []string{"area"},
		},
		"jvm_memory_used_bytes": &VecInfo{
			help:   "JVM memory currently used by area",
			labels: []string{"area"},
		},
		"jvm_memory_max_bytes": &VecInfo{
			help:   "JVM memory max",
			labels: []string{"area"},
		},
		"thread_pool_active_count": &VecInfo{
			help:   "Thread Pool threads active",
			labels: []string{"type"},
		},
		"thread_pool_largest_count": &VecInfo{
			help:   "Thread Pool largest threads count",
			labels: []string{"type"},
		},
		"thread_pool_queue_count": &VecInfo{
			help:   "Thread Pool operations queued",
			labels: []string{"type"},
		},
		"thread_pool_threads_count": &VecInfo{
			help:   "Thread Pool current threads count",
			labels: []string{"type"},
		},
	}
)

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9108", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		esURI         = flag.String("es.uri", "http://localhost:9200", "HTTP API address of an Elasticsearch node.")
		esTimeout     = flag.Duration("es.timeout", 5*time.Second, "Timeout for trying to get stats from Elasticsearch.")
		esAllNodes    = flag.Bool("es.all", false, "Export stats for all nodes in the cluster.")
	)
	flag.Parse()

	if *esAllNodes {
		*esURI = *esURI + "/_nodes/stats"
	} else {
		*esURI = *esURI + "/_nodes/_local/stats"
	}

	exporter := NewExporter(*esURI, *esTimeout, *esAllNodes)
	prometheus.MustRegister(exporter)

	log.Println("Starting Server:", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(indexHTML, *metricsPath)))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
