package collector

// ClusterSettingsFullResponse is a representation of a Elasticsearch Cluster Settings
type ClusterSettingsFullResponse struct {
	Defaults   ClusterSettingsResponse `json:"defaults"`
	Persistent ClusterSettingsResponse `json:"persistent"`
	Transient  ClusterSettingsResponse `json:"transient"`
}

// ClusterSettingsResponse is a representation of a Elasticsearch Cluster Settings
type ClusterSettingsResponse struct {
	Cluster Cluster `json:"cluster"`
}

// Cluster is a representation of a Elasticsearch Cluster Settings
type Cluster struct {
	Routing Routing `json:"routing"`
}

// Routing is a representation of a Elasticsearch Cluster shard routing configuration
type Routing struct {
	Allocation Allocation `json:"allocation"`
}

// Allocation is a representation of a Elasticsearch Cluster shard routing allocation settings
type Allocation struct {
	Enabled string `json:"enable"`
}
