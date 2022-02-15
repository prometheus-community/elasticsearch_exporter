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

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var (
	defaultHotThreadsLabels = []string{"node", "thread_name", "thread_id"}

	defaultHotThreadsLabelValues = func(HotThreads string) []string {
		return []string{
			HotThreads,
		}
	}
	NODE_OUTPUT_SEPERATOR = ":::"
	HOT_THREADS_OP_REGEX  = `^?([0-9]*[.])?[0-9]+%.*`
	CPU_PERCENTAGE_REGEX  = `^?([0-9]*[.])?[0-9]+%`
)

// HotThreads information struct
type HotThreads struct {
	logger log.Logger
	url    *url.URL

	HotThreadsMetrics        HotThreadsMetric
	HotThreadsFailureMetrics HotThreadsStepFailureMetric

	jsonParseFailures prometheus.Counter
}

type HotThreadsMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(HotThreadsExp float64) float64
	Labels func(HotThreadsDataNode string, HotThreadsName, HotThreadsId string) []string
}

type HotThreadsStepFailureMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(HotThreadsExp int64) float64
	Labels func(HotThreadsIndex string, HotThreadsPolicy string, action string, step string) []string
}

func getEnv(key, defaultVal string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultVal
	}
	return value
}

// NewHotThreadsExplain defines HotThreads Prometheus metrics
func NewHotThreads(logger log.Logger, url *url.URL) *HotThreads {
	return &HotThreads{
		logger: logger,
		url:    url,

		HotThreadsMetrics: HotThreadsMetric{
			Type: prometheus.GaugeValue,
			Desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "hot_threads", "cpu_usage_percentage"),
				"Hot Threads cpu usage on data nodes",
				defaultHotThreadsLabels, nil,
			),
			Value: func(HotThreadsCpuPercentage float64) float64 {
				return float64(HotThreadsCpuPercentage)
			},
			Labels: func(HotThreadsDataNode string, HotThreadsName, HotThreadsId string) []string {
				return []string{HotThreadsDataNode, HotThreadsName, HotThreadsId}
			},
		},

		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "hot_threads", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
	}
}

// Describe HotThreads
func (s *HotThreads) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.jsonParseFailures.Desc()
	ch <- s.HotThreadsMetrics.Desc
}

func (s *HotThreads) getAndParseURL(u *url.URL, hotThreads *[]HotThreadsRsp) error {
	res, err := http.Get(u.String())
	if err != nil {
		return fmt.Errorf("failed to get from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		_ = level.Warn(s.logger).Log(
			"msg", "failed to get resp body",
			"err", err,
		)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(s.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	sb := string(body)
	hotThreadsNodeOp := strings.Split(string(sb), NODE_OUTPUT_SEPERATOR)

	for _, nodeData := range hotThreadsNodeOp {
		nodeName := strings.Trim(strings.Split(nodeData, "}")[0], " {")

		hotThreadsOpRegex := regexp.MustCompile(HOT_THREADS_OP_REGEX)
		allHotThreads := hotThreadsOpRegex.FindAllString(nodeData, -1)
		cpuPercentageRegex := regexp.MustCompile(CPU_PERCENTAGE_REGEX)

		for _, v := range allHotThreads {
			cpu := string(cpuPercentageRegex.FindString(v))
			cpu = strings.Trim(cpu, "%")
			threadName := ""
			threadId := ""
			data := strings.Split(v, "usage by thread")
			if len(data) > 1 {
				// longThreadName would be one of these string patterns -
				// "process reaper"
				// "elasticsearch[keepAlive/7.0.1]"
				// "elasticsearch[dragoneye-es-managed-data-6][refresh][T#3]"
				// "elasticsearch[elasticsearch-data-0][[geonames][0]: Lucene Merge Thread #12]"

				longThreadName := data[1]
				threadName = longThreadName
				threadId = ""
				// does not contain "[]" or ":" with exception of elasticsearch[keepAlive/7.0.1]
				if strings.Contains(longThreadName, "[") || strings.Contains(longThreadName, ":") {
					if strings.Contains(longThreadName, "keepAlive") {
						threadName = "keepAlive"
						threadId = ""
					} else {
						if strings.Contains(longThreadName, "Lucene Merge Thread") {
							// lucene merge thread  like - elasticsearch[elasticsearch-data-0][[geonames][0]: Lucene Merge Thread #12]
							thread := strings.Trim(strings.Split(longThreadName, ":")[1], "[]'")
							threadName = "merge"
							threadId = strings.Split(thread, "#")[1]
						} else {
							// search, write, refresh, transport_worker etc. like - elasticsearch[elasticsearch-data-0][write][T#2]
							threadName = strings.Trim(strings.Split(longThreadName, "][")[1], "[]'")
							threadId = strings.Trim((strings.Split(longThreadName, "][")[2]), "T#[]'")
						}
					}
				}
			}
			cpuPercentage := 0.0
			cpuPercentage, err := strconv.ParseFloat(cpu, 64)
			if err != nil {
				_ = level.Warn(s.logger).Log(
					"msg", "error parsing cpu percentage",
					"info", err,
				)
			}
			t := &HotThreadsRsp{CpuPercentage: cpuPercentage, Node: nodeName, ThreadName: threadName, ThreadId: threadId}
			*hotThreads = append(*hotThreads, *t)
		}
	}

	return nil
}

func (s *HotThreads) fetchAndDecodeHotThreads() ([]HotThreadsRsp, error) {

	u := *s.url
	u.Path = path.Join(u.Path, "/_nodes/hot_threads")

	var MAX_HOT_THREADS_COUNT = getEnv("MAX_HOT_THREADS_COUNT", "3")
	var HOT_THREADS_SECOND_SAMPLING_INTERVAL = getEnv("HOT_THREADS_SECOND_SAMPLING_INTERVAL", "500ms")

	q := u.Query()
	q.Set("threads", MAX_HOT_THREADS_COUNT)
	q.Set("interval", HOT_THREADS_SECOND_SAMPLING_INTERVAL)
	u.RawQuery = q.Encode()
	u.RawPath = q.Encode()
	var ifr []HotThreadsRsp
	err := s.getAndParseURL(&u, &ifr)

	if err != nil {
		return ifr, err
	}
	return ifr, err
}

// Collect gets cluster hot threads metric values
func (s *HotThreads) Collect(ch chan<- prometheus.Metric) {

	defer func() {
		ch <- s.jsonParseFailures
	}()

	ir, err := s.fetchAndDecodeHotThreads()
	if err != nil {
		_ = level.Warn(s.logger).Log(
			"msg", "failed to fetch and decode HotThreads stats",
			"err", err,
		)
		return
	}

	for _, t := range ir {
		ch <- prometheus.MustNewConstMetric(
			s.HotThreadsMetrics.Desc,
			s.HotThreadsMetrics.Type,
			s.HotThreadsMetrics.Value(t.CpuPercentage),
			s.HotThreadsMetrics.Labels(t.Node, t.ThreadName, t.ThreadId)...,
		)
	}
}
