// Copyright 2023 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type taskByAction struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(action string, count int64) float64
	Labels func(action string, count int64) []string
}

var (
	taskLabels = []string{"cluster", "action"}
)

// Task Information Struct
type Task struct {
	logger  log.Logger
	client  *http.Client
	url     *url.URL
	actions string

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter

	byActionMetrics []*taskByAction
}

// NewTask defines Task Prometheus metrics
func NewTask(logger log.Logger, client *http.Client, url *url.URL, actions string) *Task {
	return &Task{
		logger:  logger,
		client:  client,
		url:     url,
		actions: actions,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "task_stats", "up"),
			Help: "Was the last scrape of the ElasticSearch Task endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "task_stats", "total_scrapes"),
			Help: "Current total Elasticsearch snapshots scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "task_stats", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
		byActionMetrics: []*taskByAction{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "task_stats", "action_total"),
					"Number of tasks of a certain action",
					[]string{"action"}, nil,
				),
				Value: func(action string, count int64) float64 {
					return float64(count)
				},
				Labels: func(action string, count int64) []string {
					return []string{action}
				},
			},
		},
	}
}

// Describe adds Task metrics descriptions
func (t *Task) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range t.byActionMetrics {
		ch <- metric.Desc
	}

	ch <- t.up.Desc()
	ch <- t.totalScrapes.Desc()
	ch <- t.jsonParseFailures.Desc()
}

func (t *Task) fetchAndDecodeAndAggregateTaskStats() (*AggregatedTaskStats, error) {
	u := *t.url
	u.Path = path.Join(u.Path, "/_tasks")
	u.RawQuery = "group_by=none&actions=" + t.actions
	res, err := t.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get data stream stats health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			level.Warn(t.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request to %v failed with code %d", u.String(), res.StatusCode)
	}

	bts, err := io.ReadAll(res.Body)
	if err != nil {
		t.jsonParseFailures.Inc()
		return nil, err
	}

	var tr TasksResponse
	if err := json.Unmarshal(bts, &tr); err != nil {
		t.jsonParseFailures.Inc()
		return nil, err
	}

	stats := AggregateTasks(tr)
	return stats, nil
}

// Collect gets Task metric values
func (ds *Task) Collect(ch chan<- prometheus.Metric) {
	ds.totalScrapes.Inc()
	defer func() {
		ch <- ds.up
		ch <- ds.totalScrapes
		ch <- ds.jsonParseFailures
	}()

	stats, err := ds.fetchAndDecodeAndAggregateTaskStats()
	if err != nil {
		ds.up.Set(0)
		level.Warn(ds.logger).Log(
			"msg", "failed to fetch and decode task stats",
			"err", err,
		)
		return
	}

	for action, count := range stats.CountByAction {
		for _, metric := range ds.byActionMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(action, count),
				metric.Labels(action, count)...,
			)
		}
	}

	ds.up.Set(1)
}
