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
	"path"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func TestIndicesSettings(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION
	// curl -XPUT http://localhost:9200/twitter
	// curl -XPUT http://localhost:9200/facebook
	// curl -XPUT http://localhost:9200/instagram
	// curl -XPUT http://localhost:9200/viber
	// curl -XPUT http://localhost:9200/instagram/_settings --header "Content-Type: application/json" -d '
	// {
	//     "index": {
	//         "mapping": {
	// 			"total_fields": {
	// 				"limit": 10000
	// 			}
	// 		},
	//         "blocks": {
	//         "read_only_allow_delete": "true"
	//         }
	//     }
	// }'
	// curl -XPUT http://localhost:9200/twitter/_settings --header "Content-Type: application/json" -d '
	// {
	//     "index": {
	//         "blocks": {
	//         "read_only_allow_delete": "true"
	//         }
	//     }
	// }'

	// curl http://localhost:9200/_all/_settings

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "6.5.4",
			file: "6.5.4.json",
			want: `# HELP elasticsearch_indices_settings_creation_timestamp_seconds index setting creation_date
             # TYPE elasticsearch_indices_settings_creation_timestamp_seconds gauge
             elasticsearch_indices_settings_creation_timestamp_seconds{index="facebook"} 1.618593199101e+09
             elasticsearch_indices_settings_creation_timestamp_seconds{index="instagram"} 1.618593203353e+09
             elasticsearch_indices_settings_creation_timestamp_seconds{index="twitter"} 1.618593193641e+09
             elasticsearch_indices_settings_creation_timestamp_seconds{index="viber"} 1.618593207186e+09
             # HELP elasticsearch_indices_settings_replicas index setting number_of_replicas
             # TYPE elasticsearch_indices_settings_replicas gauge
             elasticsearch_indices_settings_replicas{index="facebook"} 1
             elasticsearch_indices_settings_replicas{index="instagram"} 1
             elasticsearch_indices_settings_replicas{index="twitter"} 1
             elasticsearch_indices_settings_replicas{index="viber"} 1
             # HELP elasticsearch_indices_settings_stats_read_only_indices Current number of read only indices within cluster
             # TYPE elasticsearch_indices_settings_stats_read_only_indices gauge
             elasticsearch_indices_settings_stats_read_only_indices 2
             # HELP elasticsearch_indices_settings_total_fields index mapping setting for total_fields
             # TYPE elasticsearch_indices_settings_total_fields gauge
             elasticsearch_indices_settings_total_fields{index="facebook"} 1000
             elasticsearch_indices_settings_total_fields{index="instagram"} 10000
             elasticsearch_indices_settings_total_fields{index="twitter"} 1000
             elasticsearch_indices_settings_total_fields{index="viber"} 1000
						`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(path.Join("../fixtures/indices_settings", tt.file))
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

			c := NewIndicesSettings(promslog.NewNopLogger(), http.DefaultClient, u)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(c, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
