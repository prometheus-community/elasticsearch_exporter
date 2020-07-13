package collector

// RemoteInfoResponse is a representation of a Elasticsearch _remote/info
type RemoteInfoResponse map[string]RemoteCluster

// RemoteClsuter defines the struct of the tree for the Remote Cluster
type RemoteCluster struct {
	Seeds                    []string `json:"seeds"`
	Connected                bool     `json:"connected"`
	NumNodesConnected        int64    `json:"num_nodes_connected"`
	MaxConnectionsPerCluster int64    `json:"max_connections_per_cluster"`
	InitialConnectTimeout    string   `json:"initial_connect_timeout"`
	SkipUnavailable          bool     `json:"skip_unavailable"`
}
