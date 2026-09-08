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
	"log/slog"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

// Labels for remote info metrics
var defaultRemoteInfoLabels = []string{"remote_cluster"}

var (
	remoteInfoNumNodesConnected = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "remote_info", "num_nodes_connected"),
		"Number of nodes connected to the remote cluster",
		defaultRemoteInfoLabels,
		nil,
	)
	remoteInfoNumProxySocketsConnected = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "remote_info", "num_proxy_sockets_connected"),
		"Number of proxy sockets connected to the remote cluster",
		defaultRemoteInfoLabels,
		nil,
	)
	remoteInfoMaxConnectionsPerCluster = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "remote_info", "max_connections_per_cluster"),
		"Maximum number of connections allowed per remote cluster",
		defaultRemoteInfoLabels,
		nil,
	)

	remoteInfoMetrics = []*remoteInfoMetric{
		{
			Type: prometheus.GaugeValue,
			Desc: remoteInfoNumNodesConnected,
			Value: func(remoteStats RemoteCluster) float64 {
				return float64(remoteStats.NumNodesConnected)
			},
			Labels: func(remoteCluster string) []string {
				return []string{remoteCluster}
			},
		},
		{
			Type: prometheus.GaugeValue,
			Desc: remoteInfoNumProxySocketsConnected,
			Value: func(remoteStats RemoteCluster) float64 {
				return float64(remoteStats.NumProxySocketsConnected)
			},
			Labels: func(remoteCluster string) []string {
				return []string{remoteCluster}
			},
		},
		{
			Type: prometheus.GaugeValue,
			Desc: remoteInfoMaxConnectionsPerCluster,
			Value: func(remoteStats RemoteCluster) float64 {
				return float64(remoteStats.MaxConnectionsPerCluster)
			},
			Labels: func(remoteCluster string) []string {
				return []string{remoteCluster}
			},
		},
	}
)

func init() {
	registerCollector("remote-info", defaultDisabled, NewRemoteInfo)
}

type remoteInfoMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(remoteStats RemoteCluster) float64
	Labels func(remoteCluster string) []string
}

// RemoteInfo Information Struct
type RemoteInfo struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

// NewRemoteInfo creates a new RemoteInfo struct
func NewRemoteInfo(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &RemoteInfo{
		logger: logger,
		hc:     hc,
		u:      u,
	}, nil
}

// RemoteInfoResponse is a representation of a Elasticsearch _remote/info
type RemoteInfoResponse map[string]RemoteCluster

// RemoteCluster defines the struct of the tree for the Remote Cluster
type RemoteCluster struct {
	Seeds                    []string `json:"seeds"`
	Connected                bool     `json:"connected"`
	NumNodesConnected        int64    `json:"num_nodes_connected"`
	NumProxySocketsConnected int64    `json:"num_proxy_sockets_connected"`
	MaxConnectionsPerCluster int64    `json:"max_connections_per_cluster"`
	InitialConnectTimeout    string   `json:"initial_connect_timeout"`
	SkipUnavailable          bool     `json:"skip_unavailable"`
}

// Update implements [Collector].
func (r *RemoteInfo) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	var rir RemoteInfoResponse

	u := r.u.ResolveReference(&url.URL{Path: "/_remote/info"})

	resp, err := getURL(ctx, r.hc, r.logger, u.String())
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp, &rir); err != nil {
		return err
	}

	// Remote Info
	for remoteCluster, remoteInfo := range rir {
		for _, metric := range remoteInfoMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(remoteInfo),
				metric.Labels(remoteCluster)...,
			)
		}
	}

	return nil
}
