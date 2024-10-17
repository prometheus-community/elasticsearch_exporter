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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

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
# HELP elasticsearch_clustersettings_allocation_threshold_enabled Is disk allocation decider enabled.
# TYPE elasticsearch_clustersettings_allocation_threshold_enabled gauge
elasticsearch_clustersettings_allocation_threshold_enabled 1
# HELP elasticsearch_clustersettings_allocation_watermark_high_ratio High watermark for disk usage as a ratio.
# TYPE elasticsearch_clustersettings_allocation_watermark_high_ratio gauge
elasticsearch_clustersettings_allocation_watermark_high_ratio 0.9
# HELP elasticsearch_clustersettings_allocation_watermark_low_ratio Low watermark for disk usage as a ratio.
# TYPE elasticsearch_clustersettings_allocation_watermark_low_ratio gauge
elasticsearch_clustersettings_allocation_watermark_low_ratio 0.85
`,
		},

		{
			name: "5.4.2-merge",
			file: "../fixtures/settings-merge-5.4.2.json",
			want: `
# HELP elasticsearch_clustersettings_stats_shard_allocation_enabled Current mode of cluster wide shard routing allocation settings.
# TYPE elasticsearch_clustersettings_stats_shard_allocation_enabled gauge
elasticsearch_clustersettings_stats_shard_allocation_enabled 0
# HELP elasticsearch_clustersettings_allocation_threshold_enabled Is disk allocation decider enabled.
# TYPE elasticsearch_clustersettings_allocation_threshold_enabled gauge
elasticsearch_clustersettings_allocation_threshold_enabled 0
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
# HELP elasticsearch_clustersettings_allocation_threshold_enabled Is disk allocation decider enabled.
# TYPE elasticsearch_clustersettings_allocation_threshold_enabled gauge
elasticsearch_clustersettings_allocation_threshold_enabled 0
# HELP elasticsearch_clustersettings_allocation_watermark_flood_stage_ratio Flood stage watermark as a ratio.
# TYPE elasticsearch_clustersettings_allocation_watermark_flood_stage_ratio gauge
elasticsearch_clustersettings_allocation_watermark_flood_stage_ratio 0.95
# HELP elasticsearch_clustersettings_allocation_watermark_high_ratio High watermark for disk usage as a ratio.
# TYPE elasticsearch_clustersettings_allocation_watermark_high_ratio gauge
elasticsearch_clustersettings_allocation_watermark_high_ratio 0.9
# HELP elasticsearch_clustersettings_allocation_watermark_low_ratio Low watermark for disk usage as a ratio.
# TYPE elasticsearch_clustersettings_allocation_watermark_low_ratio gauge
elasticsearch_clustersettings_allocation_watermark_low_ratio 0.85
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
# HELP elasticsearch_clustersettings_allocation_threshold_enabled Is disk allocation decider enabled.
# TYPE elasticsearch_clustersettings_allocation_threshold_enabled gauge
elasticsearch_clustersettings_allocation_threshold_enabled 1
# HELP elasticsearch_clustersettings_allocation_watermark_flood_stage_bytes Flood stage watermark as in bytes.
# TYPE elasticsearch_clustersettings_allocation_watermark_flood_stage_bytes gauge
elasticsearch_clustersettings_allocation_watermark_flood_stage_bytes 100
# HELP elasticsearch_clustersettings_allocation_watermark_high_bytes High watermark for disk usage in bytes.
# TYPE elasticsearch_clustersettings_allocation_watermark_high_bytes gauge
elasticsearch_clustersettings_allocation_watermark_high_bytes 2.147483648e+11
# HELP elasticsearch_clustersettings_allocation_watermark_low_bytes Low watermark for disk usage in bytes.
# TYPE elasticsearch_clustersettings_allocation_watermark_low_bytes gauge
elasticsearch_clustersettings_allocation_watermark_low_bytes 5.24288e+07
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

			c, err := NewClusterSettings(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_getValueInBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{name: "Bytes", input: "100b", want: 100},
		{name: "Kibibytes", input: "200kb", want: 204800},
		{name: "Mebibytes", input: "300mb", want: 314572800},
		{name: "Gibibytes", input: "400gb", want: 429496729600},
		{name: "Tebibytes", input: "500tb", want: 549755813888000},
		{name: "Pebibytes", input: "600pb", want: 675539944105574400},
		{name: "Unknown", input: "9ab", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValueInBytes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getValueInBytes() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("getValueInBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getValueAsRatio(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{name: "Ratio", input: "0.5", want: 0.5},
		{name: "Percentage", input: "50%", want: 0.5},
		{name: "Invalid", input: "500b", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValueAsRatio(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getValueAsRatio() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("getValueAsRatio() = %v, want %v", got, tt.want)
			}
		})
	}
}
