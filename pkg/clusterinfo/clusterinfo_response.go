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

package clusterinfo

import (
	"github.com/blang/semver"
)

// Response is the cluster info retrievable from the / endpoint
type Response struct {
	Name        string      `json:"name"`
	ClusterName string      `json:"cluster_name"`
	ClusterUUID string      `json:"cluster_uuid"`
	Version     VersionInfo `json:"version"`
	Tagline     string      `json:"tagline"`
}

// VersionInfo is the version info retrievable from the / endpoint, embedded in Response
type VersionInfo struct {
	Number        semver.Version `json:"number"`
	BuildHash     string         `json:"build_hash"`
	BuildDate     string         `json:"build_date"`
	BuildSnapshot bool           `json:"build_snapshot"`
	LuceneVersion semver.Version `json:"lucene_version"`
}
