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

type (
	NodeStatsResponse struct {
		ClusterName string `json:"cluster_name"`
		Nodes       map[string]NodeStatsNodeResponse
	}

	NodeStatsNodeResponse struct {
		Timestamp        int64    `json:"timestamp"`
		Name             string   `json:"name"`
		TransportAddress string   `json:"transport_address"`
		Host             string   `json:"host"`
		IP               string   `json:"ip"`
		Roles            []string `json:"roles"`
		//	Indices          struct {
		//		Docs struct {
		//			Count   int `json:"count"`
		//			Deleted int `json:"deleted"`
		//		} `json:"docs"`
		//		Store struct {
		//			SizeInBytes          int64 `json:"size_in_bytes"`
		//			ThrottleTimeInMillis int   `json:"throttle_time_in_millis"`
		//		} `json:"store"`
		//		Indexing struct {
		//			IndexTotal           int  `json:"index_total"`
		//			IndexTimeInMillis    int  `json:"index_time_in_millis"`
		//			IndexCurrent         int  `json:"index_current"`
		//			IndexFailed          int  `json:"index_failed"`
		//			DeleteTotal          int  `json:"delete_total"`
		//			DeleteTimeInMillis   int  `json:"delete_time_in_millis"`
		//			DeleteCurrent        int  `json:"delete_current"`
		//			NoopUpdateTotal      int  `json:"noop_update_total"`
		//			IsThrottled          bool `json:"is_throttled"`
		//			ThrottleTimeInMillis int  `json:"throttle_time_in_millis"`
		//		} `json:"indexing"`
		//		Get struct {
		//			Total               int `json:"total"`
		//			TimeInMillis        int `json:"time_in_millis"`
		//			ExistsTotal         int `json:"exists_total"`
		//			ExistsTimeInMillis  int `json:"exists_time_in_millis"`
		//			MissingTotal        int `json:"missing_total"`
		//			MissingTimeInMillis int `json:"missing_time_in_millis"`
		//			Current             int `json:"current"`
		//		} `json:"get"`
		//		Search struct {
		//			OpenContexts        int `json:"open_contexts"`
		//			QueryTotal          int `json:"query_total"`
		//			QueryTimeInMillis   int `json:"query_time_in_millis"`
		//			QueryCurrent        int `json:"query_current"`
		//			FetchTotal          int `json:"fetch_total"`
		//			FetchTimeInMillis   int `json:"fetch_time_in_millis"`
		//			FetchCurrent        int `json:"fetch_current"`
		//			ScrollTotal         int `json:"scroll_total"`
		//			ScrollTimeInMillis  int `json:"scroll_time_in_millis"`
		//			ScrollCurrent       int `json:"scroll_current"`
		//			SuggestTotal        int `json:"suggest_total"`
		//			SuggestTimeInMillis int `json:"suggest_time_in_millis"`
		//			SuggestCurrent      int `json:"suggest_current"`
		//		} `json:"search"`
		//		Merges struct {
		//			Current                    int   `json:"current"`
		//			CurrentDocs                int   `json:"current_docs"`
		//			CurrentSizeInBytes         int   `json:"current_size_in_bytes"`
		//			Total                      int   `json:"total"`
		//			TotalTimeInMillis          int   `json:"total_time_in_millis"`
		//			TotalDocs                  int   `json:"total_docs"`
		//			TotalSizeInBytes           int64 `json:"total_size_in_bytes"`
		//			TotalStoppedTimeInMillis   int   `json:"total_stopped_time_in_millis"`
		//			TotalThrottledTimeInMillis int   `json:"total_throttled_time_in_millis"`
		//			TotalAutoThrottleInBytes   int   `json:"total_auto_throttle_in_bytes"`
		//		} `json:"merges"`
		//		Refresh struct {
		//			Total             int `json:"total"`
		//			TotalTimeInMillis int `json:"total_time_in_millis"`
		//			Listeners         int `json:"listeners"`
		//		} `json:"refresh"`
		//		Flush struct {
		//			Total             int `json:"total"`
		//			TotalTimeInMillis int `json:"total_time_in_millis"`
		//		} `json:"flush"`
		//		Warmer struct {
		//			Current           int `json:"current"`
		//			Total             int `json:"total"`
		//			TotalTimeInMillis int `json:"total_time_in_millis"`
		//		} `json:"warmer"`
		//		QueryCache struct {
		//			MemorySizeInBytes int `json:"memory_size_in_bytes"`
		//			TotalCount        int `json:"total_count"`
		//			HitCount          int `json:"hit_count"`
		//			MissCount         int `json:"miss_count"`
		//			CacheSize         int `json:"cache_size"`
		//			CacheCount        int `json:"cache_count"`
		//			Evictions         int `json:"evictions"`
		//		} `json:"query_cache"`
		//		Fielddata struct {
		//			MemorySizeInBytes int `json:"memory_size_in_bytes"`
		//			Evictions         int `json:"evictions"`
		//		} `json:"fielddata"`
		//		Completion struct {
		//			SizeInBytes int `json:"size_in_bytes"`
		//		} `json:"completion"`
		//		Segments struct {
		//			Count                     int `json:"count"`
		//			MemoryInBytes             int `json:"memory_in_bytes"`
		//			TermsMemoryInBytes        int `json:"terms_memory_in_bytes"`
		//			StoredFieldsMemoryInBytes int `json:"stored_fields_memory_in_bytes"`
		//			TermVectorsMemoryInBytes  int `json:"term_vectors_memory_in_bytes"`
		//			NormsMemoryInBytes        int `json:"norms_memory_in_bytes"`
		//			PointsMemoryInBytes       int `json:"points_memory_in_bytes"`
		//			DocValuesMemoryInBytes    int `json:"doc_values_memory_in_bytes"`
		//			IndexWriterMemoryInBytes  int `json:"index_writer_memory_in_bytes"`
		//			VersionMapMemoryInBytes   int `json:"version_map_memory_in_bytes"`
		//			FixedBitSetMemoryInBytes  int `json:"fixed_bit_set_memory_in_bytes"`
		//			MaxUnsafeAutoIDTimestamp  int `json:"max_unsafe_auto_id_timestamp"`
		//			FileSizes                 struct {
		//			} `json:"file_sizes"`
		//		} `json:"segments"`
		//		Translog struct {
		//			Operations  int `json:"operations"`
		//			SizeInBytes int `json:"size_in_bytes"`
		//		} `json:"translog"`
		//		RequestCache struct {
		//			MemorySizeInBytes int `json:"memory_size_in_bytes"`
		//			Evictions         int `json:"evictions"`
		//			HitCount          int `json:"hit_count"`
		//			MissCount         int `json:"miss_count"`
		//		} `json:"request_cache"`
		//		Recovery struct {
		//			CurrentAsSource      int `json:"current_as_source"`
		//			CurrentAsTarget      int `json:"current_as_target"`
		//			ThrottleTimeInMillis int `json:"throttle_time_in_millis"`
		//		} `json:"recovery"`
		//	} `json:"indices"`
		//	Os struct {
		//		Timestamp int64 `json:"timestamp"`
		//		CPU       struct {
		//			Percent     int `json:"percent"`
		//			LoadAverage struct {
		//				OneM  float64 `json:"1m"`
		//				FiveM float64 `json:"5m"`
		//				One5M float64 `json:"15m"`
		//			} `json:"load_average"`
		//		} `json:"cpu"`
		//		Mem struct {
		//			TotalInBytes int64 `json:"total_in_bytes"`
		//			FreeInBytes  int   `json:"free_in_bytes"`
		//			UsedInBytes  int64 `json:"used_in_bytes"`
		//			FreePercent  int   `json:"free_percent"`
		//			UsedPercent  int   `json:"used_percent"`
		//		} `json:"mem"`
		//		Swap struct {
		//			TotalInBytes int `json:"total_in_bytes"`
		//			FreeInBytes  int `json:"free_in_bytes"`
		//			UsedInBytes  int `json:"used_in_bytes"`
		//		} `json:"swap"`
		//		Cgroup struct {
		//			Cpuacct struct {
		//				ControlGroup string `json:"control_group"`
		//				UsageNanos   int64  `json:"usage_nanos"`
		//			} `json:"cpuacct"`
		//			CPU struct {
		//				ControlGroup    string `json:"control_group"`
		//				CfsPeriodMicros int    `json:"cfs_period_micros"`
		//				CfsQuotaMicros  int    `json:"cfs_quota_micros"`
		//				Stat            struct {
		//					NumberOfElapsedPeriods int `json:"number_of_elapsed_periods"`
		//					NumberOfTimesThrottled int `json:"number_of_times_throttled"`
		//					TimeThrottledNanos     int `json:"time_throttled_nanos"`
		//				} `json:"stat"`
		//			} `json:"cpu"`
		//		} `json:"cgroup"`
		//	} `json:"os"`
		//	Process struct {
		//		Timestamp           int64 `json:"timestamp"`
		//		OpenFileDescriptors int   `json:"open_file_descriptors"`
		//		MaxFileDescriptors  int   `json:"max_file_descriptors"`
		//		CPU                 struct {
		//			Percent       int `json:"percent"`
		//			TotalInMillis int `json:"total_in_millis"`
		//		} `json:"cpu"`
		//		Mem struct {
		//			TotalVirtualInBytes int64 `json:"total_virtual_in_bytes"`
		//		} `json:"mem"`
		//	} `json:"process"`
		//	Jvm struct {
		//		Timestamp      int64 `json:"timestamp"`
		//		UptimeInMillis int   `json:"uptime_in_millis"`
		//		Mem            struct {
		//			HeapUsedInBytes         int64 `json:"heap_used_in_bytes"`
		//			HeapUsedPercent         int   `json:"heap_used_percent"`
		//			HeapCommittedInBytes    int64 `json:"heap_committed_in_bytes"`
		//			HeapMaxInBytes          int64 `json:"heap_max_in_bytes"`
		//			NonHeapUsedInBytes      int   `json:"non_heap_used_in_bytes"`
		//			NonHeapCommittedInBytes int   `json:"non_heap_committed_in_bytes"`
		//			Pools                   struct {
		//				Young struct {
		//					UsedInBytes     int `json:"used_in_bytes"`
		//					MaxInBytes      int `json:"max_in_bytes"`
		//					PeakUsedInBytes int `json:"peak_used_in_bytes"`
		//					PeakMaxInBytes  int `json:"peak_max_in_bytes"`
		//				} `json:"young"`
		//				Survivor struct {
		//					UsedInBytes     int `json:"used_in_bytes"`
		//					MaxInBytes      int `json:"max_in_bytes"`
		//					PeakUsedInBytes int `json:"peak_used_in_bytes"`
		//					PeakMaxInBytes  int `json:"peak_max_in_bytes"`
		//				} `json:"survivor"`
		//				Old struct {
		//					UsedInBytes     int64 `json:"used_in_bytes"`
		//					MaxInBytes      int64 `json:"max_in_bytes"`
		//					PeakUsedInBytes int64 `json:"peak_used_in_bytes"`
		//					PeakMaxInBytes  int64 `json:"peak_max_in_bytes"`
		//				} `json:"old"`
		//			} `json:"pools"`
		//		} `json:"mem"`
		//		Threads struct {
		//			Count     int `json:"count"`
		//			PeakCount int `json:"peak_count"`
		//		} `json:"threads"`
		//		Gc struct {
		//			Collectors struct {
		//				Young struct {
		//					CollectionCount        int `json:"collection_count"`
		//					CollectionTimeInMillis int `json:"collection_time_in_millis"`
		//				} `json:"young"`
		//				Old struct {
		//					CollectionCount        int `json:"collection_count"`
		//					CollectionTimeInMillis int `json:"collection_time_in_millis"`
		//				} `json:"old"`
		//			} `json:"collectors"`
		//		} `json:"gc"`
		//		BufferPools struct {
		//			Direct struct {
		//				Count                int `json:"count"`
		//				UsedInBytes          int `json:"used_in_bytes"`
		//				TotalCapacityInBytes int `json:"total_capacity_in_bytes"`
		//			} `json:"direct"`
		//			Mapped struct {
		//				Count                int   `json:"count"`
		//				UsedInBytes          int64 `json:"used_in_bytes"`
		//				TotalCapacityInBytes int64 `json:"total_capacity_in_bytes"`
		//			} `json:"mapped"`
		//		} `json:"buffer_pools"`
		//		Classes struct {
		//			CurrentLoadedCount int `json:"current_loaded_count"`
		//			TotalLoadedCount   int `json:"total_loaded_count"`
		//			TotalUnloadedCount int `json:"total_unloaded_count"`
		//		} `json:"classes"`
		//	} `json:"jvm"`
		//	ThreadPool struct {
		//		Bulk struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"bulk"`
		//		FetchShardStarted struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"fetch_shard_started"`
		//		FetchShardStore struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"fetch_shard_store"`
		//		Flush struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"flush"`
		//		ForceMerge struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"force_merge"`
		//		Generic struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"generic"`
		//		Get struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"get"`
		//		Index struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"index"`
		//		Listener struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"listener"`
		//		Management struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"management"`
		//		Refresh struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"refresh"`
		//		Search struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"search"`
		//		Snapshot struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"snapshot"`
		//		Warmer struct {
		//			Threads   int `json:"threads"`
		//			Queue     int `json:"queue"`
		//			Active    int `json:"active"`
		//			Rejected  int `json:"rejected"`
		//			Largest   int `json:"largest"`
		//			Completed int `json:"completed"`
		//		} `json:"warmer"`
		//	} `json:"thread_pool"`
		//	Fs struct {
		//		Timestamp int64 `json:"timestamp"`
		//		Total     struct {
		//			TotalInBytes     int64 `json:"total_in_bytes"`
		//			FreeInBytes      int64 `json:"free_in_bytes"`
		//			AvailableInBytes int64 `json:"available_in_bytes"`
		//		} `json:"total"`
		//		Data []struct {
		//			Path             string `json:"path"`
		//			Mount            string `json:"mount"`
		//			Type             string `json:"type"`
		//			TotalInBytes     int64  `json:"total_in_bytes"`
		//			FreeInBytes      int64  `json:"free_in_bytes"`
		//			AvailableInBytes int64  `json:"available_in_bytes"`
		//			Spins            string `json:"spins"`
		//		} `json:"data"`
		//		IoStats struct {
		//			Devices []struct {
		//				DeviceName      string `json:"device_name"`
		//				Operations      int    `json:"operations"`
		//				ReadOperations  int    `json:"read_operations"`
		//				WriteOperations int    `json:"write_operations"`
		//				ReadKilobytes   int    `json:"read_kilobytes"`
		//				WriteKilobytes  int    `json:"write_kilobytes"`
		//			} `json:"devices"`
		//			Total struct {
		//				Operations      int `json:"operations"`
		//				ReadOperations  int `json:"read_operations"`
		//				WriteOperations int `json:"write_operations"`
		//				ReadKilobytes   int `json:"read_kilobytes"`
		//				WriteKilobytes  int `json:"write_kilobytes"`
		//			} `json:"total"`
		//		} `json:"io_stats"`
		//	} `json:"fs"`
		//	Transport struct {
		//		ServerOpen    int   `json:"server_open"`
		//		RxCount       int   `json:"rx_count"`
		//		RxSizeInBytes int64 `json:"rx_size_in_bytes"`
		//		TxCount       int   `json:"tx_count"`
		//		TxSizeInBytes int64 `json:"tx_size_in_bytes"`
		//	} `json:"transport"`
		//	HTTP struct {
		//		CurrentOpen int `json:"current_open"`
		//		TotalOpened int `json:"total_opened"`
		//	} `json:"http"`
		//	Breakers struct {
		//		Request struct {
		//			LimitSizeInBytes     int     `json:"limit_size_in_bytes"`
		//			LimitSize            string  `json:"limit_size"`
		//			EstimatedSizeInBytes int     `json:"estimated_size_in_bytes"`
		//			EstimatedSize        string  `json:"estimated_size"`
		//			Overhead             float64 `json:"overhead"`
		//			Tripped              int     `json:"tripped"`
		//		} `json:"request"`
		//		Fielddata struct {
		//			LimitSizeInBytes     int     `json:"limit_size_in_bytes"`
		//			LimitSize            string  `json:"limit_size"`
		//			EstimatedSizeInBytes int     `json:"estimated_size_in_bytes"`
		//			EstimatedSize        string  `json:"estimated_size"`
		//			Overhead             float64 `json:"overhead"`
		//			Tripped              int     `json:"tripped"`
		//		} `json:"fielddata"`
		//		InFlightRequests struct {
		//			LimitSizeInBytes     int64   `json:"limit_size_in_bytes"`
		//			LimitSize            string  `json:"limit_size"`
		//			EstimatedSizeInBytes int     `json:"estimated_size_in_bytes"`
		//			EstimatedSize        string  `json:"estimated_size"`
		//			Overhead             float64 `json:"overhead"`
		//			Tripped              int     `json:"tripped"`
		//		} `json:"in_flight_requests"`
		//		Parent struct {
		//			LimitSizeInBytes     int     `json:"limit_size_in_bytes"`
		//			LimitSize            string  `json:"limit_size"`
		//			EstimatedSizeInBytes int     `json:"estimated_size_in_bytes"`
		//			EstimatedSize        string  `json:"estimated_size"`
		//			Overhead             float64 `json:"overhead"`
		//			Tripped              int     `json:"tripped"`
		//		} `json:"parent"`
		//	} `json:"breakers"`
		//	Script struct {
		//		Compilations   int `json:"compilations"`
		//		CacheEvictions int `json:"cache_evictions"`
		//	} `json:"script"`
		//	Discovery struct {
		//		ClusterStateQueue struct {
		//			Total     int `json:"total"`
		//			Pending   int `json:"pending"`
		//			Committed int `json:"committed"`
		//		} `json:"cluster_state_queue"`
		//	} `json:"discovery"`
		//	Ingest struct {
		//		Total struct {
		//			Count        int `json:"count"`
		//			TimeInMillis int `json:"time_in_millis"`
		//			Current      int `json:"current"`
		//			Failed       int `json:"failed"`
		//		} `json:"total"`
		//		Pipelines struct {
		//			XpackMonitoring2 struct {
		//				Count        int `json:"count"`
		//				TimeInMillis int `json:"time_in_millis"`
		//				Current      int `json:"current"`
		//				Failed       int `json:"failed"`
		//			} `json:"xpack_monitoring_2"`
		//		} `json:"pipelines"`
		//	} `json:"ingest"`
	}
)

type Nodes struct {
	logger log.Logger
	client *http.Client
	url    url.URL
	all    bool

	foobar *prometheus.Desc
}

func NewNodes(logger log.Logger, client *http.Client, url url.URL, all bool) *Nodes {
	return &Nodes{
		logger: logger,
		client: client,
		url:    url,
		all:    all,

		foobar: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "foo", "bar"),
			"bla bla bla.",
			nil, nil,
		),
	}
}

func (c *Nodes) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.foobar
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
		fmt.Printf("host: %+v\n", node.Host)
	}
}
