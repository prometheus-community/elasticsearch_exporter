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
	"context"
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
	defaultClusterLicenseLabels       = []string{"issued_to", "issuer", "type", "status"}
	defaultClusterLicenseLabelsValues = func(clusterLicense clusterLicenseResponse) []string {
		return []string{clusterLicense.License.IssuedTo, clusterLicense.License.Issuer, clusterLicense.License.Type, clusterLicense.License.Status}
	}
)

var (
	licenseMaxNodes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_license", "max_nodes"),
		"The max amount of nodes allowed by the license.",
		defaultClusterLicenseLabels, nil,
	)
	licenseIssueDate = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_license", "issue_date_seconds"),
		"License issue date since unix epoch in seconds.",
		defaultClusterLicenseLabels, nil,
	)
	licenseExpiryDate = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_license", "expiry_date_seconds"),
		"License expiry date since unix epoch in seconds.",
		defaultClusterLicenseLabels, nil,
	)
	licenseStartDate = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "cluster_license", "start_date_seconds"),
		"License start date since unix epoch in seconds.",
		defaultClusterLicenseLabels, nil,
	)
)

func init() {
	registerCollector("cluster_license", defaultDisabled, NewClusterLicense)
}

// License Information Struct
type ClusterLicense struct {
	logger log.Logger
	hc     *http.Client
	u      *url.URL
}

func NewClusterLicense(logger log.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &ClusterLicense{
		logger: logger,
		u:      u,
		hc:     hc,
	}, nil
}

func (c *ClusterLicense) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	var clr clusterLicenseResponse

	u := *c.u
	u.Path = path.Join(u.Path, "/_license")
	res, err := c.hc.Get(u.String())

	if err != nil {
		return err
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			level.Warn(c.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bts, &clr); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		licenseMaxNodes,
		prometheus.GaugeValue,
		float64(clr.License.MaxNodes),
		defaultClusterLicenseLabelsValues(clr)...,
	)

	ch <- prometheus.MustNewConstMetric(
		licenseIssueDate,
		prometheus.GaugeValue,
		float64(clr.License.IssueDateInMillis/1000),
		defaultClusterLicenseLabelsValues(clr)...,
	)

	ch <- prometheus.MustNewConstMetric(
		licenseExpiryDate,
		prometheus.GaugeValue,
		float64(clr.License.ExpiryDateInMillis/1000),
		defaultClusterLicenseLabelsValues(clr)...,
	)

	ch <- prometheus.MustNewConstMetric(
		licenseStartDate,
		prometheus.GaugeValue,
		float64(clr.License.StartDateInMillis/1000),
		defaultClusterLicenseLabelsValues(clr)...,
	)

	return nil
}
