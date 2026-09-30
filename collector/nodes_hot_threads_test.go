// Copyright The Prometheus Authors
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

func TestNodesHotThreads(t *testing.T) {
	// Fixture covers two ES format variants:
	// - old: "75.0% (375ms out of 500ms) cpu usage by thread '...'"
	// - new: "25.0% [cpu=25.0%, idle=75.0%] (125ms out of 500ms) cpu usage by thread '...'"
	tcs := map[string]string{
		"old-format": `::: {node1}{abc123}{127.0.0.1}{127.0.0.1:9300}
   Hot threads at 2024-01-01T00:00:00.000Z, interval=500ms, busiestThreads=3, ignoreIdleThreads=true:

   75.0% (375ms out of 500ms) cpu usage by thread 'elasticsearch[node1][search][T#3]'
    10/10 snapshots sharing following 2 elements
      sun.nio.ch.EPoll.wait(Native Method)

   25.0% (125ms out of 500ms) cpu usage by thread 'elasticsearch[node1][bulk][T#1]'
    5/10 snapshots sharing following 2 elements
      sun.nio.ch.EPoll.wait(Native Method)
`,
		"new-format": `::: {node1}{abc123}{127.0.0.1}{127.0.0.1:9300}
   Hot threads at 2024-01-01T00:00:00.000Z, interval=500ms, busiestThreads=3, ignoreIdleThreads=true:

   75.0% [cpu=75.0%, idle=25.0%] (375ms out of 500ms) cpu usage by thread 'elasticsearch[node1][search][T#3]'
    10/10 snapshots sharing following 2 elements
      sun.nio.ch.EPoll.wait(Native Method)

   25.0% [cpu=25.0%, idle=75.0%] (125ms out of 500ms) cpu usage by thread 'elasticsearch[node1][bulk][T#1]'
    5/10 snapshots sharing following 2 elements
      sun.nio.ch.EPoll.wait(Native Method)
`,
	}

	want := `# HELP elasticsearch_node_hot_thread_cpu_usage_ratio CPU usage ratio of a hot thread sampled over the interval
# TYPE elasticsearch_node_hot_thread_cpu_usage_ratio gauge
elasticsearch_node_hot_thread_cpu_usage_ratio{cluster="",node="node1",thread="bulk[T#1]"} 0.25
elasticsearch_node_hot_thread_cpu_usage_ratio{cluster="",node="node1",thread="search[T#3]"} 0.75
`

	for name, fixture := range tcs {
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(w, fixture)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("failed to parse URL: %s", err)
			}

			c, err := NewNodesHotThreadsCollector(promslog.NewNopLogger(), u, ts.Client())
			if err != nil {
				t.Fatalf("failed to create collector: %v", err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(want)); err != nil {
				t.Fatalf("metrics did not match: %v", err)
			}
		})
	}
}

func TestParseHotThreads(t *testing.T) {
	input := `::: {node1}{abc123}{127.0.0.1}{127.0.0.1:9300}
   Hot threads at 2024-01-01T00:00:00.000Z, interval=500ms, busiestThreads=3, ignoreIdleThreads=true:

   75.0% (375ms out of 500ms) cpu usage by thread 'elasticsearch[node1][search][T#3]'
    10/10 snapshots sharing following 2 elements
      sun.nio.ch.EPoll.wait(Native Method)

   25.0% (125ms out of 500ms) cpu usage by thread 'elasticsearch[node1][bulk][T#1]'
    5/10 snapshots sharing following 2 elements
      sun.nio.ch.EPoll.wait(Native Method)

::: {node2}{def456}{10.0.0.2}{10.0.0.2:9300}
   Hot threads at 2024-01-01T00:00:00.000Z, interval=500ms, busiestThreads=3, ignoreIdleThreads=true:

   50.0% (250ms out of 500ms) cpu usage by thread 'elasticsearch[node2][write][T#2]'
    7/10 snapshots sharing following 2 elements
      sun.nio.ch.EPoll.wait(Native Method)
`

	threads, err := parseHotThreads(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parseHotThreads returned error: %v", err)
	}

	if len(threads) != 3 {
		t.Fatalf("expected 3 threads, got %d", len(threads))
	}

	cases := []struct {
		node        string
		threadLabel string
		cpuRatio    float64
	}{
		{"node1", "search[T#3]", 0.75},
		{"node1", "bulk[T#1]", 0.25},
		{"node2", "write[T#2]", 0.50},
	}

	for i, tc := range cases {
		if threads[i].nodeName != tc.node {
			t.Errorf("thread[%d].nodeName = %q, want %q", i, threads[i].nodeName, tc.node)
		}
		if threads[i].threadLabel != tc.threadLabel {
			t.Errorf("thread[%d].threadLabel = %q, want %q", i, threads[i].threadLabel, tc.threadLabel)
		}
		if threads[i].cpuRatio != tc.cpuRatio {
			t.Errorf("thread[%d].cpuRatio = %v, want %v", i, threads[i].cpuRatio, tc.cpuRatio)
		}
	}
}
