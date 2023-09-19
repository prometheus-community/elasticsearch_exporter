// Copyright 2023 The Prometheus Authors
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

// TasksResponse is a representation of the Task management API.
type TasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}

// TaskResponse is a representation of the individual task item returned by task API endpoint.
//
// We only parse a very limited amount of this API for use in aggregation.
type TaskResponse struct {
	Action string `json:"action"`
}

type AggregatedTaskStats struct {
	CountByAction map[string]int64
}

func AggregateTasks(t TasksResponse) *AggregatedTaskStats {
	actions := map[string]int64{}
	for _, task := range t.Tasks {
		actions[task.Action] += 1
	}
	agg := &AggregatedTaskStats{CountByAction: actions}
	return agg
}
