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
	"io"
	"os"
	"testing"
	"testing/iotest"
)

// These benchmarks compare the two decoding strategies discussed in PR #1159:
//
//	A) io.ReadAll(body) followed by json.Unmarshal(buf, target)   -- the old code
//	B) json.NewDecoder(body).Decode(target)                       -- the new code
//
// Both are measured against real Elasticsearch API fixtures, decoding into the
// exact response structs the collectors use in production. b.ReportAllocs()
// records bytes/op and allocs/op so the memory claim can be checked empirically.
//
// To remove the network from the measurement (and make every iteration
// identical) the body is replaced with an in-memory reader over the fixture
// bytes. We test two reader shapes:
//
//   - bytes.Reader  : the whole body is available in one Read (best case for
//                     io.ReadAll, which can then size its buffer in few growths)
//   - chunked reader: the body arrives in small pieces, like a real TCP stream,
//                     forcing io.ReadAll and the decoder to grow incrementally.

type decodeCase struct {
	name    string
	file    string
	newDest func() any
}

var decodeCases = []decodeCase{
	{
		name:    "nodestats_7.13.1_18k",
		file:    "../fixtures/nodestats/7.13.1.json",
		newDest: func() any { return &nodeStatsResponse{} },
	},
	{
		name:    "indices_7.17.3_36k",
		file:    "../fixtures/indices/7.17.3.json",
		newDest: func() any { return &indexStatsResponse{} },
	},
	{
		name:    "indices_shards_7.17.3_58k",
		file:    "../fixtures/indices/shards/7.17.3.json",
		newDest: func() any { return &indexStatsResponse{} },
	},
}

func mustReadFixture(b *testing.B, file string) []byte {
	b.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		b.Fatalf("read fixture %s: %v", file, err)
	}
	return data
}

// streamedReader yields the payload one byte per Read to emulate a streaming
// network body rather than a single in-memory blob -- the worst case for
// incremental buffer growth in both strategies.
func streamedReader(data []byte) io.Reader {
	return iotest.OneByteReader(bytes.NewReader(data))
}

func BenchmarkDecode(b *testing.B) {
	for _, tc := range decodeCases {
		data := mustReadFixture(b, tc.file)

		// Sanity: both strategies must produce a successful decode of the fixture.
		if err := json.Unmarshal(data, tc.newDest()); err != nil {
			b.Fatalf("%s: fixture does not unmarshal into target: %v", tc.name, err)
		}

		b.Run(tc.name+"/whole_body", func(b *testing.B) {
			b.Run("ReadAll+Unmarshal", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					r := bytes.NewReader(data)
					buf, err := io.ReadAll(r)
					if err != nil {
						b.Fatal(err)
					}
					if err := json.Unmarshal(buf, tc.newDest()); err != nil {
						b.Fatal(err)
					}
				}
			})
			b.Run("Decoder", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					r := bytes.NewReader(data)
					if err := json.NewDecoder(r).Decode(tc.newDest()); err != nil {
						b.Fatal(err)
					}
				}
			})
		})

		b.Run(tc.name+"/streamed_body", func(b *testing.B) {
			b.Run("ReadAll+Unmarshal", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					r := streamedReader(data)
					buf, err := io.ReadAll(r)
					if err != nil {
						b.Fatal(err)
					}
					if err := json.Unmarshal(buf, tc.newDest()); err != nil {
						b.Fatal(err)
					}
				}
			})
			b.Run("Decoder", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					r := streamedReader(data)
					if err := json.NewDecoder(r).Decode(tc.newDest()); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}
