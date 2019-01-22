package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// IndicesSettings information struct
type IndicesSettings struct {
	logger log.Logger
	client *http.Client
	url    *url.URL

	up                              prometheus.Gauge
	readOnlyIndices                 prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter
}

// NewIndicesSettings defines Indices Settings Prometheus metrics
func NewIndicesSettings(logger log.Logger, client *http.Client, url *url.URL) *IndicesSettings {
	return &IndicesSettings{
		logger: logger,
		client: client,
		url:    url,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "indices_settings_stats", "up"),
			Help: "Was the last scrape of the ElasticSearch Indices Settings endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "indices_settings_stats", "total_scrapes"),
			Help: "Current total ElasticSearch Indices Settings scrapes.",
		}),
		readOnlyIndices: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "indices_settings_stats", "read_only_indices"),
			Help: "Current number of read only indices within cluster",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "indices_settings_stats", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
	}
}

// Describe add Snapshots metrics descriptions
func (cs *IndicesSettings) Describe(ch chan<- *prometheus.Desc) {
	ch <- cs.up.Desc()
	ch <- cs.totalScrapes.Desc()
	ch <- cs.readOnlyIndices.Desc()
	ch <- cs.jsonParseFailures.Desc()
}

func (cs *IndicesSettings) getAndParseURL(u *url.URL, data interface{}) error {
	res, err := cs.client.Get(u.String())
	if err != nil {
		return fmt.Errorf("failed to get from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(cs.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(data); err != nil {
		cs.jsonParseFailures.Inc()
		return err
	}
	return nil
}

func (cs *IndicesSettings) fetchAndDecodeIndicesSettings() (IndicesSettingsResponse, error) {

	u := *cs.url
	u.Path = path.Join(u.Path, "/_all/_settings")
	var asr IndicesSettingsResponse
	err := cs.getAndParseURL(&u, &asr)
	if err != nil {
		return asr, err
	}

	return asr, err
}

// Collect gets all indices settings metric values
func (cs *IndicesSettings) Collect(ch chan<- prometheus.Metric) {

	cs.totalScrapes.Inc()
	defer func() {
		ch <- cs.up
		ch <- cs.totalScrapes
		ch <- cs.jsonParseFailures
		ch <- cs.readOnlyIndices
	}()

	asr, err := cs.fetchAndDecodeIndicesSettings()
	if err != nil {
		cs.readOnlyIndices.Set(0)
		cs.up.Set(0)
		_ = level.Warn(cs.logger).Log(
			"msg", "failed to fetch and decode cluster settings stats",
			"err", err,
		)
		return
	}
	cs.up.Set(1)

	var c int
	for _, value := range asr {
		if value.Settings.IndexInfo.Blocks.ReadOnly == "true" {
			c++
		}
	}
	cs.readOnlyIndices.Set(float64(c))
}
