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
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

const hotThreadsIdleFixture = `::: {node1}{abc123}{127.0.0.1}{127.0.0.1:9300}
   Hot threads at 2024-01-01T00:00:00.000Z, interval=500ms, busiestThreads=3, ignoreIdleThreads=true:
`

func newHotThreadsTestCollector(t *testing.T, handler http.HandlerFunc) Collector {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse URL: %s", err)
	}

	c, err := NewNodesHotThreadsCollector(promslog.NewNopLogger(), u, ts.Client())
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}
	return c
}

func serveBody(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, body)
	}
}

func TestNodesHotThreads(t *testing.T) {
	tcs := map[string]struct {
		fixture string
		want    string
	}{
		// Before 8.x the leading percentage is the CPU time of the thread.
		"old-format": {
			fixture: `::: {node1}{abc123}{127.0.0.1}{127.0.0.1:9300}
   Hot threads at 2024-01-01T00:00:00.000Z, interval=500ms, busiestThreads=3, ignoreIdleThreads=true:

   75.0% (375ms out of 500ms) cpu usage by thread 'elasticsearch[node1][search][T#3]'
    10/10 snapshots sharing following 2 elements
      sun.nio.ch.EPoll.wait(Native Method)

   25.0% (125ms out of 500ms) cpu usage by thread 'elasticsearch[node1][bulk][T#1]'
    5/10 snapshots sharing following 2 elements
      sun.nio.ch.EPoll.wait(Native Method)
`,
			want: `# HELP elasticsearch_node_hot_thread_cpu_usage_ratio CPU usage ratio of a hot thread sampled over the interval
# TYPE elasticsearch_node_hot_thread_cpu_usage_ratio gauge
elasticsearch_node_hot_thread_cpu_usage_ratio{cluster="",node="node1",thread="bulk[T#1]"} 0.25
elasticsearch_node_hot_thread_cpu_usage_ratio{cluster="",node="node1",thread="search[T#3]"} 0.75
`,
		},
		// From 8.x the leading percentage of a non-transport thread is cpu plus
		// other (runnable but not on CPU), so the value has to come from cpu=.
		// Captured from Elasticsearch 8.13.2 under bulk indexing load. 54.8/100
		// is 0.5479999999999999 in float64.
		"new-format": {
			fixture: `::: {node1}{abc123}{127.0.0.1}{127.0.0.1:9300}
   Hot threads at 2024-01-01T00:00:00.000Z, interval=500ms, busiestThreads=3, ignoreIdleThreads=true:

   91.2% [cpu=54.8%, other=36.4%] (456ms out of 500ms) cpu usage by thread 'elasticsearch[node1][write][T#2]'
   10.2% [cpu=7.6%, other=2.6%] (51ms out of 500ms) cpu usage by thread 'elasticsearch[node1][write][T#4]'
    0.0% [cpu=0.0%, idle=100.0%] (0s out of 500ms) cpu usage by thread 'elasticsearch[node1][transport_worker][T#3]'
`,
			want: `# HELP elasticsearch_node_hot_thread_cpu_usage_ratio CPU usage ratio of a hot thread sampled over the interval
# TYPE elasticsearch_node_hot_thread_cpu_usage_ratio gauge
elasticsearch_node_hot_thread_cpu_usage_ratio{cluster="",node="node1",thread="transport_worker[T#3]"} 0
elasticsearch_node_hot_thread_cpu_usage_ratio{cluster="",node="node1",thread="write[T#2]"} 0.5479999999999999
elasticsearch_node_hot_thread_cpu_usage_ratio{cluster="",node="node1",thread="write[T#4]"} 0.076
`,
		},
		// Names without a [pool][T#n] suffix embed index names, shard ids, file
		// paths and counters, so they are not reported.
		"unbounded-names": {
			fixture: `::: {node1}{abc123}{127.0.0.1}{127.0.0.1:9300}
   Hot threads at 2024-01-01T00:00:00.000Z, interval=500ms, busiestThreads=3, ignoreIdleThreads=true:

   50.0% [cpu=50.0%, other=0.0%] (250ms out of 500ms) cpu usage by thread 'elasticsearch[node1][[loadtest][0]: Lucene Merge Thread #2]'
   25.0% [cpu=25.0%, other=0.0%] (125ms out of 500ms) cpu usage by thread 'elasticsearch[node1][search][T#3]'
   12.5% [cpu=12.5%, other=0.0%] (62.5ms out of 500ms) cpu usage by thread 'elasticsearch[file-watcher[/usr/share/elasticsearch/config/operator/settings.json]]'
`,
			want: `# HELP elasticsearch_node_hot_thread_cpu_usage_ratio CPU usage ratio of a hot thread sampled over the interval
# TYPE elasticsearch_node_hot_thread_cpu_usage_ratio gauge
elasticsearch_node_hot_thread_cpu_usage_ratio{cluster="",node="node1",thread="search[T#3]"} 0.25
`,
		},
		// Elasticsearch can report several threads with the same name. A
		// duplicate series would fail the whole scrape, so they are summed.
		"duplicate-names": {
			fixture: `::: {node1}{abc123}{127.0.0.1}{127.0.0.1:9300}
   Hot threads at 2024-01-01T00:00:00.000Z, interval=500ms, busiestThreads=3, ignoreIdleThreads=true:

   25.0% [cpu=25.0%, other=0.0%] (125ms out of 500ms) cpu usage by thread 'elasticsearch[node1][trigger_engine_scheduler][T#1]'
   12.5% [cpu=12.5%, other=0.0%] (62.5ms out of 500ms) cpu usage by thread 'elasticsearch[node1][trigger_engine_scheduler][T#1]'
`,
			want: `# HELP elasticsearch_node_hot_thread_cpu_usage_ratio CPU usage ratio of a hot thread sampled over the interval
# TYPE elasticsearch_node_hot_thread_cpu_usage_ratio gauge
elasticsearch_node_hot_thread_cpu_usage_ratio{cluster="",node="node1",thread="trigger_engine_scheduler[T#1]"} 0.375
`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			c := newHotThreadsTestCollector(t, serveBody(tc.fixture))
			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tc.want)); err != nil {
				t.Fatalf("metrics did not match: %v", err)
			}
		})
	}
}

