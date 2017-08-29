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
	defaultAliasLabels      = []string{"alias", "index"}
	defaultAliasLabelValues = func(alias string, indexName string) []string {
		return []string{alias, indexName}
	}
)

type aliasMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(indexStats IndexStatsIndexResponse) float64
	Labels func(alias string, indexName string) []string
}

type Aliases struct {
	logger log.Logger
	client *http.Client
	url    *url.URL

	up                prometheus.Gauge
	totalScrapes      prometheus.Counter
	jsonParseFailures prometheus.Counter

	aliasMetrics []*aliasMetric
}

func NewAliases(logger log.Logger, client *http.Client, url *url.URL) *Aliases {
	return &Aliases{
		logger: logger,
		client: client,
		url:    url,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "alias_stats", "up"),
			Help: "Was the last scrape of the ElasticSearch alias endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "alias_stats", "total_scrapes"),
			Help: "Current total ElasticSearch alias scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "alias_stats", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),

		aliasMetrics: []*aliasMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "aliases", "docs_primary"),
					"Count of documents which only primary shards via alias",
					defaultAliasLabels, nil,
				),
				Value: func(indexStats IndexStatsIndexResponse) float64 {
					return float64(indexStats.Primaries.Docs.Count)
				},
				Labels: defaultAliasLabelValues,
			},
		},
	}
}

func (i *Aliases) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range i.aliasMetrics {
		ch <- metric.Desc
	}
	ch <- i.up.Desc()
	ch <- i.totalScrapes.Desc()
	ch <- i.jsonParseFailures.Desc()
}

func (c *Aliases) fetchAndDecodeAliasStats() (aliasesResponse, error) {
	var ar aliasesResponse

	u := *c.url
	u.Path = "/_aliases"

	res, err := c.client.Get(u.String())
	if err != nil {
		return ar, fmt.Errorf("failed to get alias from %s://%s:%s/%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ar, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&ar); err != nil {
		c.jsonParseFailures.Inc()
		return ar, err
	}
	return ar, nil
}

func (a *Aliases) Collect(ch chan<- prometheus.Metric) {
	a.totalScrapes.Inc()
	defer func() {
		ch <- a.up
		ch <- a.totalScrapes
		ch <- a.jsonParseFailures
	}()

	// indices
	indices := NewIndices(a.logger, a.client, a.url)
	indexStatsResponse, err := indices.fetchAndDecodeIndexStats()
	if err != nil {
		a.up.Set(0)
		level.Warn(a.logger).Log(
			"msg", "failed to fetch and decode index stats",
			"err", err,
		)
		return
	}

	// aliases
	aliasesResponse, err := a.fetchAndDecodeAliasStats()
	if err != nil {
		a.up.Set(0)
		level.Warn(a.logger).Log(
			"msg", "failed to fetch and decode alias",
			"err", err,
		)
		return
	}

	a.up.Set(1)

	for indexName, aliases := range aliasesResponse {
		for alias := range aliases["aliases"] {
			for _, metric := range a.aliasMetrics {
				indexStatsIndexResponse := indexStatsResponse.Indices[indexName]
				ch <- prometheus.MustNewConstMetric(
					metric.Desc,
					metric.Type,
					metric.Value(indexStatsIndexResponse),
					metric.Labels(alias, indexName)...,
				)
			}
		}
	}
}
