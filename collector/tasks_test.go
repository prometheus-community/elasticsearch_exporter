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
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func TestTasks(t *testing.T) {
	// Test data was collected by running the following:
	//   # create container
	//   docker run -d --name elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" elasticsearch:7.17.11
	//   sleep 15
	//   # start some busy work in background
	//   for i in $(seq 1 500)
	//   do
	//     curl -o /dev/null -sX POST "localhost:9200/a1/_doc" -H 'Content-Type: application/json' -d'{"a1": "'"$i"'"}'
	//     sleep .01
	//     curl -o /dev/null -sX POST "localhost:9200/a1/_doc" -H 'Content-Type: application/json' -d'{"a2": "'"$i"'"}'
	//     sleep .01
	//     curl -o /dev/null -sX POST "localhost:9200/a1/_doc" -H 'Content-Type: application/json' -d'{"a3": "'"$i"'"}'
	//     sleep .01
	//   done &
	//   # try and collect a good sample
	//   curl -X GET 'localhost:9200/_tasks?group_by=none&actions=indices:*'
	//   # cleanup
	//   docker rm --force elasticsearch
	tcs := map[string]string{
		"7.17": `{"tasks":[{"node":"9lWCm1y_QkujaAg75bVx7A","id":70,"type":"transport","action":"indices:admin/index_template/put","start_time_in_millis":1695900464655,"running_time_in_nanos":308640039,"cancellable":false,"headers":{}},{"node":"9lWCm1y_QkujaAg75bVx7A","id":73,"type":"transport","action":"indices:admin/index_template/put","start_time_in_millis":1695900464683,"running_time_in_nanos":280672000,"cancellable":false,"headers":{}},{"node":"9lWCm1y_QkujaAg75bVx7A","id":76,"type":"transport","action":"indices:admin/index_template/put","start_time_in_millis":1695900464711,"running_time_in_nanos":253247906,"cancellable":false,"headers":{}},{"node":"9lWCm1y_QkujaAg75bVx7A","id":93,"type":"transport","action":"indices:admin/index_template/put","start_time_in_millis":1695900464904,"running_time_in_nanos":60230460,"cancellable":false,"headers":{}},{"node":"9lWCm1y_QkujaAg75bVx7A","id":50,"type":"transport","action":"indices:data/write/index","start_time_in_millis":1695900464229,"running_time_in_nanos":734480468,"cancellable":false,"headers":{}},{"node":"9lWCm1y_QkujaAg75bVx7A","id":51,"type":"transport","action":"indices:admin/auto_create","start_time_in_millis":1695900464235,"running_time_in_nanos":729223933,"cancellable":false,"headers":{}}]}`,
	}
	want := `# HELP elasticsearch_task_stats_action Number of tasks of a certain action
# TYPE elasticsearch_task_stats_action gauge
elasticsearch_task_stats_action{action="indices:admin/auto_create"} 1
elasticsearch_task_stats_action{action="indices:admin/index_template/put"} 4
elasticsearch_task_stats_action{action="indices:data/write/index"} 1
`
	for ver, out := range tcs {
		t.Run(ver, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintln(w, out)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}

			c, err := NewTaskCollector(promslog.NewNopLogger(), u, ts.Client())
			if err != nil {
				t.Fatalf("Failed to create collector: %v", err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(want)); err != nil {
				t.Fatalf("Metrics did not match: %v", err)
			}
		})
	}
}
