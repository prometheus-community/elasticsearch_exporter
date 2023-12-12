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
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type clusterLicenseMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(clusterLicenseStats clusterLicenseResponse) float64
	Labels func(clusterLicenseStats clusterLicenseResponse) []string
}

type clusterLicenseResponse struct {
	License struct {
		Status             string    `json:"status"`
		UID                string    `json:"uid"`
		Type               string    `json:"type"`
		IssueDate          time.Time `json:"issue_date"`
		IssueDateInMillis  int64     `json:"issue_date_in_millis"`
		ExpiryDate         time.Time `json:"expiry_date"`
		ExpiryDateInMillis int64     `json:"expiry_date_in_millis"`
		MaxNodes           int       `json:"max_nodes"`
		IssuedTo           string    `json:"issued_to"`
		Issuer             string    `json:"issuer"`
		StartDateInMillis  int64     `json:"start_date_in_millis"`
	} `json:"license"`
}

var (
	defaultClusterLicenseLabels = []string{"cluster_license_type"}
	defaultClusterLicenseValues = func(clusterLicense clusterLicenseResponse) []string {
		return []string{clusterLicense.License.Type}
	}
)

// License Information Struct
type ClusterLicense struct {
	logger log.Logger
	client *http.Client
	url    *url.URL

	clusterLicenseMetrics []*clusterLicenseMetric
}

// NewClusterLicense defines ClusterLicense Prometheus metrics
func NewClusterLicense(logger log.Logger, client *http.Client, url *url.URL) *ClusterLicense {
	return &ClusterLicense{
		logger: logger,
		client: client,
		url:    url,

		clusterLicenseMetrics: []*clusterLicenseMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "cluster_license", "max_nodes"),
					"The max amount of nodes allowed by the license",
					defaultClusterLicenseLabels, nil,
				),
				Value: func(clusterLicenseStats clusterLicenseResponse) float64 {
					return float64(clusterLicenseStats.License.MaxNodes)
				},
				Labels: defaultClusterLicenseValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "cluster_license", "issue_date_in_millis"),
					"License issue date in milliseconds",
					defaultClusterLicenseLabels, nil,
				),
				Value: func(clusterLicenseStats clusterLicenseResponse) float64 {
					return float64(clusterLicenseStats.License.IssueDateInMillis)
				},
				Labels: defaultClusterLicenseValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "cluster_license", "expiry_date_in_millis"),
					"License expiry date in milliseconds",
					defaultClusterLicenseLabels, nil,
				),
				Value: func(clusterLicenseStats clusterLicenseResponse) float64 {
					return float64(clusterLicenseStats.License.ExpiryDateInMillis)
				},
				Labels: defaultClusterLicenseValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "cluster_license", "start_date_in_millis"),
					"License start date in milliseconds",
					defaultClusterLicenseLabels, nil,
				),
				Value: func(clusterLicenseStats clusterLicenseResponse) float64 {
					return float64(clusterLicenseStats.License.StartDateInMillis)
				},
				Labels: defaultClusterLicenseValues,
			},
		},
	}
}

// Describe adds License metrics descriptions
func (cl *ClusterLicense) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range cl.clusterLicenseMetrics {
		ch <- metric.Desc
	}
}

func (cl *ClusterLicense) fetchAndDecodeClusterLicense() (clusterLicenseResponse, error) {
	var clr clusterLicenseResponse

	u := *cl.url
	u.Path = path.Join(u.Path, "/_license")
	res, err := cl.client.Get(u.String())
	if err != nil {
		return clr, fmt.Errorf("failed to get license stats from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			level.Warn(cl.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return clr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := io.ReadAll(res.Body)
	if err != nil {
		return clr, err
	}

	if err := json.Unmarshal(bts, &clr); err != nil {
		return clr, err
	}

	return clr, nil
}

// Collect gets ClusterLicense metric values
func (cl *ClusterLicense) Collect(ch chan<- prometheus.Metric) {

	clusterLicenseResp, err := cl.fetchAndDecodeClusterLicense()
	if err != nil {
		level.Warn(cl.logger).Log(
			"msg", "failed to fetch and decode license stats",
			"err", err,
		)
		return
	}

	for _, metric := range cl.clusterLicenseMetrics {
		ch <- prometheus.MustNewConstMetric(
			metric.Desc,
			metric.Type,
			metric.Value(clusterLicenseResp),
			metric.Labels(clusterLicenseResp)...,
		)
	}
}
