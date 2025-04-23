// Copyright 2025 The Prometheus Authors
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

func TestHealthReport(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION
	//  curl -XPUT http://localhost:9200/twitter
	//  curl http://localhost:9200/_health_report

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "8.7.0",
			file: "../fixtures/healthreport/8.7.0.json",
			want: `
				# HELP elasticsearch_health_report_creating_primaries The number of creating primary shards
				# TYPE elasticsearch_health_report_creating_primaries gauge
				elasticsearch_health_report_creating_primaries{cluster="docker-cluster"} 0
				# HELP elasticsearch_health_report_creating_replicas The number of creating replica shards
				# TYPE elasticsearch_health_report_creating_replicas gauge
				elasticsearch_health_report_creating_replicas{cluster="docker-cluster"} 0
				# HELP elasticsearch_health_report_data_stream_lifecycle_status Data stream lifecycle status
				# TYPE elasticsearch_health_report_data_stream_lifecycle_status gauge
				elasticsearch_health_report_data_stream_lifecycle_status{cluster="docker-cluster",color="green"} 1
				elasticsearch_health_report_data_stream_lifecycle_status{cluster="docker-cluster",color="red"} 0
				elasticsearch_health_report_data_stream_lifecycle_status{cluster="docker-cluster",color="yellow"} 0
				# HELP elasticsearch_health_report_disk_status Disk status
				# TYPE elasticsearch_health_report_disk_status gauge
				elasticsearch_health_report_disk_status{cluster="docker-cluster",color="green"} 1
				elasticsearch_health_report_disk_status{cluster="docker-cluster",color="red"} 0
				elasticsearch_health_report_disk_status{cluster="docker-cluster",color="yellow"} 0
				# HELP elasticsearch_health_report_ilm_policies The number of ILM Policies
				# TYPE elasticsearch_health_report_ilm_policies gauge
				elasticsearch_health_report_ilm_policies{cluster="docker-cluster"} 17
				# HELP elasticsearch_health_report_ilm_stagnating_indices The number of stagnating indices
				# TYPE elasticsearch_health_report_ilm_stagnating_indices gauge
				elasticsearch_health_report_ilm_stagnating_indices{cluster="docker-cluster"} 0
				# HELP elasticsearch_health_report_ilm_status ILM status
				# TYPE elasticsearch_health_report_ilm_status gauge
				elasticsearch_health_report_ilm_status{cluster="docker-cluster",color="green"} 1
				elasticsearch_health_report_ilm_status{cluster="docker-cluster",color="red"} 0
				elasticsearch_health_report_ilm_status{cluster="docker-cluster",color="yellow"} 0
				# HELP elasticsearch_health_report_initializing_primaries The number of initializing primary shards
				# TYPE elasticsearch_health_report_initializing_primaries gauge
				elasticsearch_health_report_initializing_primaries{cluster="docker-cluster"} 0
				# HELP elasticsearch_health_report_initializing_replicas The number of initializing replica shards
				# TYPE elasticsearch_health_report_initializing_replicas gauge
				elasticsearch_health_report_initializing_replicas{cluster="docker-cluster"} 0
				# HELP elasticsearch_health_report_master_is_stable_status Master is stable status
				# TYPE elasticsearch_health_report_master_is_stable_status gauge
				elasticsearch_health_report_master_is_stable_status{cluster="docker-cluster",color="green"} 1
				elasticsearch_health_report_master_is_stable_status{cluster="docker-cluster",color="red"} 0
				elasticsearch_health_report_master_is_stable_status{cluster="docker-cluster",color="yellow"} 0
				# HELP elasticsearch_health_report_max_shards_in_cluster_data The number of maximum shards in a cluster
				# TYPE elasticsearch_health_report_max_shards_in_cluster_data gauge
				elasticsearch_health_report_max_shards_in_cluster_data{cluster="docker-cluster"} 13500
				# HELP elasticsearch_health_report_max_shards_in_cluster_frozen The number of maximum frozen shards in a cluster
				# TYPE elasticsearch_health_report_max_shards_in_cluster_frozen gauge
				elasticsearch_health_report_max_shards_in_cluster_frozen{cluster="docker-cluster"} 9000
				# HELP elasticsearch_health_report_repository_integrity_status Repository integrity status
				# TYPE elasticsearch_health_report_repository_integrity_status gauge
				elasticsearch_health_report_repository_integrity_status{cluster="docker-cluster",color="green"} 1
				elasticsearch_health_report_repository_integrity_status{cluster="docker-cluster",color="red"} 0
				elasticsearch_health_report_repository_integrity_status{cluster="docker-cluster",color="yellow"} 0
				# HELP elasticsearch_health_report_restarting_primaries The number of restarting primary shards
				# TYPE elasticsearch_health_report_restarting_primaries gauge
				elasticsearch_health_report_restarting_primaries{cluster="docker-cluster"} 0
				# HELP elasticsearch_health_report_restarting_replicas The number of restarting replica shards
				# TYPE elasticsearch_health_report_restarting_replicas gauge
				elasticsearch_health_report_restarting_replicas{cluster="docker-cluster"} 0
				# HELP elasticsearch_health_report_shards_availabilty_status Shards availabilty status
				# TYPE elasticsearch_health_report_shards_availabilty_status gauge
				elasticsearch_health_report_shards_availabilty_status{cluster="docker-cluster",color="green"} 1
				elasticsearch_health_report_shards_availabilty_status{cluster="docker-cluster",color="red"} 0
				elasticsearch_health_report_shards_availabilty_status{cluster="docker-cluster",color="yellow"} 0
				# HELP elasticsearch_health_report_shards_capacity_status Shards capacity status
				# TYPE elasticsearch_health_report_shards_capacity_status gauge
				elasticsearch_health_report_shards_capacity_status{cluster="docker-cluster",color="green"} 1
				elasticsearch_health_report_shards_capacity_status{cluster="docker-cluster",color="red"} 0
				elasticsearch_health_report_shards_capacity_status{cluster="docker-cluster",color="yellow"} 0
				# HELP elasticsearch_health_report_slm_policies The number of SLM policies
				# TYPE elasticsearch_health_report_slm_policies gauge
				elasticsearch_health_report_slm_policies{cluster="docker-cluster"} 0
				# HELP elasticsearch_health_report_slm_status SLM status
				# TYPE elasticsearch_health_report_slm_status gauge
				elasticsearch_health_report_slm_status{cluster="docker-cluster",color="green"} 1
				elasticsearch_health_report_slm_status{cluster="docker-cluster",color="red"} 0
				elasticsearch_health_report_slm_status{cluster="docker-cluster",color="yellow"} 0
				# HELP elasticsearch_health_report_started_primaries The number of started primary shards
				# TYPE elasticsearch_health_report_started_primaries gauge
				elasticsearch_health_report_started_primaries{cluster="docker-cluster"} 11703
				# HELP elasticsearch_health_report_started_replicas The number of started replica shards
				# TYPE elasticsearch_health_report_started_replicas gauge
				elasticsearch_health_report_started_replicas{cluster="docker-cluster"} 1701
				# HELP elasticsearch_health_report_status Overall cluster status
				# TYPE elasticsearch_health_report_status gauge
				elasticsearch_health_report_status{cluster="docker-cluster",color="green"} 1
				elasticsearch_health_report_status{cluster="docker-cluster",color="red"} 0
				elasticsearch_health_report_status{cluster="docker-cluster",color="yellow"} 0
				# HELP elasticsearch_health_report_total_repositories The number of snapshot repositories
				# TYPE elasticsearch_health_report_total_repositories gauge
				elasticsearch_health_report_total_repositories{cluster="docker-cluster"} 1
				# HELP elasticsearch_health_report_unassigned_primaries The number of unassigned primary shards
				# TYPE elasticsearch_health_report_unassigned_primaries gauge
				elasticsearch_health_report_unassigned_primaries{cluster="docker-cluster"} 0
				# HELP elasticsearch_health_report_unassigned_replicas The number of unassigned replica shards
				# TYPE elasticsearch_health_report_unassigned_replicas gauge
				elasticsearch_health_report_unassigned_replicas{cluster="docker-cluster"} 0
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

			c, err := NewHealthReport(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
