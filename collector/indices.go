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

func (c *Indices) fetchAndDecodeIndexStats() (indexStatsResponse, error) {
	var isr indexStatsResponse

	u := *c.url
	u.Path = "/_all/_stats"

	res, err := c.client.Get(u.String())
	if err != nil {
		return isr, fmt.Errorf("failed to get index stats from %s://%s:%s/%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return isr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&isr); err != nil {
		c.jsonParseFailures.Inc()
		return isr, err
	}
	return isr, nil
}

func (i *Indices) Collect(ch chan<- prometheus.Metric) {
	i.totalScrapes.Inc()
	defer func() {
		ch <- i.up
		ch <- i.totalScrapes
		ch <- i.jsonParseFailures
	}()

	// clusterHealth
	clusterHealth := NewClusterHealth(i.logger, i.client, i.url)
	clusterHealthResponse, err := clusterHealth.fetchAndDecodeClusterHealth()
	if err != nil {
		i.up.Set(0)
		level.Warn(i.logger).Log(
			"msg", "failed to fetch and decode cluster health",
			"err", err,
		)
		return
	}

	// indices
	indexStatsResponse, err := i.fetchAndDecodeIndexStats()
	if err != nil {
		i.up.Set(0)
		level.Warn(i.logger).Log(
			"msg", "failed to fetch and decode index stats",
			"err", err,
		)
		return
	}
	i.up.Set(1)

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
