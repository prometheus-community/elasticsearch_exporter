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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/grafana/regexp"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	hotThreadCPUDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "node", "hot_thread_cpu_usage_ratio"),
		"CPU usage ratio of a hot thread sampled over the interval",
		[]string{"cluster", "node", "thread"}, nil,
	)

	nodeHeaderRe = regexp.MustCompile(`^::: \{([^}]+)\}`)
	// The leading percentage is the CPU time before 8.x. From 8.x it is
	// followed by [cpu=..%, other=..%] or [cpu=..%, idle=..%], and for
	// non-transport threads it also counts time spent runnable but off CPU,
	// so the cpu= value is used when present.
	threadLineRe = regexp.MustCompile(`^\s*(\d+(?:\.\d+)?)%(?:\s+\[cpu=(\d+(?:\.\d+)?)%[^\]]*\])?.* cpu usage by thread '([^']+)'`)
	threadNameRe = regexp.MustCompile(`\[([^\]]+)\]\[T#(\d+)\]$`)
)

// Globals because the collector interface does not take constructor args.
var (
	hotThreadsNodes    string
	hotThreadsThreads  int
	hotThreadsInterval time.Duration
)

func init() {
	kingpin.Flag("collector.nodes_hot_threads.nodes",
		"Nodes to sample hot threads on, as an Elasticsearch node specification (e.g. _local or _all).").
		Default("_all").StringVar(&hotThreadsNodes)
	kingpin.Flag("collector.nodes_hot_threads.threads",
		"Number of hot threads to report per node.").
		Default("3").IntVar(&hotThreadsThreads)
	kingpin.Flag("collector.nodes_hot_threads.interval",
		"Interval over which Elasticsearch samples thread CPU usage on each scrape.").
		Default("500ms").DurationVar(&hotThreadsInterval)
	registerCollector("nodes_hot_threads", defaultDisabled, NewNodesHotThreadsCollector)
}

type NodesHotThreadsCollector struct {
	logger *slog.Logger
	hc     *http.Client
	u      *url.URL
}

func NewNodesHotThreadsCollector(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return &NodesHotThreadsCollector{logger: logger, hc: hc, u: u}, nil
}

func (c *NodesHotThreadsCollector) Update(ctx context.Context, uc UpdateContext, ch chan<- prometheus.Metric) error {
	clusterInfo, err := uc.GetClusterInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cluster info: %w", err)
	}

	u := c.u.ResolveReference(&url.URL{Path: "_nodes/" + hotThreadsNodes + "/hot_threads"})
	q := u.Query()
	q.Set("type", "cpu")
	q.Set("threads", strconv.Itoa(hotThreadsThreads))
	q.Set("interval", strconv.FormatInt(hotThreadsInterval.Milliseconds(), 10)+"ms")
	u.RawQuery = q.Encode()

	var threads map[hotThreadKey]float64
	parse := func(r io.Reader) error {
		var err error
		threads, err = parseHotThreads(r)
		return err
	}
	if err := fetchURL(ctx, c.hc, c.logger, u.String(), parse); err != nil {
		return fmt.Errorf("failed to fetch hot threads: %w", err)
	}

	for k, cpuRatio := range threads {
		ch <- prometheus.MustNewConstMetric(
			hotThreadCPUDesc,
			prometheus.GaugeValue,
			cpuRatio,
			clusterInfo.ClusterName,
			k.node,
			k.thread,
		)
	}
	return nil
}

type hotThreadKey struct {
	node   string
	thread string
}

// parseHotThreads returns the CPU usage ratio of each hot thread by node and
// thread label. Threads that share a name on the same node are summed, since
// Elasticsearch can report several of them and a duplicate series would fail
// the whole scrape.
func parseHotThreads(r io.Reader) (map[hotThreadKey]float64, error) {
	threads := map[hotThreadKey]float64{}
	var node string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if m := nodeHeaderRe.FindStringSubmatch(line); m != nil {
			node = m[1]
			continue
		}

		if !strings.Contains(line, " cpu usage by thread '") {
			continue
		}
		if node == "" {
			return nil, fmt.Errorf("hot thread line before any node header: %q", line)
		}
		m := threadLineRe.FindStringSubmatch(line)
		if m == nil {
			return nil, fmt.Errorf("unrecognized hot thread line: %q", line)
		}

		thread, ok := threadLabel(m[3])
		if !ok {
			continue
		}
		pct := m[2]
		if pct == "" {
			pct = m[1]
		}
		// The regex only captures decimal numbers.
		v, _ := strconv.ParseFloat(pct, 64)
		threads[hotThreadKey{node: node, thread: thread}] += v / 100
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if node == "" {
		return nil, errors.New("no node header in hot threads response")
	}
	return threads, nil
}

// threadLabel turns "elasticsearch[node][search][T#3]" into "search[T#3]".
// Other names embed index names, shard ids, file paths or counters, which
// would make the label unbounded, so they are not reported.
func threadLabel(name string) (string, bool) {
	m := threadNameRe.FindStringSubmatch(name)
	if m == nil {
		return "", false
	}
	return m[1] + "[T#" + m[2] + "]", true
}
