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

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/log"
)

func TestTasks(t *testing.T) {
	// Test data was collected by running the following:
	//   docker run -d --name elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" elasticsearch:7.17.11
	//   sleep 15
	//   # start some busy work
	//   for i in $(seq 1 1000); do \
	//     curl -o /dev/null -s -X POST "localhost:9200/a1/_doc" -H 'Content-Type: application/json' \
	//     -d'{"abc": "'$i'"}'; done &
	//   curl -X POST "localhost:9200/a1/_delete_by_query?requests_per_second=1&wait_for_completion=false" \
	//     -H 'Content-Type: application/json' -d'{"query": {"match_all": {}}}
	//   # try and collect a good sample
	//   curl -X GET 'localhost:9200/_tasks?group_by=none&actions=indices:*'
	//   docker rm elasticsearch
	tcs := map[string]string{
		"7.17": `{"tasks":[{"node":"NVe9ksxcSu6AJTKlIfI24A","id":17223,"type":"transport","action":"indices:data/write/delete/byquery","start_time_in_millis":1695214684290,"running_time_in_nanos":8003510219,"cancellable":true,"cancelled":false,"headers":{}},{"node":"NVe9ksxcSu6AJTKlIfI24A","id":20890,"type":"transport","action":"indices:data/write/index","start_time_in_millis":1695214692292,"running_time_in_nanos":1611966,"cancellable":false,"headers":{}},{"node":"NVe9ksxcSu6AJTKlIfI24A","id":20891,"type":"transport","action":"indices:data/write/bulk[s]","start_time_in_millis":1695214692292,"running_time_in_nanos":1467298,"cancellable":false,"parent_task_id":"NVe9ksxcSu6AJTKlIfI24A:20890","headers":{}},{"node":"NVe9ksxcSu6AJTKlIfI24A","id":20892,"type":"direct","action":"indices:data/write/bulk[s][p]","start_time_in_millis":1695214692292,"running_time_in_nanos":1437170,"cancellable":false,"parent_task_id":"NVe9ksxcSu6AJTKlIfI24A:20891","headers":{}}]}`,
	}
	for ver, out := range tcs {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprintln(w, out)
		}))
		defer ts.Close()

		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatalf("Failed to parse URL: %s", err)
		}

		task := NewTask(log.NewNopLogger(), http.DefaultClient, u, "indices:*")
		stats, err := task.fetchAndDecodeAndAggregateTaskStats()
		if err != nil {
			t.Fatalf("Failed to fetch or decode data stream stats: %s", err)
		}
		t.Logf("[%s] Task Response: %+v", ver, stats)

		// validate actions aggregations
		if len(stats.CountByAction) != 4 {
			t.Fatal("expected to get 4 tasks")
		}
		if stats.CountByAction["indices:data/write/index"] != 1 {
			t.Fatal("excpected action indices:data/write/delete/byquery to have count 1")
		}
		if stats.CountByAction["indices:data/write/bulk[s]"] != 1 {
			t.Fatal("excpected action indices:data/write/bulk[s] to have count 1")
		}
		if stats.CountByAction["indices:data/write/bulk[s][p]"] != 1 {
			t.Fatal("excpected action indices:data/write/bulk[s][p] to have count 1")
		}
		if stats.CountByAction["indices:data/write/delete/byquery"] != 1 {
			t.Fatal("excpected action indices:data/write/delete/byquery to have count 1")
		}
	}
}
