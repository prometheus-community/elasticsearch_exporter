// Copyright 2022 The Prometheus Authors
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

// SLMStatsResponse is a representation of the SLM stats
type SLMStatsResponse struct {
	RetentionRuns                 int64         `json:"retention_runs"`
	RetentionFailed               int64         `json:"retention_failed"`
	RetentionTimedOut             int64         `json:"retention_timed_out"`
	RetentionDeletionTime         string        `json:"retention_deletion_time"`
	RetentionDeletionTimeMillis   int64         `json:"retention_deletion_time_millis"`
	TotalSnapshotsTaken           int64         `json:"total_snapshots_taken"`
	TotalSnapshotsFailed          int64         `json:"total_snapshots_failed"`
	TotalSnapshotsDeleted         int64         `json:"total_snapshots_deleted"`
	TotalSnapshotDeletionFailures int64         `json:"total_snapshot_deletion_failures"`
	PolicyStats                   []PolicyStats `json:"policy_stats"`
}

// PolicyStats is a representation of SLM stats for specific policies
type PolicyStats struct {
	Policy                   string `json:"policy"`
	SnapshotsTaken           int64  `json:"snapshots_taken"`
	SnapshotsFailed          int64  `json:"snapshots_failed"`
	SnapshotsDeleted         int64  `json:"snapshots_deleted"`
	SnapshotDeletionFailures int64  `json:"snapshot_deletion_failures"`
}

// SLMStatusResponse is a representation of the SLM status
type SLMStatusResponse struct {
	OperationMode string `json:"operation_mode"`
}
