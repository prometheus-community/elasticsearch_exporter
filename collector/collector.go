// Copyright 2022 The Prometheus Authors
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
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Namespace defines the common namespace to be used by all metrics.
	namespace = "elasticsearch"

	defaultEnabled  = true
	defaultDisabled = false
)

type factoryFunc func(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error)

var (
	factories              = make(map[string]factoryFunc)
	initiatedCollectorsMtx = sync.Mutex{}
	initiatedCollectors    = make(map[string]Collector)
	collectorState         = make(map[string]*bool)
	forcedCollectors       = map[string]bool{} // collectors which have been explicitly enabled or disabled
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

func registerCollector(name string, isDefaultEnabled bool, createFunc factoryFunc) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	// Create flag for this collector
	flagName := fmt.Sprintf("collector.%s", name)
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", name, helpDefaultState)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)

	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Action(collectorFlagAction(name)).Bool()
	collectorState[name] = flag

	// Register the create function for this collector
	factories[name] = createFunc
}

type ElasticsearchCollector struct {
	Collectors map[string]Collector
	logger     *slog.Logger
	esURL      *url.URL
	httpClient *http.Client
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

	f := make(map[string]bool)
	for _, filter := range filters {
		enabled, exist := collectorState[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		if !*enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}
		f[filter] = true
	}
	collectors := make(map[string]Collector)
	initiatedCollectorsMtx.Lock()
	defer initiatedCollectorsMtx.Unlock()
	for key, enabled := range collectorState {
		if !*enabled || (len(f) > 0 && !f[key]) {
			continue
		}
		if collector, ok := initiatedCollectors[key]; ok {
			collectors[key] = collector
		} else {
			collector, err := factories[key](logger.With("collector", key), e.esURL, e.httpClient)
			if err != nil {
				return nil, err
			}
			collectors[key] = collector
			initiatedCollectors[key] = collector
		}
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

func execute(ctx context.Context, name string, c Collector, ch chan<- prometheus.Metric, logger *slog.Logger) {
	begin := time.Now()
	err := c.Update(ctx, ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			logger.Debug("collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			logger.Error("collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		logger.Debug("collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

// collectorFlagAction generates a new action function for the given collector
// to track whether it has been explicitly enabled or disabled from the command line.
// A new action function is needed for each collector flag because the ParseContext
// does not contain information about which flag called the action.
// See: https://github.com/alecthomas/kingpin/issues/294
func collectorFlagAction(collector string) func(ctx *kingpin.ParseContext) error {
	return func(ctx *kingpin.ParseContext) error {
		forcedCollectors[collector] = true
		return nil
	}
}

// ErrNoData indicates the collector found no data to collect, but had no other error.
var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}
