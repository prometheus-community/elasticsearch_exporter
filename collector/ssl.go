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
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var sslCertificateExpiry = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, "ssl", "certificate_expiry_timestamp_seconds"),
	"Expiry timestamp of an Elasticsearch SSL certificate in Unix time.",
	[]string{"node", "name", "path", "format", "alias", "subject_dn", "serial_number"},
	nil,
)

func init() {
	registerCollector("ssl", defaultDisabled, NewSSLCertificates)
}

type SSLCertificates struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

func NewSSLCertificates(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &SSLCertificates{
		logger: logger,
		hc:     hc,
		u:      u,
	}, nil
}

type SSLCertificate struct {
	Path         string          `json:"path"`
	Format       string          `json:"format"`
	Alias        string          `json:"alias"`
	SubjectDN    string          `json:"subject_dn"`
	SerialNumber string          `json:"serial_number"`
	Expiry       json.RawMessage `json:"expiry"`
}

type sslNodesResponse struct {
	Nodes map[string]sslNodeResponse `json:"nodes"`
}

type sslNodeResponse struct {
	Name string              `json:"name"`
	HTTP sslNodeHTTPResponse `json:"http"`
}

type sslNodeHTTPResponse struct {
	PublishAddress string `json:"publish_address"`
}

func (s *SSLCertificates) Update(ctx context.Context, _ UpdateContext, ch chan<- prometheus.Metric) error {
	var nodes sslNodesResponse

	nodesURL := s.u.ResolveReference(&url.URL{Path: "/_nodes/http"})
	if err := getAndDecodeURL(ctx, s.hc, s.logger, nodesURL.String(), &nodes); err != nil {
		return fmt.Errorf("failed to get Elasticsearch node HTTP addresses: %w", err)
	}

	nodeIDs := make([]string, 0, len(nodes.Nodes))
	for nodeID := range nodes.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)

	for _, nodeID := range nodeIDs {
		node := nodes.Nodes[nodeID]
		certificatesURL, err := sslCertificatesURL(s.u, node.HTTP.PublishAddress)
		if err != nil {
			return fmt.Errorf("failed to build SSL certificates URL for node %q: %w", nodeID, err)
		}

		var certificates []SSLCertificate
		if err := getAndDecodeURL(ctx, s.hc, s.logger, certificatesURL.String(), &certificates); err != nil {
			return fmt.Errorf("failed to get SSL certificates from node %q: %w", nodeID, err)
		}

		for _, certificate := range certificates {
			expiry, err := certificateExpiryTimestamp(certificate.Expiry)
			if err != nil {
				return fmt.Errorf("failed to parse certificate expiry from node %q: %w", nodeID, err)
			}

			ch <- prometheus.MustNewConstMetric(
				sslCertificateExpiry,
				prometheus.GaugeValue,
				expiry,
				nodeID,
				node.Name,
				certificate.Path,
				certificate.Format,
				certificate.Alias,
				certificate.SubjectDN,
				certificate.SerialNumber,
			)
		}
	}

	return nil
}

func sslCertificatesURL(baseURL *url.URL, publishAddress string) (*url.URL, error) {
	host, port, err := net.SplitHostPort(strings.TrimSpace(publishAddress))
	if err != nil {
		return nil, fmt.Errorf("invalid publish address %q: %w", publishAddress, err)
	}

	u := baseURL.ResolveReference(&url.URL{Path: "/_ssl/certificates"})
	u.Host = net.JoinHostPort(host, port)
	return u, nil
}

func certificateExpiryTimestamp(expiry json.RawMessage) (float64, error) {
	var expiryString string
	if err := json.Unmarshal(expiry, &expiryString); err == nil {
		timestamp, err := time.Parse(time.RFC3339, expiryString)
		if err != nil {
			return 0, err
		}
		return float64(timestamp.Unix()), nil
	}

	var expiryMilliseconds float64
	if err := json.Unmarshal(expiry, &expiryMilliseconds); err == nil {
		return expiryMilliseconds / 1000, nil
	}

	return 0, fmt.Errorf("unsupported expiry value %s", string(expiry))
}
