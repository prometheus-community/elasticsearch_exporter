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
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/elasticsearch_exporter/cluster"
)

const (
	// Namespace defines the common namespace to be used by all metrics.
	namespace = "elasticsearch"

	defaultEnabled  = true
	defaultDisabled = false
)

// collectorConfig is passed to every collector factory. Add a field here for a
// new collector option instead of changing every factory signature.
type collectorConfig struct {
	logger *slog.Logger
	u      *url.URL
	hc     *http.Client
	tasks  TasksConfig
}

type factoryFunc func(collectorConfig) (Collector, error)

var (
	factories        = make(map[string]factoryFunc)
	collectorState   = make(map[string]*bool)
	forcedCollectors = map[string]bool{} // collectors which have been explicitly enabled or disabled
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
	cluserInfo *cluster.InfoProvider

	tasksConfig TasksConfig
}

type Option func(*ElasticsearchCollector) error

// NewElasticsearchCollector creates a new ElasticsearchCollector
func NewElasticsearchCollector(logger *slog.Logger, filters []string, options ...Option) (*ElasticsearchCollector, error) {
	e := &ElasticsearchCollector{
		logger:      logger,
		tasksConfig: DefaultTasksConfig(),
	}
	// Apply options to customize the collector
	for _, o := range options {
		if err := o(e); err != nil {
			return nil, err
		}
	}

	if e.cluserInfo == nil {
		return nil, fmt.Errorf("cluster info provider is not set")
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
	for key, enabled := range collectorState {
		if !*enabled || (len(f) > 0 && !f[key]) {
			continue
		}
		collector, err := factories[key](collectorConfig{
			logger: logger.With("collector", key),
			u:      e.esURL,
			hc:     e.httpClient,
			tasks:  e.tasksConfig,
		})
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

// TasksConfig holds the options for the tasks collector.
type TasksConfig struct {
	// Actions filters tasks by action, like the Task API actions parameter.
	Actions string
}

// DefaultTasksConfig returns the tasks collector options used when
// WithTasksConfig is not passed.
func DefaultTasksConfig() TasksConfig {
	return TasksConfig{Actions: "indices:*"}
}

func WithTasksConfig(c TasksConfig) Option {
	return func(e *ElasticsearchCollector) error {
		e.tasksConfig = c
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
