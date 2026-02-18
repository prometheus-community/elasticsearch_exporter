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

// RemoteInfoResponse is a representation of a Elasticsearch _remote/info
type RemoteInfoResponse map[string]RemoteCluster

// RemoteClsuter defines the struct of the tree for the Remote Cluster
type RemoteCluster struct {
	Seeds                    []string `json:"seeds"`
	Connected                bool     `json:"connected"`
	NumNodesConnected        int64    `json:"num_nodes_connected"`
	NumProxySocketsConnected int64    `json:"num_proxy_sockets_connected"`
	MaxConnectionsPerCluster int64    `json:"max_connections_per_cluster"`
	InitialConnectTimeout    string   `json:"initial_connect_timeout"`
	SkipUnavailable          bool     `json:"skip_unavailable"`
}
