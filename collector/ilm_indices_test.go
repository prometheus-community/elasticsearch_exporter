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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/log"
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
	tcs := map[string]string{
		"6.6.0": `{
			"indices": {
			  "twitter": { "index": "twitter", "managed": false },
			  "facebook": {
				"index": "facebook",
				"managed": true,
				"policy": "my_policy",
				"lifecycle_date_millis": 1660799138565,
				"phase": "new",
				"phase_time_millis": 1660799138651,
				"action": "complete",
				"action_time_millis": 1660799138651,
				"step": "complete",
				"step_time_millis": 1660799138651
			  }
			}
		  }`,
	}
	for ver, out := range tcs {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, out)
		}))
		defer ts.Close()

		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatalf("Failed to parse URL: %s", err)
		}
		c := NewIlmIndicies(log.NewNopLogger(), http.DefaultClient, u)
		chr, err := c.fetchAndDecodeIlm()
		if err != nil {
			t.Fatalf("Failed to fetch or decode indices ilm metrics: %s", err)
		}
		t.Logf("[%s] indices ilm metrics Response: %+v", ver, chr)

		if chr.Indices["twitter"].Managed != false {
			t.Errorf("Invalid ilm metrics at twitter.managed")
		}
		if chr.Indices["facebook"].Managed != true {
			t.Errorf("Invalid ilm metrics at facebook.managed")
		}
		if chr.Indices["facebook"].Phase != "new" {
			t.Errorf("Invalid ilm metrics at facebook.phase")
		}
		if chr.Indices["facebook"].Action != "complete" {
			t.Errorf("Invalid ilm metrics at facebook.action")
		}
		if chr.Indices["facebook"].Step != "complete" {
			t.Errorf("Invalid ilm metrics at facebook.step")
		}

	}
}
