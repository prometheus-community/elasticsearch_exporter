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

package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultTotalFieldsValue = 1000
	defaultDateCreation     = 0

	indicesSettingsTotalFields = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices_settings", "total_fields"),
		"index mapping setting for total_fields",
		[]string{"index"}, nil,
	)
	indicesSettingsReplicas = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices_settings", "replicas"),
		"index setting number_of_replicas",
		[]string{"index"}, nil,
	)
	indicesSettingsShardsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices_settings", "shards"),
		"index setting number_of_shards",
		[]string{"index"}, nil,
	)
	indicesSettingsCreationTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices_settings", "creation_timestamp_seconds"),
		"index setting creation_date",
		[]string{"index"}, nil,
	)
	indicesSettingsReadOnly = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "indices_settings_stats", "read_only_indices"),
		"Current number of read only indices within cluster",
		nil, nil,
	)
)

func init() {
	registerCollector("indices_settings", defaultDisabled, NewIndicesSettings)
}

// IndicesSettings information struct
type IndicesSettings struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

// IndicesSettingsResponse is a representation of Elasticsearch Settings for each Index
type IndicesSettingsResponse map[string]Index

// Index defines the struct of the tree for the settings of each index
type Index struct {
	Settings Settings `json:"settings"`
}

// Settings defines current index settings
type Settings struct {
	IndexInfo IndexInfo `json:"index"`
}

// IndexInfo defines the blocks of the current index
type IndexInfo struct {
	Blocks           Blocks  `json:"blocks"`
	Mapping          Mapping `json:"mapping"`
	NumberOfReplicas string  `json:"number_of_replicas"`
	NumberOfShards   string  `json:"number_of_shards"`
	CreationDate     string  `json:"creation_date"`
}

// Blocks defines whether current index has read_only_allow_delete enabled
type Blocks struct {
	ReadOnly string `json:"read_only_allow_delete"`
}

// Mapping defines mapping settings
type Mapping struct {
	TotalFields TotalFields `json:"total_fields"`
}

// TotalFields defines the limit on the number of mapped fields
type TotalFields struct {
	Limit string `json:"limit"`
}

// NewIndicesSettings defines Indices Settings Prometheus metrics
func NewIndicesSettings(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &IndicesSettings{
		logger: logger,
		hc:     hc,
		u:      u,
	}, nil
}

// Update gets all indices settings metric values
func (cs *IndicesSettings) Update(_ context.Context, _ UpdateContext, ch chan<- prometheus.Metric) error {
	var asr IndicesSettingsResponse
	u := *cs.u
	u.Path = path.Join(u.Path, "/_all/_settings")

	res, err := cs.hc.Get(u.String())
	if err != nil {
		return fmt.Errorf("failed to get from %s://%s:%s%s: %w", u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			cs.logger.Warn("failed to close response body", "err", cerr)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&asr); err != nil {
		return err
	}

	var readOnly int
	for indexName, idx := range asr {
		if idx.Settings.IndexInfo.Blocks.ReadOnly == "true" {
			readOnly++
		}

		totalFields := float64(defaultTotalFieldsValue)
		if val, err := strconv.ParseFloat(idx.Settings.IndexInfo.Mapping.TotalFields.Limit, 64); err == nil {
			totalFields = val
		}
		ch <- prometheus.MustNewConstMetric(indicesSettingsTotalFields, prometheus.GaugeValue, totalFields, indexName)

		replicas := float64(defaultTotalFieldsValue)
		if val, err := strconv.ParseFloat(idx.Settings.IndexInfo.NumberOfReplicas, 64); err == nil {
			replicas = val
		}
		ch <- prometheus.MustNewConstMetric(indicesSettingsReplicas, prometheus.GaugeValue, replicas, indexName)

		shards := float64(defaultTotalFieldsValue)
		if val, err := strconv.ParseFloat(idx.Settings.IndexInfo.NumberOfShards, 64); err == nil {
			shards = val
		}
		ch <- prometheus.MustNewConstMetric(indicesSettingsShardsDesc, prometheus.GaugeValue, shards, indexName)

		creationDate := float64(defaultDateCreation)
		if val, err := strconv.ParseFloat(idx.Settings.IndexInfo.CreationDate, 64); err == nil {
			creationDate = val / 1000.0
		}
		ch <- prometheus.MustNewConstMetric(indicesSettingsCreationTimestamp, prometheus.GaugeValue, creationDate, indexName)
	}

	ch <- prometheus.MustNewConstMetric(indicesSettingsReadOnly, prometheus.GaugeValue, float64(readOnly))
	return nil
}
