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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus/client_golang/prometheus"
)

var defaultIndicesMappingsLabels = []string{"index"}

type indicesMappingsMetric struct {
	Type  prometheus.ValueType
	Desc  *prometheus.Desc
	Value func(indexMapping IndexMapping) float64
}

// IndicesMappings information struct
type IndicesMappings struct {
	logger *slog.Logger
	client *http.Client
	url    *url.URL

	metrics []*indicesMappingsMetric
}

// NewIndicesMappings defines Indices IndexMappings Prometheus metrics
func NewIndicesMappings(logger *slog.Logger, client *http.Client, url *url.URL) *IndicesMappings {
	subsystem := "indices_mappings_stats"

	return &IndicesMappings{
		logger: logger,
		client: client,
		url:    url,

		metrics: []*indicesMappingsMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "fields"),
					"Current number fields within cluster.",
					defaultIndicesMappingsLabels, nil,
				),
				Value: func(indexMapping IndexMapping) float64 {
					return countFieldsRecursive(indexMapping.Mappings.Properties, 0)
				},
			},
		},
	}
}

func countFieldsRecursive(properties IndexMappingProperties, fieldCounter float64) float64 {
	// iterate over all properties
	for _, property := range properties {
		if property.Type != nil && *property.Type != "object" {
			// property has a type set - counts as a field unless the value is object
			// as the recursion below will handle counting that
			fieldCounter++

			// iterate over all fields of that property
			for _, field := range property.Fields {
				// field has a type set - counts as a field
				if field.Type != nil {
					fieldCounter++
				}
			}
		}

		// count recursively in case the property has more properties
		if property.Properties != nil {
			fieldCounter = 1 + countFieldsRecursive(property.Properties, fieldCounter)
		}
	}

	return fieldCounter
}

// Describe add Snapshots metrics descriptions
func (im *IndicesMappings) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range im.metrics {
		ch <- metric.Desc
	}
}

func (im *IndicesMappings) getAndParseURL(u *url.URL) (*IndicesMappingsResponse, error) {
	res, err := im.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		im.logger.Warn("failed to read response body", "err", err)
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		im.logger.Warn("failed to close response body", "err", err)
		return nil, err
	}

	var imr IndicesMappingsResponse
	if err := json.Unmarshal(body, &imr); err != nil {
		return nil, err
	}

	return &imr, nil
}

func (im *IndicesMappings) fetchAndDecodeIndicesMappings() (*IndicesMappingsResponse, error) {
	u := *im.url
	u.Path = path.Join(u.Path, "/_all/_mappings")
	return im.getAndParseURL(&u)
}

// Collect gets all indices mappings metric values
func (im *IndicesMappings) Collect(ch chan<- prometheus.Metric) {
	indicesMappingsResponse, err := im.fetchAndDecodeIndicesMappings()
	if err != nil {
		im.logger.Warn(
			"failed to fetch and decode cluster mappings stats",
			"err", err,
		)
		return
	}

	for _, metric := range im.metrics {
		for indexName, mappings := range *indicesMappingsResponse {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(mappings),
				indexName,
			)
		}
	}
}
