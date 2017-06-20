package collector

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "elasticsearch"
)

type ClusterHealth struct {
	logger log.Logger
	client *http.Client
	url    *url.URL

	ActivePrimaryShards     *prometheus.Desc
	ActiveShards            *prometheus.Desc
	DelayedUnassignedShards *prometheus.Desc
	InitializingShards      *prometheus.Desc
	NumberOfDataNodes       *prometheus.Desc
	NumberOfInFlightFetch   *prometheus.Desc
	NumberOfNodes           *prometheus.Desc
	NumberOfPendingTasks    *prometheus.Desc
	RelocatingShards        *prometheus.Desc
	StatusIsGreen           *prometheus.Desc
	Status                  *prometheus.Desc
	StatusIsYellow          *prometheus.Desc
	StatusIsRed             *prometheus.Desc
	TimedOut                *prometheus.Desc
	UnassignedShards        *prometheus.Desc
}

// Generated with https://mholt.github.io/json-to-go/
type clusterHealthResponse struct {
	ClusterName                 string  `json:"cluster_name"`
	Status                      string  `json:"status"`
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}

func NewClusterHealth(logger log.Logger, client *http.Client, url *url.URL) *ClusterHealth {
	subsystem := "cluster_health"

	return &ClusterHealth{
		logger: logger,
		client: client,
		url:    url,

		ActivePrimaryShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "active_primary_shards"),
			"Tthe number of primary shards in your cluster. This is an aggregate total across all indices.",
			[]string{"cluster"}, nil,
		),
		ActiveShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "active_shards"),
			"Aggregate total of all shards across all indices, which includes replica shards.",
			[]string{"cluster"}, nil,
		),
		DelayedUnassignedShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "delayed_unassigned_shards"),
			"XXX WHAT DOES THIS MEAN?",
			[]string{"cluster"}, nil,
		),
		InitializingShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "initializing_shards"),
			"Count of shards that are being freshly created.",
			[]string{"cluster"}, nil,
		),
		NumberOfDataNodes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "number_of_data_nodes"),
			"Number of data nodes in the cluster.",
			[]string{"cluster"}, nil,
		),
		NumberOfInFlightFetch: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "number_of_in_flight_fetch"),
			"The number of ongoing shard info requests.",
			[]string{"cluster"}, nil,
		),
		NumberOfNodes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "number_of_nodes"),
			"Number of nodes in the cluster.",
			[]string{"cluster"}, nil,
		),
		NumberOfPendingTasks: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "number_of_pending_tasks"),
			"XXX WHAT DOES THIS MEAN?",
			[]string{"cluster"}, nil,
		),
		RelocatingShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "relocating_shards"),
			"The number of shards that are currently moving from one node to another node.",
			[]string{"cluster"}, nil,
		),
		StatusIsGreen: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "status_is_green"),
			"Whether all primary and replica shards are allocated.",
			[]string{"cluster"}, nil,
		),
		Status: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "status"),
			"Whether all primary and replica shards are allocated.",
			[]string{"cluster", "color"}, nil,
		),
		StatusIsYellow: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "status_is_yellow"),
			"Whether all primary and replica shards are allocated.",
			[]string{"cluster"}, nil,
		),
		StatusIsRed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "status_is_red"),
			"Whether all primary and replica shards are allocated.",
			[]string{"cluster"}, nil,
		),
		TimedOut: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "timed_out"),
			"XXX WHAT DOES THIS MEAN?",
			[]string{"cluster"}, nil,
		),
		UnassignedShards: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "unassigned_shards"),
			"The number of shards that exist in the cluster state, but cannot be found in the cluster itself.",
			[]string{"cluster"}, nil,
		),
	}
}

func (c *ClusterHealth) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Status
}

func (c *ClusterHealth) Collect(ch chan<- prometheus.Metric) {
	c.url.Path = "/_cluster/health"
	res, err := c.client.Get(c.url.String())
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "failed to get cluster health",
			"url", c.url.String(),
			"err", err,
		)
		return
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)

	var clusterHealthResponse clusterHealthResponse
	if err := dec.Decode(&clusterHealthResponse); err != nil {
		level.Warn(c.logger).Log(
			"msg", "failed to decode cluster health",
			"err", err,
		)
		return
	}

	cluster := clusterHealthResponse.ClusterName

	ch <- prometheus.MustNewConstMetric(
		c.ActivePrimaryShards,
		prometheus.GaugeValue,
		float64(clusterHealthResponse.ActivePrimaryShards),
		cluster,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ActiveShards,
		prometheus.GaugeValue,
		float64(clusterHealthResponse.ActiveShards),
		cluster,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DelayedUnassignedShards,
		prometheus.GaugeValue,
		float64(clusterHealthResponse.DelayedUnassignedShards),
		cluster,
	)

	ch <- prometheus.MustNewConstMetric(
		c.InitializingShards,
		prometheus.GaugeValue,
		float64(clusterHealthResponse.InitializingShards),
		cluster,
	)

	ch <- prometheus.MustNewConstMetric(
		c.NumberOfDataNodes,
		prometheus.GaugeValue,
		float64(clusterHealthResponse.NumberOfDataNodes),
		cluster,
	)

	ch <- prometheus.MustNewConstMetric(
		c.NumberOfInFlightFetch,
		prometheus.GaugeValue,
		float64(clusterHealthResponse.NumberOfInFlightFetch),
		cluster,
	)

	ch <- prometheus.MustNewConstMetric(
		c.NumberOfNodes,
		prometheus.GaugeValue,
		float64(clusterHealthResponse.NumberOfNodes),
		cluster,
	)

	ch <- prometheus.MustNewConstMetric(
		c.NumberOfPendingTasks,
		prometheus.GaugeValue,
		float64(clusterHealthResponse.NumberOfPendingTasks),
		cluster,
	)

	ch <- prometheus.MustNewConstMetric(
		c.RelocatingShards,
		prometheus.GaugeValue,
		float64(clusterHealthResponse.RelocatingShards),
		cluster,
	)

	var statusValue float64
	if clusterHealthResponse.Status == "green" {
		statusValue = 1.0
	}
	ch <- prometheus.MustNewConstMetric(
		c.Status,
		prometheus.GaugeValue,
		statusValue,
		cluster, clusterHealthResponse.Status,
	)

	var timedOut float64
	if clusterHealthResponse.TimedOut {
		timedOut = 1.0
	}
	ch <- prometheus.MustNewConstMetric(
		c.TimedOut,
		prometheus.GaugeValue,
		timedOut,
		cluster,
	)

	ch <- prometheus.MustNewConstMetric(
		c.UnassignedShards,
		prometheus.GaugeValue,
		float64(clusterHealthResponse.UnassignedShards),
		cluster,
	)
}
