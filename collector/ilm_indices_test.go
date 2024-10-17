// Copyright 2023 The Prometheus Authors
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

func TestILMMetrics(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION
	//  curl -XPUT http://localhost:9200/twitter
	// 	curl -X PUT "localhost:9200/_ilm/policy/my_policy?pretty" -H 'Content-Type: application/json' -d'
	// 	{
	// 	  "policy": {
	// 		"phases": {
	// 		  "warm": {
	// 			"min_age": "10d",
	// 			"actions": {
	// 			  "forcemerge": {
	// 				"max_num_segments": 1
	// 			  }
	// 			}
	// 		  },
	// 		  "delete": {
	// 			"min_age": "30d",
	// 			"actions": {
	// 			  "delete": {}
	// 			}
	// 		  }
	// 		}
	// 	  }
	// 	}
	// 	'
	// 	curl -X PUT "localhost:9200/facebook?pretty" -H 'Content-Type: application/json' -d'
	// 	{
	// 	"settings": {
	// 		"index": {
	// 		"lifecycle": {
	// 			"name": "my_policy"
	// 		}
	// 		}
	// 	}
	// 	}
	// 	'
	//  curl http://localhost:9200/_all/_ilm/explain
	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "6.6.0",
			file: "../fixtures/ilm_indices/6.6.0.json",
			want: `
# HELP elasticsearch_ilm_index_status Status of ILM policy for index
# TYPE elasticsearch_ilm_index_status gauge
elasticsearch_ilm_index_status{action="",index="twitter",phase="",step=""} 0
elasticsearch_ilm_index_status{action="complete",index="facebook",phase="new",step="complete"} 1
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

			c := NewIlmIndicies(promslog.NewNopLogger(), http.DefaultClient, u)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(c, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
