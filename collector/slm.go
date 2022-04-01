// Copyright 2021 The Prometheus Authors
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
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type policyMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(policyStats PolicyStats) float64
	Labels func(policyStats PolicyStats) []string
}

type slmMetric struct {
	Type  prometheus.ValueType
	Desc  *prometheus.Desc
	Value func(slmStats SLMStatsResponse) float64
}

var (
	defaultPolicyLabels      = []string{"policy"}
	defaultPolicyLabelValues = func(policyStats PolicyStats) []string {
		return []string{policyStats.Policy}
	}
)

// SLM information struct
type SLM struct {
	logger log.Logger
	client *http.Client
	url    *url.URL

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter

	slmMetrics    []*slmMetric
	policyMetrics []*policyMetric
}

// NewSLM defines SLM Prometheus metrics
func NewSLM(logger log.Logger, client *http.Client, url *url.URL) *SLM {
	return &SLM{
		logger: logger,
		client: client,
		url:    url,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "slm_stats", "up"),
			Help: "Was the last scrape of the ElasticSearch SLM endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "slm_stats", "total_scrapes"),
			Help: "Current total ElasticSearch SLM scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "slm_stats", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
		slmMetrics: []*slmMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "retention_runs"),
					"Total retention runs",
					nil, nil,
				),
				Value: func(slmStats SLMStatsResponse) float64 {
					return float64(slmStats.RetentionRuns)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "retention_failed"),
					"Total failed retention runs",
					nil, nil,
				),
				Value: func(slmStats SLMStatsResponse) float64 {
					return float64(slmStats.RetentionFailed)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "retention_timed_out"),
					"Total timed out retention runs",
					nil, nil,
				),
				Value: func(slmStats SLMStatsResponse) float64 {
					return float64(slmStats.RetentionTimedOut)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "retention_deletion_time_millis"),
					"Retention run deletion time",
					nil, nil,
				),
				Value: func(slmStats SLMStatsResponse) float64 {
					return float64(slmStats.RetentionDeletionTimeMillis)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "total_snapshots_taken"),
					"Total snapshots taken",
					nil, nil,
				),
				Value: func(slmStats SLMStatsResponse) float64 {
					return float64(slmStats.TotalSnapshotsTaken)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "total_snapshots_failed"),
					"Total snapshots failed",
					nil, nil,
				),
				Value: func(slmStats SLMStatsResponse) float64 {
					return float64(slmStats.TotalSnapshotsFailed)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "total_snapshots_deleted"),
					"Total snapshots deleted",
					nil, nil,
				),
				Value: func(slmStats SLMStatsResponse) float64 {
					return float64(slmStats.TotalSnapshotsDeleted)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "total_snapshot_deletion_failures"),
					"Total snapshot deletion failures",
					nil, nil,
				),
				Value: func(slmStats SLMStatsResponse) float64 {
					return float64(slmStats.TotalSnapshotDeletionFailures)
				},
			},
		},
		policyMetrics: []*policyMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "snapshots_taken"),
					"Total snapshots taken",
					defaultPolicyLabels, nil,
				),
				Value: func(policyStats PolicyStats) float64 {
					return float64(policyStats.SnapshotsTaken)
				},
				Labels: defaultPolicyLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "snapshots_failed"),
					"Total snapshots failed",
					defaultPolicyLabels, nil,
				),
				Value: func(policyStats PolicyStats) float64 {
					return float64(policyStats.SnapshotsFailed)
				},
				Labels: defaultPolicyLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "snapshots_deleted"),
					"Total snapshots deleted",
					defaultPolicyLabels, nil,
				),
				Value: func(policyStats PolicyStats) float64 {
					return float64(policyStats.SnapshotsDeleted)
				},
				Labels: defaultPolicyLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "slm_stats", "snapshot_deletion_failures"),
					"Total snapshot deletion failures",
					defaultPolicyLabels, nil,
				),
				Value: func(policyStats PolicyStats) float64 {
					return float64(policyStats.SnapshotDeletionFailures)
				},
				Labels: defaultPolicyLabelValues,
			},
		},
	}
}

// Describe adds SLM metrics descriptions
func (s *SLM) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range s.slmMetrics {
		ch <- metric.Desc
	}

	for _, metric := range s.policyMetrics {
		ch <- metric.Desc
	}
	ch <- s.up.Desc()
	ch <- s.totalScrapes.Desc()
	ch <- s.jsonParseFailures.Desc()
}

func (s *SLM) fetchAndDecodeSLMStats() (SLMStatsResponse, error) {
	var ssr SLMStatsResponse

	u := *s.url
	u.Path = path.Join(u.Path, "/_slm/stats")
	res, err := s.client.Get(u.String())
	if err != nil {
		return ssr, fmt.Errorf("failed to get slm stats health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(s.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return ssr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		s.jsonParseFailures.Inc()
		return ssr, err
	}

	if err := json.Unmarshal(bts, &ssr); err != nil {
		s.jsonParseFailures.Inc()
		return ssr, err
	}

	return ssr, nil
}

// Collect gets SLM metric values
func (s *SLM) Collect(ch chan<- prometheus.Metric) {
	s.totalScrapes.Inc()
	defer func() {
		ch <- s.up
		ch <- s.totalScrapes
		ch <- s.jsonParseFailures
	}()

	slmStatsResp, err := s.fetchAndDecodeSLMStats()
	if err != nil {
		s.up.Set(0)
		_ = level.Warn(s.logger).Log(
			"msg", "failed to fetch and decode slm stats",
			"err", err,
		)
		return
	}
	s.up.Set(1)

	for _, metric := range s.slmMetrics {
		ch <- prometheus.MustNewConstMetric(
			metric.Desc,
			metric.Type,
			metric.Value(slmStatsResp),
		)
	}

	for _, metric := range s.policyMetrics {
		for _, policy := range slmStatsResp.PolicyStats {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(policy),
				metric.Labels(policy)...,
			)
		}
	}
}
