// Copyright 2022 The Prometheus Authors
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

func TestDataStream(t *testing.T) {

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "7.15.0",
			file: "../fixtures/datastream/7.15.0.json",
			want: `# HELP elasticsearch_data_stream_backing_indices_total Number of backing indices
            # TYPE elasticsearch_data_stream_backing_indices_total counter
            elasticsearch_data_stream_backing_indices_total{data_stream="bar"} 2
            elasticsearch_data_stream_backing_indices_total{data_stream="foo"} 5
            # HELP elasticsearch_data_stream_store_size_bytes Store size of data stream
            # TYPE elasticsearch_data_stream_store_size_bytes counter
            elasticsearch_data_stream_store_size_bytes{data_stream="bar"} 6.7382272e+08
            elasticsearch_data_stream_store_size_bytes{data_stream="foo"} 4.29205396e+08
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

			c := NewDataStream(promslog.NewNopLogger(), http.DefaultClient, u)
			if err != nil {
				t.Fatal(err)
			}

			if err := testutil.CollectAndCompare(c, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
