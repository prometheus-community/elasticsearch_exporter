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
	"testing"

	"github.com/prometheus/common/promslog"

	"github.com/prometheus-community/elasticsearch_exporter/config"
)

func TestNewRuntimeRequiresValidatedConfig(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	cfg.ElasticsearchURL = "http://localhost:9200"

	if _, err := NewRuntime(context.Background(), promslog.NewNopLogger(), cfg); err == nil {
		t.Fatalf("expected unvalidated config error")
	}
}

func TestConfigCollectorDefaultsMatchRegisteredCollectors(t *testing.T) {
	configDefaults := config.DefaultCollectorConfig()
	collectorDefaults := DefaultCollectorStates()
	for name, enabled := range collectorDefaults {
		configEnabled, ok := configDefaults[name]
		if !ok {
			t.Fatalf("registered collector %q missing from config defaults", name)
		}
		if configEnabled != enabled {
			t.Fatalf("collector %q default mismatch: config=%v collector=%v", name, configEnabled, enabled)
		}
	}
}

func TestNewRuntimeKeepsCollectorStateInstanceScoped(t *testing.T) {
	cfgA := config.NewConfigWithDefaults()
	cfgA.ElasticsearchURL = "http://localhost:9200"
	cfgA.Collectors[config.CollectorTasks] = true
	if err := cfgA.Validate(); err != nil {
		t.Fatalf("validate cfgA: %v", err)
	}
	rtA, err := NewRuntime(context.Background(), promslog.NewNopLogger(), cfgA)
	if err != nil {
		t.Fatalf("new runtime A: %v", err)
	}
	defer rtA.Close()

	cfgB := config.NewConfigWithDefaults()
	cfgB.ElasticsearchURL = "http://localhost:9201"
	cfgB.Collectors[config.CollectorTasks] = false
	if err := cfgB.Validate(); err != nil {
		t.Fatalf("validate cfgB: %v", err)
	}
	rtB, err := NewRuntime(context.Background(), promslog.NewNopLogger(), cfgB)
	if err != nil {
		t.Fatalf("new runtime B: %v", err)
	}
	defer rtB.Close()

	exporterA := elasticsearchCollectorFromRuntime(t, rtA)
	if _, ok := exporterA.Collectors[config.CollectorTasks]; !ok {
		t.Fatalf("runtime A should include tasks collector")
	}
	exporterB := elasticsearchCollectorFromRuntime(t, rtB)
	if _, ok := exporterB.Collectors[config.CollectorTasks]; ok {
		t.Fatalf("runtime B should not include tasks collector")
	}
}

func elasticsearchCollectorFromRuntime(t *testing.T, rt *Runtime) *ElasticsearchCollector {
	t.Helper()
	collectors, err := rt.Collectors()
	if err != nil {
		t.Fatalf("collectors: %v", err)
	}
	for _, c := range collectors {
		if exporter, ok := c.(*ElasticsearchCollector); ok {
			return exporter
		}
	}
	t.Fatalf("runtime did not include ElasticsearchCollector")
	return nil
}
