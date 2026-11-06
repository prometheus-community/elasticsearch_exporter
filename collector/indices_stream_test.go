// Copyright The Prometheus Authors
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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

// TestStreamIndexStatsEquivalence verifies that decoding /_all/_stats one index
// at a time (streamIndexStats) yields exactly the same per-index data as the
// previous whole-response json.Unmarshal into indexStatsResponse.
func TestStreamIndexStatsEquivalence(t *testing.T) {
	for _, file := range []string{
		"../fixtures/indices/7.17.3.json",
		"../fixtures/indices/shards/7.17.3.json",
	} {
		t.Run(file, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}

			var want indexStatsResponse
			if err := json.Unmarshal(data, &want); err != nil {
				t.Fatal(err)
			}

			got := map[string]IndexStatsIndexResponse{}
			if err := streamIndexStats(bytes.NewReader(data), func(name string, s IndexStatsIndexResponse) {
				got[name] = s
			}); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, want.Indices) {
				t.Fatalf("streamed index stats differ from json.Unmarshal result")
			}
		})
	}
}

// BenchmarkIndexStatsDecode contrasts the previous approach (read whole body,
// json.Unmarshal into the full index-stats map, then range it) with the
// streaming decode now used by the collector. The streaming path never
// materializes the read buffer or the full map, so b.ReportAllocs() shows a
// large drop in bytes/op (the metric that drives GC pressure). Peak live heap
// drops by a similar factor since only one index is held at a time.
func BenchmarkIndexStatsDecode(b *testing.B) {
	for _, n := range []int{100, 1000} {
		data := buildIndicesPayload(b, n)
		label := fmt.Sprintf("indices_%d_(%dKB)", n, len(data)/1024)

		b.Run(label+"/readall_unmarshal_map", func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			var sink float64
			for i := 0; i < b.N; i++ {
				buf, err := io.ReadAll(bytes.NewReader(data))
				if err != nil {
					b.Fatal(err)
				}
				var isr indexStatsResponse
				if err := json.Unmarshal(buf, &isr); err != nil {
					b.Fatal(err)
				}
				for _, s := range isr.Indices {
					sink += float64(s.Primaries.Docs.Count)
				}
			}
			_ = sink
		})

		b.Run(label+"/streaming", func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			var sink float64
			for i := 0; i < b.N; i++ {
				if err := streamIndexStats(bytes.NewReader(data), func(_ string, s IndexStatsIndexResponse) {
					sink += float64(s.Primaries.Docs.Count)
				}); err != nil {
					b.Fatal(err)
				}
			}
			_ = sink
		})
	}
}
