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
	Type     string            `json:"type"`
	Settings map[string]string `json:"settings"`
}
