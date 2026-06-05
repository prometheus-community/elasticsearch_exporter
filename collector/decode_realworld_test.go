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
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// This file extends decode_bench_test.go with more realistic, real-world tests
// for PR #1159:
//
//   1. BenchmarkDecodeHTTP        - decodes over a real *http.Response.Body
//                                   served by httptest, i.e. the actual
//                                   production I/O path (transport-buffered
//                                   network stream) rather than a bytes.Reader.
//   2. BenchmarkDecodeSizeSweep   - synthetic but struct-typed index-stats
//                                   payloads from ~13 KB to ~6.6 MB to confirm
//                                   the decoder does not undercut ReadAll at
//                                   any size.

// ---- 1. Real HTTP path -----------------------------------------------------

func serveFixture(b *testing.B, file string) (*httptest.Server, func()) {
	b.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		b.Fatalf("read fixture %s: %v", file, err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	return ts, ts.Close
}

func BenchmarkDecodeHTTP(b *testing.B) {
	for _, tc := range decodeCases {
		ts, closeFn := serveFixture(b, tc.file)
		client := ts.Client()

		b.Run(tc.name+"/ReadAll+Unmarshal", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				resp, err := client.Get(ts.URL)
				if err != nil {
					b.Fatal(err)
				}
				buf, err := io.ReadAll(resp.Body)
				if err != nil {
					b.Fatal(err)
				}
				resp.Body.Close()
				if err := json.Unmarshal(buf, tc.newDest()); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(tc.name+"/Decoder", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				resp, err := client.Get(ts.URL)
				if err != nil {
					b.Fatal(err)
				}
				err = json.NewDecoder(resp.Body).Decode(tc.newDest())
				resp.Body.Close()
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		closeFn()
	}
}

// ---- 2. Payload-size sweep -------------------------------------------------

// buildIndicesPayload synthesizes a valid index-stats response containing n
// index entries, each a copy of the real "_all" block from the 7.17.3 fixture.
// This scales a realistic, struct-typed payload from a few KB to several MB,
// mirroring large clusters with many indices.
func buildIndicesPayload(tb testing.TB, n int) []byte {
	tb.Helper()
	raw, err := os.ReadFile("../fixtures/indices/7.17.3.json")
	if err != nil {
		tb.Fatalf("read fixture: %v", err)
	}
	var doc struct {
		Shards json.RawMessage `json:"_shards"`
		All    json.RawMessage `json:"_all"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		tb.Fatalf("unmarshal fixture: %v", err)
	}

	var sb []byte
	sb = append(sb, []byte(`{"_shards":`)...)
	sb = append(sb, doc.Shards...)
	sb = append(sb, []byte(`,"_all":`)...)
	sb = append(sb, doc.All...)
	sb = append(sb, []byte(`,"indices":{`)...)
	for i := 0; i < n; i++ {
		if i > 0 {
			sb = append(sb, ',')
		}
		sb = append(sb, []byte(fmt.Sprintf("%q:", fmt.Sprintf("index-%06d", i)))...)
		sb = append(sb, doc.All...)
	}
	sb = append(sb, []byte(`}}`)...)

	// Verify it decodes into the production struct.
	if err := json.Unmarshal(sb, &indexStatsResponse{}); err != nil {
		tb.Fatalf("synthetic payload (n=%d) does not decode: %v", n, err)
	}
	return sb
}

func BenchmarkDecodeSizeSweep(b *testing.B) {
	for _, n := range []int{1, 10, 100, 1000} {
		data := buildIndicesPayload(b, n)
		label := fmt.Sprintf("indices_%d_(%dKB)", n, len(data)/1024)

		b.Run(label+"/ReadAll+Unmarshal", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				buf, err := io.ReadAll(bytes.NewReader(data))
				if err != nil {
					b.Fatal(err)
				}
				if err := json.Unmarshal(buf, &indexStatsResponse{}); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(label+"/Decoder", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := json.NewDecoder(bytes.NewReader(data)).Decode(&indexStatsResponse{}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
