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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/log"
)

func TestDataStream(t *testing.T) {
	tcs := map[string]string{
		"7.15.0": `{"_shards":{"total":30,"successful":30,"failed":0},"data_stream_count":2,"backing_indices":7,"total_store_size_bytes":1103028116,"data_streams":[{"data_stream":"foo","backing_indices":5,"store_size_bytes":429205396,"maximum_timestamp":1656079894000},{"data_stream":"bar","backing_indices":2,"store_size_bytes":673822720,"maximum_timestamp":1656028796000}]}`,
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
		s := NewDataStream(log.NewNopLogger(), http.DefaultClient, u)
		stats, err := s.fetchAndDecodeDataStreamStats()
		if err != nil {
			t.Fatalf("Failed to fetch or decode data stream stats: %s", err)
		}
		t.Logf("[%s] Data Stream Response: %+v", ver, stats)
		dataStreamStats := stats.DataStreamStats[0]

		if dataStreamStats.BackingIndices != 5 {
			t.Errorf("Bad number of backing indices")
		}

		if dataStreamStats.StoreSizeBytes != 429205396 {
			t.Errorf("Bad store size bytes valuee")
		}
	}

}
