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
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type wrapCollector struct {
	c Collector
}

func (w wrapCollector) Describe(ch chan<- *prometheus.Desc) {
}

func (w wrapCollector) Collect(ch chan<- prometheus.Metric) {
	w.c.Update(context.Background(), ch)
}

func TestClusterSettingsStats(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine
	//  curl http://localhost:9200/_cluster/settings/?include_defaults=true

	tests := []struct {
		name string
		file string
		want string
	}{
		// MaxShardsPerNode is empty in older versions
		{
			name: "5.4.2",
			file: "../fixtures/settings-5.4.2.json",
			want: `
# HELP elasticsearch_clustersettings_stats_shard_allocation_enabled Current mode of cluster wide shard routing allocation settings.
# TYPE elasticsearch_clustersettings_stats_shard_allocation_enabled gauge
elasticsearch_clustersettings_stats_shard_allocation_enabled 0
`,
		},
		{
			name: "5.4.2-merge",
			file: "../fixtures/settings-merge-5.4.2.json",
			want: `
# HELP elasticsearch_clustersettings_stats_shard_allocation_enabled Current mode of cluster wide shard routing allocation settings.
# TYPE elasticsearch_clustersettings_stats_shard_allocation_enabled gauge
elasticsearch_clustersettings_stats_shard_allocation_enabled 0
`,
		},
		{
			name: "7.3.0",
			file: "../fixtures/settings-7.3.0.json",
			want: `
# HELP elasticsearch_clustersettings_stats_max_shards_per_node Current maximum number of shards per node setting.
# TYPE elasticsearch_clustersettings_stats_max_shards_per_node gauge
elasticsearch_clustersettings_stats_max_shards_per_node 1000
# HELP elasticsearch_clustersettings_stats_shard_allocation_enabled Current mode of cluster wide shard routing allocation settings.
# TYPE elasticsearch_clustersettings_stats_shard_allocation_enabled gauge
elasticsearch_clustersettings_stats_shard_allocation_enabled 0
`,
		},
		{
			name: "7.17.5-persistent-clustermaxshardspernode",
			file: "../fixtures/settings-persistent-clustermaxshardspernode-7.17.5.json",
			want: `
# HELP elasticsearch_clustersettings_stats_max_shards_per_node Current maximum number of shards per node setting.
# TYPE elasticsearch_clustersettings_stats_max_shards_per_node gauge
elasticsearch_clustersettings_stats_max_shards_per_node 1000
# HELP elasticsearch_clustersettings_stats_shard_allocation_enabled Current mode of cluster wide shard routing allocation settings.
# TYPE elasticsearch_clustersettings_stats_shard_allocation_enabled gauge
elasticsearch_clustersettings_stats_shard_allocation_enabled 0
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(w, f)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			c, err := NewClusterSettings(log.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
