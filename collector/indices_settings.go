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
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// IndicesSettings information struct
type IndicesSettings struct {
	logger log.Logger
	client *http.Client
	url    *url.URL

	up              prometheus.Gauge
	readOnlyIndices prometheus.Gauge

	totalScrapes, jsonParseFailures prometheus.Counter
	metrics                         []*indicesSettingsMetric
}

var (
	defaultIndicesTotalFieldsLabels = []string{"index"}
	defaultTotalFieldsValue         = 1000 //es default configuration for total fields
)

type indicesSettingsMetric struct {
	Type  prometheus.ValueType
	Desc  *prometheus.Desc
	Value func(indexSettings Settings) float64
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
		metrics: []*indicesSettingsMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices_settings", "total_fields"),
					"index mapping setting for total_fields",
					defaultIndicesTotalFieldsLabels, nil,
				),
				Value: func(indexSettings Settings) float64 {
					val, err := strconv.ParseFloat(indexSettings.IndexInfo.Mapping.TotalFields.Limit, 10)
					if err != nil {
						return float64(defaultTotalFieldsValue)
					}
					return val
				},
			},
		},
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

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		cs.jsonParseFailures.Inc()
		return err
	}

	if err := json.Unmarshal(bts, data); err != nil {
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
	for indexName, value := range asr {
		if value.Settings.IndexInfo.Blocks.ReadOnly == "true" {
			c++
		}
		for _, metric := range cs.metrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(value.Settings),
				indexName,
			)
		}
	}
	cs.readOnlyIndices.Set(float64(c))
}
