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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/elasticsearch_exporter/config"
)

var taskActionDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, "task_stats", "action"),
	"Number of tasks of a certain action",
	[]string{"action"}, nil)

func init() {
	registerCollectorWithOptions("tasks", defaultDisabled, func(logger *slog.Logger, u *url.URL, hc *http.Client, options CollectorOptions) (Collector, error) {
		return NewTaskCollectorWithActions(logger, u, hc, options.TasksActions)
	})
}

// Task Information Struct
type TaskCollector struct {
	logger  *slog.Logger
	hc      *http.Client
	u       *url.URL
	actions string
}

// NewTaskCollector defines Task Prometheus metrics
func NewTaskCollector(logger *slog.Logger, u *url.URL, hc *http.Client) (Collector, error) {
	return NewTaskCollectorWithActions(logger, u, hc, config.DefaultTasksActions)
}

func NewTaskCollectorWithActions(logger *slog.Logger, u *url.URL, hc *http.Client, actions string) (Collector, error) {
	if actions == "" {
		actions = config.DefaultTasksActions
	}
	logger.Info("task collector created",
		"actionFilter", actions,
	)

	return &TaskCollector{
		logger:  logger,
		hc:      hc,
		u:       u,
		actions: actions,
	}, nil
}

func (t *TaskCollector) Update(ctx context.Context, uc UpdateContext, ch chan<- prometheus.Metric) error {
	tasks, err := t.fetchTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch and decode task stats: %w", err)
	}

	stats := AggregateTasks(tasks)
	for action, count := range stats.CountByAction {
		ch <- prometheus.MustNewConstMetric(
			taskActionDesc,
			prometheus.GaugeValue,
			float64(count),
			action,
		)
	}
	return nil
}

func (t *TaskCollector) fetchTasks(_ context.Context) (tasksResponse, error) {
	u := t.u.ResolveReference(&url.URL{Path: "_tasks"})
	q := u.Query()
	q.Set("group_by", "none")
	q.Set("actions", t.actions)
	u.RawQuery = q.Encode()

	var tr tasksResponse
	res, err := t.hc.Get(u.String())
	if err != nil {
		return tr, fmt.Errorf("failed to get data stream stats health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			t.logger.Warn("failed to close response body", "err", cerr)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return tr, fmt.Errorf("HTTP Request to %v failed with code %d", u.String(), res.StatusCode)
	}

	bts, err := io.ReadAll(res.Body)
	if err != nil {
		return tr, err
	}

	if err := json.Unmarshal(bts, &tr); err != nil {
		return tr, err
	}
	return tr, nil
}

// tasksResponse is a representation of the Task management API.
type tasksResponse struct {
	Tasks []taskResponse `json:"tasks"`
}

// taskResponse is a representation of the individual task item returned by task API endpoint.
//
// We only parse a very limited amount of this API for use in aggregation.
type taskResponse struct {
	Action string `json:"action"`
}

type aggregatedTaskStats struct {
	CountByAction map[string]int64
}

func AggregateTasks(t tasksResponse) aggregatedTaskStats {
	actions := map[string]int64{}
	for _, task := range t.Tasks {
		actions[task.Action]++
	}
	return aggregatedTaskStats{CountByAction: actions}
}
