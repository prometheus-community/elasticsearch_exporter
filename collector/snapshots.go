package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type snapshotMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(snapshotStats SnapshotStatDataResponse) float64
	Labels func(repositoryName string, snapshotStats SnapshotStatDataResponse) []string
}

type repositoryMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(snapshotsStats SnapshotStatsResponse) float64
	Labels func(repositoryName string) []string
}

var (
	defaultSnapshotLabels      = []string{"repository", "state", "version"}
	defaultSnapshotLabelValues = func(repositoryName string, snapshotStats SnapshotStatDataResponse) []string {
		return []string{repositoryName, snapshotStats.State, snapshotStats.Version}
	}
	defaultSnapshotRepositoryLabels      = []string{"repository"}
	defaultSnapshotRepositoryLabelValues = func(repositoryName string) []string {
		return []string{repositoryName}
	}
)

// Snapshots information struct
type Snapshots struct {
	logger  log.Logger
	updater *snapshotsUpdater

	up                prometheus.Gauge
	totalScrapes      prometheus.Counter
	totalScrapeTime   prometheus.Counter
	jsonParseFailures prometheus.Counter

	snapshotMetrics   []*snapshotMetric
	repositoryMetrics []*repositoryMetric
}

// NewSnapshots defines Snapshots Prometheus metrics
func NewSnapshots(ctx context.Context, logger log.Logger, client *http.Client, url *url.URL, interval time.Duration) *Snapshots {
	snapshots := &Snapshots{
		logger: logger,

		updater: &snapshotsUpdater{
			logger:   logger,
			client:   client,
			url:      url,
			interval: interval,
			sync:     make(chan struct{}, 1),
		},

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "snapshot_stats", "up"),
			Help: "Was the last scrape of the ElasticSearch snapshots endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "snapshot_stats", "total_scrapes"),
			Help: "Current total ElasticSearch snapshots scrapes.",
		}),
		totalScrapeTime: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "snapshot_stats", "total_scrape_time_seconds"),
			Help: "Current total time spent in ElasticSearch snapshots scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "snapshot_stats", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
		snapshotMetrics: []*snapshotMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_number_of_indices"),
					"Number of indices in the last snapshot",
					defaultSnapshotLabels, nil,
				),
				Value: func(snapshotStats SnapshotStatDataResponse) float64 {
					return float64(len(snapshotStats.Indices))
				},
				Labels: defaultSnapshotLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_start_time_timestamp"),
					"Last snapshot start timestamp",
					defaultSnapshotLabels, nil,
				),
				Value: func(snapshotStats SnapshotStatDataResponse) float64 {
					return float64(snapshotStats.StartTimeInMillis / 1000)
				},
				Labels: defaultSnapshotLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_end_time_timestamp"),
					"Last snapshot end timestamp",
					defaultSnapshotLabels, nil,
				),
				Value: func(snapshotStats SnapshotStatDataResponse) float64 {
					return float64(snapshotStats.EndTimeInMillis / 1000)
				},
				Labels: defaultSnapshotLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_number_of_failures"),
					"Last snapshot number of failures",
					defaultSnapshotLabels, nil,
				),
				Value: func(snapshotStats SnapshotStatDataResponse) float64 {
					return float64(len(snapshotStats.Failures))
				},
				Labels: defaultSnapshotLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_total_shards"),
					"Last snapshot total shards",
					defaultSnapshotLabels, nil,
				),
				Value: func(snapshotStats SnapshotStatDataResponse) float64 {
					return float64(snapshotStats.Shards.Total)
				},
				Labels: defaultSnapshotLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_failed_shards"),
					"Last snapshot failed shards",
					defaultSnapshotLabels, nil,
				),
				Value: func(snapshotStats SnapshotStatDataResponse) float64 {
					return float64(snapshotStats.Shards.Failed)
				},
				Labels: defaultSnapshotLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "snapshot_stats", "snapshot_successful_shards"),
					"Last snapshot successful shards",
					defaultSnapshotLabels, nil,
				),
				Value: func(snapshotStats SnapshotStatDataResponse) float64 {
					return float64(snapshotStats.Shards.Successful)
				},
				Labels: defaultSnapshotLabelValues,
			},
		},
		repositoryMetrics: []*repositoryMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "snapshot_stats", "number_of_snapshots"),
					"Number of snapshots in a repository",
					defaultSnapshotRepositoryLabels, nil,
				),
				Value: func(snapshotsStats SnapshotStatsResponse) float64 {
					return float64(len(snapshotsStats.Snapshots))
				},
				Labels: defaultSnapshotRepositoryLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "snapshot_stats", "oldest_snapshot_timestamp"),
					"Timestamp of the oldest snapshot",
					defaultSnapshotRepositoryLabels, nil,
				),
				Value: func(snapshotsStats SnapshotStatsResponse) float64 {
					if len(snapshotsStats.Snapshots) == 0 {
						return 0
					}
					return float64(snapshotsStats.Snapshots[0].StartTimeInMillis / 1000)
				},
				Labels: defaultSnapshotRepositoryLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "snapshot_stats", "latest_snapshot_timestamp_seconds"),
					"Timestamp of the latest SUCCESS or PARTIAL snapshot",
					defaultSnapshotRepositoryLabels, nil,
				),
				Value: func(snapshotsStats SnapshotStatsResponse) float64 {
					for i := len(snapshotsStats.Snapshots) - 1; i >= 0; i-- {
						var snap = snapshotsStats.Snapshots[i]
						if snap.State == "SUCCESS" || snap.State == "PARTIAL" {
							return float64(snap.StartTimeInMillis / 1000)
						}
					}
					return 0
				},
				Labels: defaultSnapshotRepositoryLabelValues,
			},
		},
	}
	snapshots.updater.Run(ctx)
	return snapshots
}

