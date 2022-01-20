// Copyright 2022y The Prometheus Authors
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

// Package collector includes all individual collectors to gather and export elasticsearch metrics.
package collector

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Namespace defines the common namespace to be used by all metrics.
const (
	namespace    = "elasticsearch"
	clusterLabel = "cluster"
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "duration_seconds"),
		"elasticsearch_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "success"),
		"elasticsearch_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(context.Context, chan<- prometheus.Metric) error
}

type ElasticsearchCollector struct {
	Collectors map[string]Collector
	logger     log.Logger
}

// NewElasticsearchCollector creates a new ElasticsearchCollector
func NewElasticsearchCollector(logger log.Logger, httpClient *http.Client, esURL *url.URL) (*ElasticsearchCollector, error) {
	collectors := make(map[string]Collector)

	collectors["cluster_health"] = NewClusterHealth(logger, httpClient, esURL)

	return &ElasticsearchCollector{Collectors: collectors, logger: logger}, nil
}

// Describe implements the prometheus.Collector interface.
func (e ElasticsearchCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface.
func (e ElasticsearchCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	ctx := context.TODO()
	wg.Add(len(e.Collectors))
	for name, c := range e.Collectors {
		go func(name string, c Collector) {
			execute(ctx, name, c, ch, e.logger)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(ctx context.Context, name string, c Collector, ch chan<- prometheus.Metric, logger log.Logger) {
	begin := time.Now()
	err := c.Update(ctx, ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		_ = level.Error(logger).Log("msg", "collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		success = 0
	} else {
		_ = level.Debug(logger).Log("msg", "collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}
