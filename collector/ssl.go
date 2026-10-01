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

// Update reports certificates from the connected node by default, or from
// every node in the cluster when --es.all is set (see QueryAllNodes).
// Unlike /_nodes/stats, /_ssl/certificates is not fanned out across the
// cluster by Elasticsearch itself: it only ever returns the certificates
// visible to the node that receives the request, so cluster-wide coverage
// requires querying each node's published address directly.
func (s *SSLCertificates) Update(ctx context.Context, _ UpdateContext, ch chan<- prometheus.Metric) error {
	if !QueryAllNodes {
		return s.updateConnectedNode(ctx, ch)
	}
	return s.updateAllNodes(ctx, ch)
}

// updateConnectedNode queries /_ssl/certificates once, directly on the
// connected node (es.uri). This is the default: it costs exactly one extra
// HTTP request regardless of cluster size, and composes correctly whether
// the exporter is scraping one node or the whole cluster is covered by
// running one exporter per node (cluster-wide coverage then comes from
// Prometheus scraping every node's exporter, not from this collector
// fanning out itself).
func (s *SSLCertificates) updateConnectedNode(ctx context.Context, ch chan<- prometheus.Metric) error {
	certificatesURL := s.u.ResolveReference(&url.URL{Path: "/_ssl/certificates"})

	var certificates []SSLCertificate
	if err := getAndDecodeURL(ctx, s.hc, s.logger, certificatesURL.String(), &certificates); err != nil {
		return fmt.Errorf("failed to get SSL certificates: %w", err)
	}

	return emitSSLCertificates(ch, "", "", certificates)
}

// updateAllNodes discovers every node via /_nodes/http and queries
// /_ssl/certificates directly on each node's published HTTP address, which
// must be reachable from the exporter. Only use this with a single exporter
// instance responsible for the whole cluster: running it from one exporter
// per node turns this into an O(n^2) number of /_ssl/certificates calls
// across the cluster.
func (s *SSLCertificates) updateAllNodes(ctx context.Context, ch chan<- prometheus.Metric) error {
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

		if err := emitSSLCertificates(ch, nodeID, node.Name, certificates); err != nil {
			return fmt.Errorf("failed to parse certificate expiry from node %q: %w", nodeID, err)
		}
	}

	return nil
}

func emitSSLCertificates(ch chan<- prometheus.Metric, nodeID, nodeName string, certificates []SSLCertificate) error {
	for _, certificate := range certificates {
		expiry, err := certificateExpiryTimestamp(certificate.Expiry)
		if err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			sslCertificateExpiry,
			prometheus.GaugeValue,
			expiry,
			nodeID,
			nodeName,
			certificate.Path,
			certificate.Format,
			certificate.Alias,
			certificate.SubjectDN,
			certificate.SerialNumber,
		)
	}

	return nil
}

// sslCertificatesURL builds the /_ssl/certificates URL for a node's
// published HTTP address. Elasticsearch may report that address as either
// "ip:port" or "hostname/ip:port" (when the configured publish host is a
// resolvable name rather than a bare IP), so any "hostname/" prefix is
// stripped before splitting host and port.
func sslCertificatesURL(baseURL *url.URL, publishAddress string) (*url.URL, error) {
	publishAddress = strings.TrimSpace(publishAddress)
	if idx := strings.LastIndex(publishAddress, "/"); idx != -1 {
		publishAddress = publishAddress[idx+1:]
	}

	host, port, err := net.SplitHostPort(publishAddress)
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
