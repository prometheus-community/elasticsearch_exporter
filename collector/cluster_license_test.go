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

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestIndicesHealth(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION
	//  curl http://localhost:9200/_license
	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "basic",
			file: "../fixtures/clusterlicense/basic.json",
			want: `
            # HELP elasticsearch_cluster_license_expiry_date_in_millis License expiry date in milliseconds
            # TYPE elasticsearch_cluster_license_expiry_date_in_millis gauge
            elasticsearch_cluster_license_expiry_date_in_millis{cluster_license_type="basic"} 0
            # HELP elasticsearch_cluster_license_issue_date_in_millis License issue date in milliseconds
            # TYPE elasticsearch_cluster_license_issue_date_in_millis gauge
            elasticsearch_cluster_license_issue_date_in_millis{cluster_license_type="basic"} 1.702196247064e+12
            # HELP elasticsearch_cluster_license_max_nodes The max amount of nodes allowed by the license
            # TYPE elasticsearch_cluster_license_max_nodes gauge
            elasticsearch_cluster_license_max_nodes{cluster_license_type="basic"} 1000
            # HELP elasticsearch_cluster_license_start_date_in_millis License start date in milliseconds
            # TYPE elasticsearch_cluster_license_start_date_in_millis gauge
            elasticsearch_cluster_license_start_date_in_millis{cluster_license_type="basic"} -1
            `,
		},
		{
			name: "platinum",
			file: "../fixtures/clusterlicense/platinum.json",
			want: `
            # HELP elasticsearch_cluster_license_expiry_date_in_millis License expiry date in milliseconds
            # TYPE elasticsearch_cluster_license_expiry_date_in_millis gauge
            elasticsearch_cluster_license_expiry_date_in_millis{cluster_license_type="platinum"} 1.714521599999e+12
            # HELP elasticsearch_cluster_license_issue_date_in_millis License issue date in milliseconds
            # TYPE elasticsearch_cluster_license_issue_date_in_millis gauge
            elasticsearch_cluster_license_issue_date_in_millis{cluster_license_type="platinum"} 1.6192224e+12
            # HELP elasticsearch_cluster_license_max_nodes The max amount of nodes allowed by the license
            # TYPE elasticsearch_cluster_license_max_nodes gauge
            elasticsearch_cluster_license_max_nodes{cluster_license_type="platinum"} 10
            # HELP elasticsearch_cluster_license_start_date_in_millis License start date in milliseconds
            # TYPE elasticsearch_cluster_license_start_date_in_millis gauge
            elasticsearch_cluster_license_start_date_in_millis{cluster_license_type="platinum"} 1.6192224e+12
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

			c := NewClusterLicense(log.NewNopLogger(), http.DefaultClient, u)

			if err := testutil.CollectAndCompare(c, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
