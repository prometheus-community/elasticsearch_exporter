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

import "time"

// SnapshotStatsResponse is a representation of the snapshots stats
type SnapshotStatsResponse struct {
	Snapshots []SnapshotStatDataResponse `json:"snapshots"`
}

// SnapshotStatDataResponse is a representation of the single snapshot stat
type SnapshotStatDataResponse struct {
	Snapshot          string        `json:"snapshot"`
	UUID              string        `json:"uuid"`
	VersionID         int64         `json:"version_id"`
	Version           string        `json:"version"`
	Indices           []string      `json:"indices"`
	State             string        `json:"state"`
	StartTime         time.Time     `json:"start_time"`
	StartTimeInMillis int64         `json:"start_time_in_millis"`
	EndTime           time.Time     `json:"end_time"`
	EndTimeInMillis   int64         `json:"end_time_in_millis"`
	DurationInMillis  int64         `json:"duration_in_millis"`
	Failures          []interface{} `json:"failures"`
	Shards            struct {
		Total      int64 `json:"total"`
		Failed     int64 `json:"failed"`
		Successful int64 `json:"successful"`
	} `json:"shards"`
}

// SnapshotRepositoriesResponse is a representation snapshots repositories
type SnapshotRepositoriesResponse map[string]struct {
	Type string `json:"type"`
}
