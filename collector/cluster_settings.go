package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/imdario/mergo"
	"github.com/prometheus/client_golang/prometheus"
)

// ClusterSettings information struct
type ClusterSettings struct {
	logger log.Logger
	client *http.Client
	url    *url.URL

	up                              prometheus.Gauge
	shardAllocationEnabled          prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter
}

// NewClusterSettings defines Cluster Settings Prometheus metrics
func NewClusterSettings(logger log.Logger, client *http.Client, url *url.URL) *ClusterSettings {
	return &ClusterSettings{
		logger: logger,
		client: client,
		url:    url,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "clustersettings_stats", "up"),
			Help: "Was the last scrape of the ElasticSearch cluster settings endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "clustersettings_stats", "total_scrapes"),
			Help: "Current total ElasticSearch cluster settings scrapes.",
		}),
		shardAllocationEnabled: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "clustersettings_stats", "shard_allocation_enabled"),
			Help: "Current mode of cluster wide shard routing allocation settings.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "clustersettings_stats", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
	}
}

// Describe add Snapshots metrics descriptions
func (cs *ClusterSettings) Describe(ch chan<- *prometheus.Desc) {
	ch <- cs.up.Desc()
	ch <- cs.totalScrapes.Desc()
	ch <- cs.shardAllocationEnabled.Desc()
	ch <- cs.jsonParseFailures.Desc()
}

func (cs *ClusterSettings) getAndParseURL(u *url.URL, data interface{}) error {
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

func (cs *ClusterSettings) fetchAndDecodeClusterSettingsStats() (ClusterSettingsResponse, error) {

	u := *cs.url
	u.Path = path.Join(u.Path, "/_cluster/settings")
	q := u.Query()
	q.Set("include_defaults", "true")
	u.RawPath = q.Encode()
	var csfr ClusterSettingsFullResponse
	var csr ClusterSettingsResponse
	err := cs.getAndParseURL(&u, &csfr)
	if err != nil {
		return csr, err
	}
	err = mergo.Merge(&csr, csfr.Defaults, mergo.WithOverride)
	if err != nil {
		return csr, err
	}
	err = mergo.Merge(&csr, csfr.Persistent, mergo.WithOverride)
	if err != nil {
		return csr, err
	}
	err = mergo.Merge(&csr, csfr.Transient, mergo.WithOverride)

	return csr, err
}

// Collect gets cluster settings  metric values
func (cs *ClusterSettings) Collect(ch chan<- prometheus.Metric) {

	cs.totalScrapes.Inc()
	defer func() {
		ch <- cs.up
		ch <- cs.totalScrapes
		ch <- cs.jsonParseFailures
		ch <- cs.shardAllocationEnabled
	}()

	csr, err := cs.fetchAndDecodeClusterSettingsStats()
	if err != nil {
		cs.shardAllocationEnabled.Set(0)
		cs.up.Set(0)
		_ = level.Warn(cs.logger).Log(
			"msg", "failed to fetch and decode cluster settings stats",
			"err", err,
		)
		return
	}
	cs.up.Set(1)

	shardAllocationMap := map[string]int{
		"all":           0,
		"primaries":     1,
		"new_primaries": 2,
		"none":          3,
	}

	cs.shardAllocationEnabled.Set(float64(shardAllocationMap[csr.Cluster.Routing.Allocation.Enabled]))
}
