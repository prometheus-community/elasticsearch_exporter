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

func TestILMStatus(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION
	//  curl http://localhost:9200/_ilm/status
	tcs := map[string]string{
		"6.6.0": `{ "operation_mode": "RUNNING" }`,
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
		c := NewIlmStatus(log.NewNopLogger(), http.DefaultClient, u)
		chr, err := c.fetchAndDecodeIlm()
		if err != nil {
			t.Fatalf("Failed to fetch or decode ilm status: %s", err)
		}
		t.Logf("[%s] ILM Status Response: %+v", ver, chr)
		if chr.OperationMode != "RUNNING" {
			t.Errorf("Invalid ilm status")
		}
	}
}
