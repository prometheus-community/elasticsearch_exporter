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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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

	tcs := map[string]string{
		"6.5.4": `{"viber":{"settings":{"index":{"creation_date":"1618593207186","number_of_shards":"5","number_of_replicas":"1","uuid":"lWg86KTARzO3r7lELytT1Q","version":{"created":"6050499"},"provided_name":"viber"}}},"instagram":{"settings":{"index":{"mapping":{"total_fields":{"limit":"10000"}},"number_of_shards":"5","blocks":{"read_only_allow_delete":"true"},"provided_name":"instagram","creation_date":"1618593203353","number_of_replicas":"1","uuid":"msb6eG7aT8GmNe-a4oyVtQ","version":{"created":"6050499"}}}},"twitter":{"settings":{"index":{"number_of_shards":"5","blocks":{"read_only_allow_delete":"true"},"provided_name":"twitter","creation_date":"1618593193641","number_of_replicas":"1","uuid":"YRUT8t4aSkKsNmGl7K3y4Q","version":{"created":"6050499"}}}},"facebook":{"settings":{"index":{"creation_date":"1618593199101","number_of_shards":"5","number_of_replicas":"1","uuid":"trZhb_YOTV-RWKitTYw81A","version":{"created":"6050499"},"provided_name":"facebook"}}}}`,
	}
	for ver, out := range tcs {
		for hn, handler := range map[string]http.Handler{
			"plain": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, out)
			}),
		} {
			ts := httptest.NewServer(handler)
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}
			c := NewIndicesSettings(promslog.NewNopLogger(), http.DefaultClient, u)
			nsr, err := c.fetchAndDecodeIndicesSettings()
			if err != nil {
				t.Fatalf("Failed to fetch or decode indices settings: %s", err)
			}
			t.Logf("[%s/%s] All Indices Settings Response: %+v", hn, ver, nsr)
			// if nsr.Cluster.Routing.Allocation.Enabled != "ALL" {
			// 	t.Errorf("Wrong setting for cluster routing allocation enabled")
			// }
			var counter int
			var totalFields int
			for key, value := range nsr {
				if value.Settings.IndexInfo.Blocks.ReadOnly == "true" {
					counter++
					if key != "instagram" && key != "twitter" {
						t.Errorf("Wrong read_only index")
					}
				}
				if value.Settings.IndexInfo.Mapping.TotalFields.Limit == "10000" {
					totalFields++
					if key != "instagram" {
						t.Errorf("Expected 10000 total_fields only for  instagram")
					}
				}
			}
			if counter != 2 {
				t.Errorf("Wrong number of read_only indexes")
			}
			if totalFields != 1 {
				t.Errorf(("Wrong number of total_fields found"))
			}
		}
	}
}