// Describe add Snapshots metrics descriptions
func (s *Snapshots) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range s.snapshotMetrics {
		ch <- metric.Desc
	}
	ch <- s.up.Desc()
	ch <- s.totalScrapeTime.Desc()
	ch <- s.totalScrapes.Desc()
	ch <- s.jsonParseFailures.Desc()
}

// Collect gets Snapshots metric values
func (s *Snapshots) Collect(ch chan<- prometheus.Metric) {
	var now = time.Now()
	s.totalScrapes.Inc()
	defer func() {
		_ = level.Debug(s.logger).Log("msg", "scrape took", "seconds", time.Since(now).Seconds())
		s.totalScrapeTime.Add(time.Since(now).Seconds())
		ch <- s.up
		ch <- s.totalScrapes
		ch <- s.totalScrapeTime
		ch <- s.jsonParseFailures
	}()

	if s.updater.lastError != nil {
		s.up.Set(0)
		if _, ok := s.updater.lastError.(*json.MarshalerError); ok {
			s.jsonParseFailures.Inc()
		}
		_ = level.Warn(s.logger).Log(
			"msg", "failed to fetch and decode snapshot stats",
			"err", s.updater.lastError,
		)
	}
	s.up.Set(1)

	// Snapshots stats
	for repositoryName, snapshotStats := range s.updater.lastResponse {
		for _, metric := range s.repositoryMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(snapshotStats),
				metric.Labels(repositoryName)...,
			)
		}
		if len(snapshotStats.Snapshots) == 0 {
			continue
		}

		lastSnapshot := snapshotStats.Snapshots[len(snapshotStats.Snapshots)-1]
		for _, metric := range s.snapshotMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(lastSnapshot),
				metric.Labels(repositoryName, lastSnapshot)...,
			)
		}
	}
}

type snapshotsUpdater struct {
	logger       log.Logger
	client       *http.Client
	url          *url.URL
	sync         chan struct{}
	interval     time.Duration
	lastResponse map[string]SnapshotStatsResponse
	lastError    error
}

func (upt *snapshotsUpdater) Run(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				_ = level.Info(upt.logger).Log(
					"msg", "context cancelled, exiting snapshots update loop",
					"err", ctx.Err(),
				)
				return
			case <-upt.sync:
				upt.lastResponse, upt.lastError = upt.fetchAndDecodeSnapshotsStats()
				continue
			}
		}
	}(ctx)

	_ = level.Info(upt.logger).Log("msg", "triggering initial snapshots call")
	upt.sync <- struct{}{}

	go func(ctx context.Context) {
		ticker := time.NewTicker(upt.interval)
		for {
			select {
			case <-ctx.Done():
				_ = level.Info(upt.logger).Log(
					"msg", "context cancelled, exiting snapshots trigger loop",
					"err", ctx.Err(),
				)
				return
			case <-ticker.C:
				_ = level.Debug(upt.logger).Log(
					"msg", "triggering periodic snapshots update",
				)
				upt.sync <- struct{}{}
			}
		}
	}(ctx)
}

func (upt *snapshotsUpdater) getAndParseURL(u *url.URL, data interface{}) error {
	res, err := upt.client.Get(u.String())
	if err != nil {
		return fmt.Errorf("failed to get from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(upt.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bts, data); err != nil {
		return err
	}
	return nil
}

func (upt *snapshotsUpdater) fetchAndDecodeSnapshotsStats() (map[string]SnapshotStatsResponse, error) {
	_ = level.Debug(upt.logger).Log("msg", "getting fresh snapshots metrics")
	mssr := make(map[string]SnapshotStatsResponse)

	u := *upt.url
	u.Path = path.Join(u.Path, "/_snapshot")
	var srr SnapshotRepositoriesResponse
	err := upt.getAndParseURL(&u, &srr)
	if err != nil {
		return nil, err
	}
	for repository := range srr {
		u := *upt.url
		u.Path = path.Join(u.Path, "/_snapshot", repository, "/_all")
		var ssr SnapshotStatsResponse
		err := upt.getAndParseURL(&u, &ssr)
		if err != nil {
			continue
		}
		mssr[repository] = ssr
	}

	return mssr, nil
}
