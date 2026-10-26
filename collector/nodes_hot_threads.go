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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

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
	threadLineRe = regexp.MustCompile(`^\s+(\d+\.?\d*)%.*cpu usage by thread '([^']+)'`)
	threadNameRe = regexp.MustCompile(`\[([^\]]+)\]\[T#(\d+)\]$`)
)

func init() {
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

	threads, err := c.fetchHotThreads()
	if err != nil {
		return fmt.Errorf("failed to fetch hot threads: %w", err)
	}

	if len(threads) == 0 {
		return ErrNoData
	}

	for _, t := range threads {
		ch <- prometheus.MustNewConstMetric(
			hotThreadCPUDesc,
			prometheus.GaugeValue,
			t.cpuRatio,
			clusterInfo.ClusterName,
			t.nodeName,
			t.threadLabel,
		)
	}
	return nil
}

type hotThread struct {
	nodeName    string
	threadLabel string
	cpuRatio    float64
}

func (c *NodesHotThreadsCollector) fetchHotThreads() ([]hotThread, error) {
	u := c.u.ResolveReference(&url.URL{Path: "_nodes/hot_threads"})
	q := u.Query()
	q.Set("type", "cpu")
	u.RawQuery = q.Encode()

	res, err := c.hc.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hot threads from %s: %w", u.String(), err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			c.logger.Warn("failed to close response body", "err", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request to %s failed with code %d", u.String(), res.StatusCode)
	}

	return parseHotThreads(res.Body)
}

func parseHotThreads(r io.Reader) ([]hotThread, error) {
	var threads []hotThread
	var currentNode string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if m := nodeHeaderRe.FindStringSubmatch(line); m != nil {
			currentNode = m[1]
			continue
		}

		if m := threadLineRe.FindStringSubmatch(line); m != nil {
			pct, err := strconv.ParseFloat(m[1], 64)
			if err != nil {
				continue
			}
			threads = append(threads, hotThread{
				nodeName:    currentNode,
				threadLabel: threadLabel(m[2]),
				cpuRatio:    pct / 100,
			})
		}
	}

	return threads, scanner.Err()
}

// threadLabel extracts a readable label from an ES thread name.
// For names like "elasticsearch[node][search][T#3]" it returns "search[T#3]".
// Falls back to the raw name if the pattern doesn't match.
func threadLabel(rawName string) string {
	if m := threadNameRe.FindStringSubmatch(rawName); m != nil {
		return m[1] + "[T#" + m[2] + "]"
	}
	return rawName
}
