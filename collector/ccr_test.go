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
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

func newCCRTestCollector(t *testing.T) (Collector, func()) {
	t.Helper()

	statsBody, err := os.ReadFile(path.Join("../fixtures/ccr/stats", "7.17.0.json"))
	if err != nil {
		t.Fatal(err)
	}

	infoBody, err := os.ReadFile(path.Join("../fixtures/ccr/info", "7.17.0.json"))
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/_ccr/stats":
			_, _ = w.Write(statsBody)
			return
		case "/_all/_ccr/info":
			_, _ = w.Write(infoBody)
			return
		}

		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewCCR(promslog.NewNopLogger(), u, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	return c, ts.Close
}

func TestCCRMinimal(t *testing.T) {
	previous := ccrDetailedMetrics
	ccrDetailedMetrics = false
	defer func() {
		ccrDetailedMetrics = previous
	}()

	c, cleanup := newCCRTestCollector(t)
	defer cleanup()
	descs, err := collectCCRMetricDescs(c)
	if err != nil {
		t.Fatal(err)
	}
	if !hasMetric(descs, "elasticsearch_ccr_follow_index_global_checkpoint_lag") {
		t.Fatal("expected core CCR metric elasticsearch_ccr_follow_index_global_checkpoint_lag")
	}
	if !hasMetric(descs, "elasticsearch_ccr_follower_index_status") {
		t.Fatal("expected core CCR metric elasticsearch_ccr_follower_index_status")
	}
	if hasMetric(descs, "elasticsearch_ccr_follow_shard_successful_read_requests_total") {
		t.Fatal("expected detailed shard metrics to be disabled in minimal CCR mode")
	}
	if hasMetric(descs, "elasticsearch_ccr_follower_parameters_max_outstanding_read_requests") {
		t.Fatal("expected detailed follower parameter metrics to be disabled in minimal CCR mode")
	}
}

func TestCCRDetailed(t *testing.T) {
	previous := ccrDetailedMetrics
	ccrDetailedMetrics = true
	defer func() {
		ccrDetailedMetrics = previous
	}()

	c, cleanup := newCCRTestCollector(t)
	defer cleanup()
	descs, err := collectCCRMetricDescs(c)
	if err != nil {
		t.Fatal(err)
	}
	if !hasMetric(descs, "elasticsearch_ccr_follow_shard_successful_read_requests_total") {
		t.Fatal("expected detailed shard metrics to be enabled in detailed CCR mode")
	}
	if !hasMetric(descs, "elasticsearch_ccr_follower_parameters_max_outstanding_read_requests") {
		t.Fatal("expected detailed follower parameter metrics to be enabled in detailed CCR mode")
	}
}

func collectCCRMetricDescs(c Collector) ([]string, error) {
	ch := make(chan prometheus.Metric, 2048)
	errCh := make(chan error, 1)

	go func() {
		errCh <- c.Update(context.Background(), ch)
		close(ch)
	}()

	descs := make([]string, 0, 256)
	for m := range ch {
		descs = append(descs, m.Desc().String())
	}
	return descs, <-errCh
}

func hasMetric(descs []string, metricName string) bool {
	needle := "fqName: \"" + metricName + "\""
	for _, d := range descs {
		if strings.Contains(d, needle) {
			return true
		}
	}
	return false
}
