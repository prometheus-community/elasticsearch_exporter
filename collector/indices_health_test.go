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

	"github.com/go-kit/log"
)

func TestIndicesHealth(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION
	//  curl -XPUT http://localhost:9200/twitter
	//  curl http://localhost:9200/_cat/indices?format=json&h=health,index
	tcs := map[string]string{
		"1.7.6": `[{"health":"yellow","index":"twitter"}]`,
		"2.4.5": `[{"health":"yellow","index":"twitter"}]`,
		"5.4.2": `[{"health":"yellow","index":"twitter"}]`,
		"5.5.2": `[{"health":"yellow","index":"twitter"}]`,
		"8.2.3": `[{"health":"yellow","index":"twitter"}]`,
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
		c := NewIndicesHealth(log.NewNopLogger(), http.DefaultClient, u)
		ihr, err := c.fetchAndDecodeIndicesHealth()
		if err != nil {
			t.Fatalf("Failed to fetch or decode cluster health: %s", err)
		}
		t.Logf("[%s] Cluster Health Response: %+v", ver, ihr)
		if ihr[0].Index != "twitter" {
			t.Errorf("is not twitter")
		}
		if ihr[0].Health != "yellow" {
			t.Errorf("twitter is not yellow")
		}
	}
}
