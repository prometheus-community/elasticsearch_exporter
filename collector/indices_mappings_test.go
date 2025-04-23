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

func TestMapping(t *testing.T) {
	// Testcases created using:
	//  docker run -p 9200:9200 -e "discovery.type=single-node" elasticsearch:7.8.0
	//  curl -XPUT http://localhost:9200/twitter
	//  curl -XPUT http://localhost:9200/facebook
	/*  curl -XPUT http://localhost:9200/twitter/_mapping -H 'Content-Type: application/json' -d'{
	    "properties": {
	        "email": {
	            "type": "keyword"
	        },
	        "phone": {
	            "type": "keyword"
	        }
	    }
	}'*/
	/*  curl -XPUT http://localhost:9200/facebook/_mapping -H 'Content-Type: application/json' -d'{
	    "properties": {
	        "name": {
	            "type": "text",
	            "fields": {
	                "raw": {
	                    "type": "keyword"
	                }
	            }
	        },
	        "contact": {
	            "properties": {
	                "email": {
	                    "type": "text",
	                    "fields": {
	                        "raw": {
	                            "type": "keyword"
	                        }
	                    }
	                },
	                "phone": {
	                    "type": "text"
	                }
	            }
	        }
	    }
	}'*/
	//  curl http://localhost:9200/_all/_mapping
	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "7.8.0",
			file: "../fixtures/indices_mappings/7.8.0.json",
			want: `
# HELP elasticsearch_indices_mappings_stats_fields Current number fields within cluster.
# TYPE elasticsearch_indices_mappings_stats_fields gauge
elasticsearch_indices_mappings_stats_fields{index="facebook"} 6
elasticsearch_indices_mappings_stats_fields{index="twitter"} 2
			`,
		},
		{
			name: "counts",
			file: "../fixtures/indices_mappings/counts.json",
			want: `
# HELP elasticsearch_indices_mappings_stats_fields Current number fields within cluster.
# TYPE elasticsearch_indices_mappings_stats_fields gauge
elasticsearch_indices_mappings_stats_fields{index="test-data-2023.01.20"} 40
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

			c := NewIndicesMappings(promslog.NewNopLogger(), http.DefaultClient, u)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(c, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