func TestNodesHotThreadsIdleCluster(t *testing.T) {
	c := newHotThreadsTestCollector(t, serveBody(hotThreadsIdleFixture))

	ch := make(chan prometheus.Metric, 10)
	if err := c.Update(context.Background(), &mockUpdateContext{}, ch); err != nil {
		t.Fatalf("Update on an idle cluster returned %v, want nil", err)
	}
	if len(ch) != 0 {
		t.Fatalf("Update on an idle cluster emitted %d metrics, want 0", len(ch))
	}
}

func TestNodesHotThreadsRequest(t *testing.T) {
	nodes, threads, interval := hotThreadsNodes, hotThreadsThreads, hotThreadsInterval
	t.Cleanup(func() { hotThreadsNodes, hotThreadsThreads, hotThreadsInterval = nodes, threads, interval })
	hotThreadsNodes, hotThreadsThreads, hotThreadsInterval = "_local", 5, 2*time.Second

	var got *url.URL
	c := newHotThreadsTestCollector(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.URL
		fmt.Fprint(w, hotThreadsIdleFixture)
	})

	if err := c.Update(context.Background(), &mockUpdateContext{}, make(chan prometheus.Metric, 10)); err != nil {
		t.Fatalf("Update returned %v", err)
	}

	if got.Path != "/_nodes/_local/hot_threads" {
		t.Errorf("path = %q, want %q", got.Path, "/_nodes/_local/hot_threads")
	}
	want := url.Values{"type": {"cpu"}, "threads": {"5"}, "interval": {"2000ms"}}
	if q := got.Query(); !maps.EqualFunc(q, want, slices.Equal[[]string]) {
		t.Errorf("query = %v, want %v", q, want)
	}
}

func TestNodesHotThreadsCanceledContext(t *testing.T) {
	c := newHotThreadsTestCollector(t, serveBody(hotThreadsIdleFixture))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.Update(ctx, &mockUpdateContext{}, make(chan prometheus.Metric, 10))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Update with a canceled context returned %v, want context.Canceled", err)
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

	got, err := parseHotThreads(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parseHotThreads returned error: %v", err)
	}

	want := map[hotThreadKey]float64{
		{node: "node1", thread: "search[T#3]"}: 0.75,
		{node: "node1", thread: "bulk[T#1]"}:   0.25,
		{node: "node2", thread: "write[T#2]"}:  0.50,
	}
	if !maps.Equal(got, want) {
		t.Fatalf("parseHotThreads = %v, want %v", got, want)
	}
}

func TestParseHotThreadsErrors(t *testing.T) {
	tcs := map[string]string{
		"thread before node header": `   75.0% (375ms out of 500ms) cpu usage by thread 'elasticsearch[node1][search][T#3]'
`,
		"no node header": `Hot threads at 2024-01-01T00:00:00.000Z, interval=500ms, busiestThreads=3, ignoreIdleThreads=true:
`,
		"unrecognized thread line": `::: {node1}{abc123}{127.0.0.1}{127.0.0.1:9300}
   75,0% (375ms out of 500ms) cpu usage by thread 'elasticsearch[node1][search][T#3]'
`,
	}

	for name, input := range tcs {
		t.Run(name, func(t *testing.T) {
			if _, err := parseHotThreads(strings.NewReader(input)); err == nil {
				t.Fatal("parseHotThreads returned nil error")
			}
		})
	}
}
