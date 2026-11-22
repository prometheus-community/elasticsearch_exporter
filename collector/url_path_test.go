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
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

// TestCollectorsKeepURLPathPrefix checks that collectors keep the path of the
// configured Elasticsearch URL, for example when it sits behind a reverse
// proxy that serves it under /es.
func TestCollectorsKeepURLPathPrefix(t *testing.T) {
	factories := map[string]func(*slog.Logger, *url.URL, *http.Client) (Collector, error){
		"cluster-settings": NewClusterSettings,
		"data-stream":      NewDataStream,
		"health-report":    NewHealthReport,
		"ilm":              NewILM,
		"slm":              NewSLM,
		"snapshots":        NewSnapshots,
		"tasks":            NewTaskCollector,
	}

	for name, newCollector := range factories {
		t.Run(name, func(t *testing.T) {
			var (
				mu    sync.Mutex
				paths []string
			)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				paths = append(paths, r.URL.Path)
				mu.Unlock()
				fmt.Fprint(w, `{"repo":{}}`)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL + "/es")
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}

			c, err := newCollector(promslog.NewNopLogger(), u, ts.Client())
			if err != nil {
				t.Fatalf("Failed to create collector: %v", err)
			}

			ch := make(chan prometheus.Metric, 100)
			// The response bodies are not valid for every collector, only the
			// requested paths matter here.
			_ = c.Update(context.Background(), &mockUpdateContext{}, ch)

			mu.Lock()
			defer mu.Unlock()
			if len(paths) == 0 {
				t.Fatal("no request was made")
			}
			for _, p := range paths {
				if !strings.HasPrefix(p, "/es/") {
					t.Errorf("request path %q does not keep the /es prefix", p)
				}
			}
		})
	}
}
