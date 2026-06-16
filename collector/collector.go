// Copyright The Prometheus Authors
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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/elasticsearch_exporter/cluster"
	"github.com/prometheus-community/elasticsearch_exporter/config"
)

const (
	// Namespace defines the common namespace to be used by all metrics.
	namespace = "elasticsearch"

	defaultEnabled  = true
	defaultDisabled = false
)

type CollectorOptions struct {
	TasksActions string
}

type factoryFunc func(logger *slog.Logger, u *url.URL, hc *http.Client, options CollectorOptions) (Collector, error)

type collectorFactory struct {
	name           string
	defaultEnabled bool
	create         factoryFunc
}

var (
	factories = make(map[string]collectorFactory)
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
	Update(context.Context, UpdateContext, chan<- prometheus.Metric) error
}

func registerCollector(name string, isDefaultEnabled bool, createFunc func(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error)) {
	registerCollectorWithOptions(name, isDefaultEnabled, func(logger *slog.Logger, u *url.URL, hc *http.Client, _ CollectorOptions) (Collector, error) {
		return createFunc(logger, u, hc)
	})
}

func registerCollectorWithOptions(name string, isDefaultEnabled bool, createFunc factoryFunc) {
	factories[name] = collectorFactory{
		name:           name,
		defaultEnabled: isDefaultEnabled,
		create:         createFunc,
	}
}

func DefaultCollectorStates() map[string]bool {
	states := make(map[string]bool, len(factories))
	for name, factory := range factories {
		states[name] = factory.defaultEnabled
	}
	return states
}

func CollectorNames() []string {
	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type ElasticsearchCollector struct {
	Collectors      map[string]Collector
	logger          *slog.Logger
	esURL           *url.URL
	httpClient      *http.Client
	cluserInfo      *cluster.InfoProvider
	collectorStates map[string]bool
	options         CollectorOptions
}

type Option func(*ElasticsearchCollector) error

// NewElasticsearchCollector creates a new ElasticsearchCollector
func NewElasticsearchCollector(logger *slog.Logger, filters []string, options ...Option) (*ElasticsearchCollector, error) {
	e := &ElasticsearchCollector{logger: logger}
	// Apply options to customize the collector
	for _, o := range options {
		if err := o(e); err != nil {
			return nil, err
		}
	}

	if e.cluserInfo == nil {
		return nil, fmt.Errorf("cluster info provider is not set")
	}
	if e.collectorStates == nil {
		e.collectorStates = DefaultCollectorStates()
	}
	if e.options.TasksActions == "" {
		e.options.TasksActions = config.DefaultTasksActions
	}

	f := make(map[string]bool)
	for _, filter := range filters {
		enabled, exist := e.collectorStates[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		if !enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}
		f[filter] = true
	}
	collectors := make(map[string]Collector)
	for _, key := range CollectorNames() {
		enabled := e.collectorStates[key]
		if !enabled || (len(f) > 0 && !f[key]) {
			continue
		}
		factory, ok := factories[key]
		if !ok {
			return nil, fmt.Errorf("missing collector factory: %s", key)
		}
		collector, err := factory.create(logger.With("collector", key), e.esURL, e.httpClient, e.options)
		if err != nil {
			return nil, err
		}
		collectors[key] = collector
	}

	e.Collectors = collectors

	return e, nil
}

func WithElasticsearchURL(esURL *url.URL) Option {
	return func(e *ElasticsearchCollector) error {
		e.esURL = esURL
		return nil
	}
}

func WithHTTPClient(hc *http.Client) Option {
	return func(e *ElasticsearchCollector) error {
		e.httpClient = hc
		return nil
	}
}

func WithClusterInfoProvider(cl *cluster.InfoProvider) Option {
	return func(e *ElasticsearchCollector) error {
		e.cluserInfo = cl
		return nil
	}
}

func WithCollectorStates(states map[string]bool) Option {
	return func(e *ElasticsearchCollector) error {
		e.collectorStates = states
		return nil
	}
}

func WithCollectorOptions(options CollectorOptions) Option {
	return func(e *ElasticsearchCollector) error {
		e.options = options
		return nil
	}
}

// Describe implements the prometheus.Collector interface.
func (e ElasticsearchCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface.
func (e ElasticsearchCollector) Collect(ch chan<- prometheus.Metric) {
	uc := NewDefaultUpdateContext(e.cluserInfo)
	wg := sync.WaitGroup{}
	ctx := context.TODO()
	wg.Add(len(e.Collectors))
	for name, c := range e.Collectors {
		go func(name string, c Collector) {
			execute(ctx, name, c, ch, e.logger, uc)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(ctx context.Context, name string, c Collector, ch chan<- prometheus.Metric, logger *slog.Logger, uc UpdateContext) {
	begin := time.Now()
	err := c.Update(ctx, uc, ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			logger.Debug("collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			logger.Warn("collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		logger.Debug("collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

// ErrNoData indicates the collector found no data to collect, but had no other error.
var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}
