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

func TestClusterLicense(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION
	//  curl http://localhost:9200/_license
	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "7.17.10-basic",
			file: "../fixtures/clusterlicense/7.17.10-basic.json",
			want: `
            # HELP elasticsearch_cluster_license_expiry_date_seconds License expiry date since unix epoch in seconds.
            # TYPE elasticsearch_cluster_license_expiry_date_seconds gauge
            elasticsearch_cluster_license_expiry_date_seconds{issued_to="redacted",issuer="elasticsearch",status="active",type="basic"} 0
            # HELP elasticsearch_cluster_license_issue_date_seconds License issue date since unix epoch in seconds.
            # TYPE elasticsearch_cluster_license_issue_date_seconds gauge
            elasticsearch_cluster_license_issue_date_seconds{issued_to="redacted",issuer="elasticsearch",status="active",type="basic"} 1.702196247e+09
            # HELP elasticsearch_cluster_license_max_nodes The max amount of nodes allowed by the license.
            # TYPE elasticsearch_cluster_license_max_nodes gauge
            elasticsearch_cluster_license_max_nodes{issued_to="redacted",issuer="elasticsearch",status="active",type="basic"} 1000
            # HELP elasticsearch_cluster_license_start_date_seconds License start date since unix epoch in seconds.
            # TYPE elasticsearch_cluster_license_start_date_seconds gauge
            elasticsearch_cluster_license_start_date_seconds{issued_to="redacted",issuer="elasticsearch",status="active",type="basic"} 0
            `,
		},
		{
			name: "7.17.10-platinum",
			file: "../fixtures/clusterlicense/7.17.10-platinum.json",
			want: `
            # HELP elasticsearch_cluster_license_expiry_date_seconds License expiry date since unix epoch in seconds.
            # TYPE elasticsearch_cluster_license_expiry_date_seconds gauge
            elasticsearch_cluster_license_expiry_date_seconds{issued_to="redacted",issuer="API",status="active",type="platinum"} 1.714521599e+09
            # HELP elasticsearch_cluster_license_issue_date_seconds License issue date since unix epoch in seconds.
            # TYPE elasticsearch_cluster_license_issue_date_seconds gauge
            elasticsearch_cluster_license_issue_date_seconds{issued_to="redacted",issuer="API",status="active",type="platinum"} 1.6192224e+09
            # HELP elasticsearch_cluster_license_max_nodes The max amount of nodes allowed by the license.
            # TYPE elasticsearch_cluster_license_max_nodes gauge
            elasticsearch_cluster_license_max_nodes{issued_to="redacted",issuer="API",status="active",type="platinum"} 10
            # HELP elasticsearch_cluster_license_start_date_seconds License start date since unix epoch in seconds.
            # TYPE elasticsearch_cluster_license_start_date_seconds gauge
            elasticsearch_cluster_license_start_date_seconds{issued_to="redacted",issuer="API",status="active",type="platinum"} 1.6192224e+09
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

			c, err := NewClusterLicense(log.NewNopLogger(), u, http.DefaultClient)

			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(wrapCollector{c}, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
