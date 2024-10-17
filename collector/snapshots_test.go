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

func TestSnapshots(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine  -Des.path.repo="/tmp" (1.7.6, 2.4.5)
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine  -E path.repo="/tmp" (5.4.2)
	//  curl -XPUT http://localhost:9200/foo_1/type1/1 -d '{"title":"abc","content":"hello"}'
	//  curl -XPUT http://localhost:9200/foo_1/type1/2 -d '{"title":"def","content":"world"}'
	//  curl -XPUT http://localhost:9200/foo_2/type1/1 -d '{"title":"abc001","content":"hello001"}'
	//  curl -XPUT http://localhost:9200/foo_2/type1/2 -d '{"title":"def002","content":"world002"}'
	//  curl -XPUT http://localhost:9200/foo_2/type1/3 -d '{"title":"def003","content":"world003"}'
	//  curl -XPUT http://localhost:9200/_snapshot/test1 -d '{"type": "fs","settings":{"location": "/tmp/test1"}}'
	//  curl -XPUT "http://localhost:9200/_snapshot/test1/snapshot_1?wait_for_completion=true"
	//  curl http://localhost:9200/_snapshot/
	//  curl http://localhost:9200/_snapshot/test1/_all

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "1.7.6",
			file: "../fixtures/snapshots/1.7.6.json",
			want: `# HELP elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds Timestamp of the latest SUCCESS or PARTIAL snapshot
						# TYPE elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds gauge
						elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds{repository="test1"} 1.536052142e+09
						# HELP elasticsearch_snapshot_stats_number_of_snapshots Number of snapshots in a repository
						# TYPE elasticsearch_snapshot_stats_number_of_snapshots gauge
						elasticsearch_snapshot_stats_number_of_snapshots{repository="test1"} 1
						# HELP elasticsearch_snapshot_stats_oldest_snapshot_timestamp Timestamp of the oldest snapshot
						# TYPE elasticsearch_snapshot_stats_oldest_snapshot_timestamp gauge
						elasticsearch_snapshot_stats_oldest_snapshot_timestamp{repository="test1"} 1.536052142e+09
						# HELP elasticsearch_snapshot_stats_snapshot_end_time_timestamp Last snapshot end timestamp
						# TYPE elasticsearch_snapshot_stats_snapshot_end_time_timestamp gauge
						elasticsearch_snapshot_stats_snapshot_end_time_timestamp{repository="test1",state="SUCCESS",version="1.7.6"} 1.536052142e+09
						# HELP elasticsearch_snapshot_stats_snapshot_failed_shards Last snapshot failed shards
						# TYPE elasticsearch_snapshot_stats_snapshot_failed_shards gauge
						elasticsearch_snapshot_stats_snapshot_failed_shards{repository="test1",state="SUCCESS",version="1.7.6"} 0
						# HELP elasticsearch_snapshot_stats_snapshot_number_of_failures Last snapshot number of failures
						# TYPE elasticsearch_snapshot_stats_snapshot_number_of_failures gauge
						elasticsearch_snapshot_stats_snapshot_number_of_failures{repository="test1",state="SUCCESS",version="1.7.6"} 0
						# HELP elasticsearch_snapshot_stats_snapshot_number_of_indices Number of indices in the last snapshot
						# TYPE elasticsearch_snapshot_stats_snapshot_number_of_indices gauge
						elasticsearch_snapshot_stats_snapshot_number_of_indices{repository="test1",state="SUCCESS",version="1.7.6"} 2
						# HELP elasticsearch_snapshot_stats_snapshot_start_time_timestamp Last snapshot start timestamp
						# TYPE elasticsearch_snapshot_stats_snapshot_start_time_timestamp gauge
						elasticsearch_snapshot_stats_snapshot_start_time_timestamp{repository="test1",state="SUCCESS",version="1.7.6"} 1.536052142e+09
						# HELP elasticsearch_snapshot_stats_snapshot_successful_shards Last snapshot successful shards
						# TYPE elasticsearch_snapshot_stats_snapshot_successful_shards gauge
						elasticsearch_snapshot_stats_snapshot_successful_shards{repository="test1",state="SUCCESS",version="1.7.6"} 10
						# HELP elasticsearch_snapshot_stats_snapshot_total_shards Last snapshot total shards
						# TYPE elasticsearch_snapshot_stats_snapshot_total_shards gauge
						elasticsearch_snapshot_stats_snapshot_total_shards{repository="test1",state="SUCCESS",version="1.7.6"} 10
`,
		},
		{
			name: "2.4.5",
			file: "../fixtures/snapshots/2.4.5.json",
			want: `# HELP elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds Timestamp of the latest SUCCESS or PARTIAL snapshot
						# TYPE elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds gauge
						elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds{repository="test1"} 1.536053125e+09
						# HELP elasticsearch_snapshot_stats_number_of_snapshots Number of snapshots in a repository
						# TYPE elasticsearch_snapshot_stats_number_of_snapshots gauge
						elasticsearch_snapshot_stats_number_of_snapshots{repository="test1"} 1
						# HELP elasticsearch_snapshot_stats_oldest_snapshot_timestamp Timestamp of the oldest snapshot
						# TYPE elasticsearch_snapshot_stats_oldest_snapshot_timestamp gauge
						elasticsearch_snapshot_stats_oldest_snapshot_timestamp{repository="test1"} 1.536053125e+09
						# HELP elasticsearch_snapshot_stats_snapshot_end_time_timestamp Last snapshot end timestamp
						# TYPE elasticsearch_snapshot_stats_snapshot_end_time_timestamp gauge
						elasticsearch_snapshot_stats_snapshot_end_time_timestamp{repository="test1",state="SUCCESS",version="2.4.5"} 1.536053126e+09
						# HELP elasticsearch_snapshot_stats_snapshot_failed_shards Last snapshot failed shards
						# TYPE elasticsearch_snapshot_stats_snapshot_failed_shards gauge
						elasticsearch_snapshot_stats_snapshot_failed_shards{repository="test1",state="SUCCESS",version="2.4.5"} 0
						# HELP elasticsearch_snapshot_stats_snapshot_number_of_failures Last snapshot number of failures
						# TYPE elasticsearch_snapshot_stats_snapshot_number_of_failures gauge
						elasticsearch_snapshot_stats_snapshot_number_of_failures{repository="test1",state="SUCCESS",version="2.4.5"} 0
						# HELP elasticsearch_snapshot_stats_snapshot_number_of_indices Number of indices in the last snapshot
						# TYPE elasticsearch_snapshot_stats_snapshot_number_of_indices gauge
						elasticsearch_snapshot_stats_snapshot_number_of_indices{repository="test1",state="SUCCESS",version="2.4.5"} 2
						# HELP elasticsearch_snapshot_stats_snapshot_start_time_timestamp Last snapshot start timestamp
						# TYPE elasticsearch_snapshot_stats_snapshot_start_time_timestamp gauge
						elasticsearch_snapshot_stats_snapshot_start_time_timestamp{repository="test1",state="SUCCESS",version="2.4.5"} 1.536053125e+09
						# HELP elasticsearch_snapshot_stats_snapshot_successful_shards Last snapshot successful shards
						# TYPE elasticsearch_snapshot_stats_snapshot_successful_shards gauge
						elasticsearch_snapshot_stats_snapshot_successful_shards{repository="test1",state="SUCCESS",version="2.4.5"} 10
						# HELP elasticsearch_snapshot_stats_snapshot_total_shards Last snapshot total shards
						# TYPE elasticsearch_snapshot_stats_snapshot_total_shards gauge
						elasticsearch_snapshot_stats_snapshot_total_shards{repository="test1",state="SUCCESS",version="2.4.5"} 10
						`,
		},
		{
			name: "5.4.2",
			file: "../fixtures/snapshots/5.4.2.json",
			want: `# HELP elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds Timestamp of the latest SUCCESS or PARTIAL snapshot
						# TYPE elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds gauge
						elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds{repository="test1"} 1.536053353e+09
						# HELP elasticsearch_snapshot_stats_number_of_snapshots Number of snapshots in a repository
						# TYPE elasticsearch_snapshot_stats_number_of_snapshots gauge
						elasticsearch_snapshot_stats_number_of_snapshots{repository="test1"} 1
						# HELP elasticsearch_snapshot_stats_oldest_snapshot_timestamp Timestamp of the oldest snapshot
						# TYPE elasticsearch_snapshot_stats_oldest_snapshot_timestamp gauge
						elasticsearch_snapshot_stats_oldest_snapshot_timestamp{repository="test1"} 1.536053353e+09
						# HELP elasticsearch_snapshot_stats_snapshot_end_time_timestamp Last snapshot end timestamp
						# TYPE elasticsearch_snapshot_stats_snapshot_end_time_timestamp gauge
						elasticsearch_snapshot_stats_snapshot_end_time_timestamp{repository="test1",state="SUCCESS",version="5.4.2"} 1.536053354e+09
						# HELP elasticsearch_snapshot_stats_snapshot_failed_shards Last snapshot failed shards
						# TYPE elasticsearch_snapshot_stats_snapshot_failed_shards gauge
						elasticsearch_snapshot_stats_snapshot_failed_shards{repository="test1",state="SUCCESS",version="5.4.2"} 0
						# HELP elasticsearch_snapshot_stats_snapshot_number_of_failures Last snapshot number of failures
						# TYPE elasticsearch_snapshot_stats_snapshot_number_of_failures gauge
						elasticsearch_snapshot_stats_snapshot_number_of_failures{repository="test1",state="SUCCESS",version="5.4.2"} 0
						# HELP elasticsearch_snapshot_stats_snapshot_number_of_indices Number of indices in the last snapshot
						# TYPE elasticsearch_snapshot_stats_snapshot_number_of_indices gauge
						elasticsearch_snapshot_stats_snapshot_number_of_indices{repository="test1",state="SUCCESS",version="5.4.2"} 2
						# HELP elasticsearch_snapshot_stats_snapshot_start_time_timestamp Last snapshot start timestamp
						# TYPE elasticsearch_snapshot_stats_snapshot_start_time_timestamp gauge
						elasticsearch_snapshot_stats_snapshot_start_time_timestamp{repository="test1",state="SUCCESS",version="5.4.2"} 1.536053353e+09
						# HELP elasticsearch_snapshot_stats_snapshot_successful_shards Last snapshot successful shards
						# TYPE elasticsearch_snapshot_stats_snapshot_successful_shards gauge
						elasticsearch_snapshot_stats_snapshot_successful_shards{repository="test1",state="SUCCESS",version="5.4.2"} 10
						# HELP elasticsearch_snapshot_stats_snapshot_total_shards Last snapshot total shards
						# TYPE elasticsearch_snapshot_stats_snapshot_total_shards gauge
						elasticsearch_snapshot_stats_snapshot_total_shards{repository="test1",state="SUCCESS",version="5.4.2"} 10
						`,
		},
		{
			name: "5.4.2-failure",
			file: "../fixtures/snapshots/5.4.2-failed.json",
			want: `# HELP elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds Timestamp of the latest SUCCESS or PARTIAL snapshot
						# TYPE elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds gauge
						elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds{repository="test1"} 1.536053353e+09
						# HELP elasticsearch_snapshot_stats_number_of_snapshots Number of snapshots in a repository
						# TYPE elasticsearch_snapshot_stats_number_of_snapshots gauge
						elasticsearch_snapshot_stats_number_of_snapshots{repository="test1"} 1
						# HELP elasticsearch_snapshot_stats_oldest_snapshot_timestamp Timestamp of the oldest snapshot
						# TYPE elasticsearch_snapshot_stats_oldest_snapshot_timestamp gauge
						elasticsearch_snapshot_stats_oldest_snapshot_timestamp{repository="test1"} 1.536053353e+09
						# HELP elasticsearch_snapshot_stats_snapshot_end_time_timestamp Last snapshot end timestamp
						# TYPE elasticsearch_snapshot_stats_snapshot_end_time_timestamp gauge
						elasticsearch_snapshot_stats_snapshot_end_time_timestamp{repository="test1",state="SUCCESS",version="5.4.2"} 1.536053354e+09
						# HELP elasticsearch_snapshot_stats_snapshot_failed_shards Last snapshot failed shards
						# TYPE elasticsearch_snapshot_stats_snapshot_failed_shards gauge
						elasticsearch_snapshot_stats_snapshot_failed_shards{repository="test1",state="SUCCESS",version="5.4.2"} 1
						# HELP elasticsearch_snapshot_stats_snapshot_number_of_failures Last snapshot number of failures
						# TYPE elasticsearch_snapshot_stats_snapshot_number_of_failures gauge
						elasticsearch_snapshot_stats_snapshot_number_of_failures{repository="test1",state="SUCCESS",version="5.4.2"} 1
						# HELP elasticsearch_snapshot_stats_snapshot_number_of_indices Number of indices in the last snapshot
						# TYPE elasticsearch_snapshot_stats_snapshot_number_of_indices gauge
						elasticsearch_snapshot_stats_snapshot_number_of_indices{repository="test1",state="SUCCESS",version="5.4.2"} 2
						# HELP elasticsearch_snapshot_stats_snapshot_start_time_timestamp Last snapshot start timestamp
						# TYPE elasticsearch_snapshot_stats_snapshot_start_time_timestamp gauge
						elasticsearch_snapshot_stats_snapshot_start_time_timestamp{repository="test1",state="SUCCESS",version="5.4.2"} 1.536053353e+09
						# HELP elasticsearch_snapshot_stats_snapshot_successful_shards Last snapshot successful shards
						# TYPE elasticsearch_snapshot_stats_snapshot_successful_shards gauge
						elasticsearch_snapshot_stats_snapshot_successful_shards{repository="test1",state="SUCCESS",version="5.4.2"} 10
						# HELP elasticsearch_snapshot_stats_snapshot_total_shards Last snapshot total shards
						# TYPE elasticsearch_snapshot_stats_snapshot_total_shards gauge
						elasticsearch_snapshot_stats_snapshot_total_shards{repository="test1",state="SUCCESS",version="5.4.2"} 10
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
				if r.RequestURI == "/_snapshot" {
					fmt.Fprint(w, `{"test1":{"type":"fs","settings":{"location":"/tmp/test1"}}}`)
					return
				}
				io.Copy(w, f)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			c, err := NewSnapshots(promslog.NewNopLogger(), u, http.DefaultClient)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
