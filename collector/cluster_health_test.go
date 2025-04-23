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

func TestClusterHealth(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine
	//  curl -XPUT http://localhost:9200/twitter
	//  curl http://localhost:9200/_cluster/health

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "1.7.6",
			file: "../fixtures/clusterhealth/1.7.6.json",
			want: `
				# HELP elasticsearch_cluster_health_active_primary_shards The number of primary shards in your cluster. This is an aggregate total across all indices.
				# TYPE elasticsearch_cluster_health_active_primary_shards gauge
				elasticsearch_cluster_health_active_primary_shards{cluster="elasticsearch"} 5
				# HELP elasticsearch_cluster_health_active_shards Aggregate total of all shards across all indices, which includes replica shards.
				# TYPE elasticsearch_cluster_health_active_shards gauge
				elasticsearch_cluster_health_active_shards{cluster="elasticsearch"} 5
				# HELP elasticsearch_cluster_health_delayed_unassigned_shards Shards delayed to reduce reallocation overhead
				# TYPE elasticsearch_cluster_health_delayed_unassigned_shards gauge
				elasticsearch_cluster_health_delayed_unassigned_shards{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_initializing_shards Count of shards that are being freshly created.
				# TYPE elasticsearch_cluster_health_initializing_shards gauge
				elasticsearch_cluster_health_initializing_shards{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_number_of_data_nodes Number of data nodes in the cluster.
				# TYPE elasticsearch_cluster_health_number_of_data_nodes gauge
				elasticsearch_cluster_health_number_of_data_nodes{cluster="elasticsearch"} 1
				# HELP elasticsearch_cluster_health_number_of_in_flight_fetch The number of ongoing shard info requests.
				# TYPE elasticsearch_cluster_health_number_of_in_flight_fetch gauge
				elasticsearch_cluster_health_number_of_in_flight_fetch{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_number_of_nodes Number of nodes in the cluster.
				# TYPE elasticsearch_cluster_health_number_of_nodes gauge
				elasticsearch_cluster_health_number_of_nodes{cluster="elasticsearch"} 1
				# HELP elasticsearch_cluster_health_number_of_pending_tasks Cluster level changes which have not yet been executed
				# TYPE elasticsearch_cluster_health_number_of_pending_tasks gauge
				elasticsearch_cluster_health_number_of_pending_tasks{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_relocating_shards The number of shards that are currently moving from one node to another node.
				# TYPE elasticsearch_cluster_health_relocating_shards gauge
				elasticsearch_cluster_health_relocating_shards{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_status Whether all primary and replica shards are allocated.
				# TYPE elasticsearch_cluster_health_status gauge
				elasticsearch_cluster_health_status{cluster="elasticsearch",color="green"} 0
				elasticsearch_cluster_health_status{cluster="elasticsearch",color="red"} 0
				elasticsearch_cluster_health_status{cluster="elasticsearch",color="yellow"} 1
				# HELP elasticsearch_cluster_health_task_max_waiting_in_queue_millis Tasks max time waiting in queue.
				# TYPE elasticsearch_cluster_health_task_max_waiting_in_queue_millis gauge
				elasticsearch_cluster_health_task_max_waiting_in_queue_millis{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_unassigned_shards The number of shards that exist in the cluster state, but cannot be found in the cluster itself.
				# TYPE elasticsearch_cluster_health_unassigned_shards gauge
				elasticsearch_cluster_health_unassigned_shards{cluster="elasticsearch"} 5
      `,
		},
		{
			name: "2.4.5",
			file: "../fixtures/clusterhealth/2.4.5.json",
			want: `
				# HELP elasticsearch_cluster_health_active_primary_shards The number of primary shards in your cluster. This is an aggregate total across all indices.
				# TYPE elasticsearch_cluster_health_active_primary_shards gauge
				elasticsearch_cluster_health_active_primary_shards{cluster="elasticsearch"} 5
				# HELP elasticsearch_cluster_health_active_shards Aggregate total of all shards across all indices, which includes replica shards.
				# TYPE elasticsearch_cluster_health_active_shards gauge
				elasticsearch_cluster_health_active_shards{cluster="elasticsearch"} 5
				# HELP elasticsearch_cluster_health_delayed_unassigned_shards Shards delayed to reduce reallocation overhead
				# TYPE elasticsearch_cluster_health_delayed_unassigned_shards gauge
				elasticsearch_cluster_health_delayed_unassigned_shards{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_initializing_shards Count of shards that are being freshly created.
				# TYPE elasticsearch_cluster_health_initializing_shards gauge
				elasticsearch_cluster_health_initializing_shards{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_number_of_data_nodes Number of data nodes in the cluster.
				# TYPE elasticsearch_cluster_health_number_of_data_nodes gauge
				elasticsearch_cluster_health_number_of_data_nodes{cluster="elasticsearch"} 1
				# HELP elasticsearch_cluster_health_number_of_in_flight_fetch The number of ongoing shard info requests.
				# TYPE elasticsearch_cluster_health_number_of_in_flight_fetch gauge
				elasticsearch_cluster_health_number_of_in_flight_fetch{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_number_of_nodes Number of nodes in the cluster.
				# TYPE elasticsearch_cluster_health_number_of_nodes gauge
				elasticsearch_cluster_health_number_of_nodes{cluster="elasticsearch"} 1
				# HELP elasticsearch_cluster_health_number_of_pending_tasks Cluster level changes which have not yet been executed
				# TYPE elasticsearch_cluster_health_number_of_pending_tasks gauge
				elasticsearch_cluster_health_number_of_pending_tasks{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_relocating_shards The number of shards that are currently moving from one node to another node.
				# TYPE elasticsearch_cluster_health_relocating_shards gauge
				elasticsearch_cluster_health_relocating_shards{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_status Whether all primary and replica shards are allocated.
				# TYPE elasticsearch_cluster_health_status gauge
				elasticsearch_cluster_health_status{cluster="elasticsearch",color="green"} 0
				elasticsearch_cluster_health_status{cluster="elasticsearch",color="red"} 0
				elasticsearch_cluster_health_status{cluster="elasticsearch",color="yellow"} 1
				# HELP elasticsearch_cluster_health_task_max_waiting_in_queue_millis Tasks max time waiting in queue.
				# TYPE elasticsearch_cluster_health_task_max_waiting_in_queue_millis gauge
				elasticsearch_cluster_health_task_max_waiting_in_queue_millis{cluster="elasticsearch"} 12
				# HELP elasticsearch_cluster_health_unassigned_shards The number of shards that exist in the cluster state, but cannot be found in the cluster itself.
				# TYPE elasticsearch_cluster_health_unassigned_shards gauge
				elasticsearch_cluster_health_unassigned_shards{cluster="elasticsearch"} 5
      `,
		},
		{
			name: "5.4.2",
			file: "../fixtures/clusterhealth/5.4.2.json",
			want: `
				# HELP elasticsearch_cluster_health_active_primary_shards The number of primary shards in your cluster. This is an aggregate total across all indices.
				# TYPE elasticsearch_cluster_health_active_primary_shards gauge
				elasticsearch_cluster_health_active_primary_shards{cluster="elasticsearch"} 5
				# HELP elasticsearch_cluster_health_active_shards Aggregate total of all shards across all indices, which includes replica shards.
				# TYPE elasticsearch_cluster_health_active_shards gauge
				elasticsearch_cluster_health_active_shards{cluster="elasticsearch"} 5
				# HELP elasticsearch_cluster_health_delayed_unassigned_shards Shards delayed to reduce reallocation overhead
				# TYPE elasticsearch_cluster_health_delayed_unassigned_shards gauge
				elasticsearch_cluster_health_delayed_unassigned_shards{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_initializing_shards Count of shards that are being freshly created.
				# TYPE elasticsearch_cluster_health_initializing_shards gauge
				elasticsearch_cluster_health_initializing_shards{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_number_of_data_nodes Number of data nodes in the cluster.
				# TYPE elasticsearch_cluster_health_number_of_data_nodes gauge
				elasticsearch_cluster_health_number_of_data_nodes{cluster="elasticsearch"} 1
				# HELP elasticsearch_cluster_health_number_of_in_flight_fetch The number of ongoing shard info requests.
				# TYPE elasticsearch_cluster_health_number_of_in_flight_fetch gauge
				elasticsearch_cluster_health_number_of_in_flight_fetch{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_number_of_nodes Number of nodes in the cluster.
				# TYPE elasticsearch_cluster_health_number_of_nodes gauge
				elasticsearch_cluster_health_number_of_nodes{cluster="elasticsearch"} 1
				# HELP elasticsearch_cluster_health_number_of_pending_tasks Cluster level changes which have not yet been executed
				# TYPE elasticsearch_cluster_health_number_of_pending_tasks gauge
				elasticsearch_cluster_health_number_of_pending_tasks{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_relocating_shards The number of shards that are currently moving from one node to another node.
				# TYPE elasticsearch_cluster_health_relocating_shards gauge
				elasticsearch_cluster_health_relocating_shards{cluster="elasticsearch"} 0
				# HELP elasticsearch_cluster_health_status Whether all primary and replica shards are allocated.
				# TYPE elasticsearch_cluster_health_status gauge
				elasticsearch_cluster_health_status{cluster="elasticsearch",color="green"} 0
				elasticsearch_cluster_health_status{cluster="elasticsearch",color="red"} 0
				elasticsearch_cluster_health_status{cluster="elasticsearch",color="yellow"} 1
				# HELP elasticsearch_cluster_health_task_max_waiting_in_queue_millis Tasks max time waiting in queue.
				# TYPE elasticsearch_cluster_health_task_max_waiting_in_queue_millis gauge
				elasticsearch_cluster_health_task_max_waiting_in_queue_millis{cluster="elasticsearch"} 12
				# HELP elasticsearch_cluster_health_unassigned_shards The number of shards that exist in the cluster state, but cannot be found in the cluster itself.
				# TYPE elasticsearch_cluster_health_unassigned_shards gauge
				elasticsearch_cluster_health_unassigned_shards{cluster="elasticsearch"} 5
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

			c := NewClusterHealth(promslog.NewNopLogger(), http.DefaultClient, u)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(c, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
