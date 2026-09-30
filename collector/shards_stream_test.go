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
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"testing"
)

// This file is the empirical companion to the streaming change in shards.go,
// mirroring the approach used for /_all/_stats in PR #1159
// (collector/indices_stream_test.go and collector/decode_realworld_test.go):
//
//   - TestStreamShardsEquivalence : correctness — streaming decode yields the
//                                   exact same shards (and the exact same
//                                   per-node aggregation the collector emits)
//                                   as the previous whole-array json decode.
//   - BenchmarkShardsDecode        : buffered-vs-streaming on synthetic payloads
//                                   scaled to real-cluster shard counts, with
//                                   SetBytes + ReportAllocs (bytes/op is the
//                                   metric that drives GC pressure).
//   - BenchmarkShardsDecodeHTTP    : the same contrast over a real
//                                   *http.Response.Body served by httptest, i.e.
//                                   the actual production I/O path.
//   - BenchmarkShardsRetainedHeap  : a direct measurement of live heap retained
//                                   while the collector works — the buffered
//                                   path holds the whole []ShardResponse for the
//                                   duration of the emit loop, the streaming
//                                   path holds only the small per-node map.
//
// The win being demonstrated: streaming keeps peak/retained heap proportional
// to a single shard entry instead of the full shard list, which on large
// clusters (and especially under concurrent /probe scrapes) is the difference.

// buildCatShardsPayload synthesizes a /_cat/shards?format=json array with
// nShards entries, modeled on the real fixture (fixtures/shards/7.15.0.json):
// every object carries the full eight fields the API returns (index, shard,
// prirep, state, docs, store, ip, node) even though ShardResponse only decodes
// four, and roughly one shard in twenty is UNASSIGNED with null docs/store/ip/
// node — so the buffered baseline pays the realistic parse/allocation cost.
func buildCatShardsPayload(tb testing.TB, nShards int) []byte {
	tb.Helper()
	// nodes scales with shard count so the per-node aggregation map — the only
	// thing the streaming path retains — reflects a realistic large-cluster node
	// count instead of a fixed best case.
	nodes := min(max(nShards/100, 3), 500)
	var b bytes.Buffer
	b.Grow(nShards * 160)
	b.WriteByte('[')
	for i := 0; i < nShards; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		index := fmt.Sprintf("logs-2026.06.%02d-%06d", i%30, i/8)
		prirep := "p"
		if i%3 == 0 {
			prirep = "r"
		}
		if i%20 == 7 { // unassigned replica, null fields like a real response
			fmt.Fprintf(&b, `{"index":%q,"shard":"%d","prirep":%q,"state":"UNASSIGNED","docs":null,"store":null,"ip":null,"node":null}`,
				index, i%8, prirep)
			continue
		}
		fmt.Fprintf(&b, `{"index":%q,"shard":"%d","prirep":%q,"state":"STARTED","docs":"%d","store":"%dmb","ip":"10.0.%d.%d","node":"node-%04d"}`,
			index, i%8, prirep, (i*7)%100000, (i%900)+1, i%256, (i*3)%256, i%nodes)
	}
	b.WriteByte(']')
	data := b.Bytes()

	// Verify the synthetic payload decodes into the production struct.
	var sink []ShardResponse
	if err := json.Unmarshal(data, &sink); err != nil {
		tb.Fatalf("synthetic /_cat/shards payload (n=%d) does not decode: %v", nShards, err)
	}
	return data
}

// startedPerNode is the exact aggregation the Shards collector performs.
func startedPerNode(shards []ShardResponse) map[string]float64 {
	agg := make(map[string]float64)
	for _, s := range shards {
		if s.State == "STARTED" {
			agg[s.Node]++
		}
	}
	return agg
}

func TestStreamShardsEquivalence(t *testing.T) {
	realFixture, err := os.ReadFile("../fixtures/shards/7.15.0.json")
	if err != nil {
		t.Fatal(err)
	}
	inputs := map[string][]byte{
		"fixture_7.15.0": realFixture,
		"synthetic_2000": buildCatShardsPayload(t, 2000),
	}

	for name, data := range inputs {
		t.Run(name, func(t *testing.T) {
			var want []ShardResponse
			if err := json.Unmarshal(data, &want); err != nil {
				t.Fatal(err)
			}

			var got []ShardResponse
			if err := streamShards(bytes.NewReader(data), func(s ShardResponse) {
				got = append(got, s)
			}); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("streamed shards differ from json.Unmarshal result (got %d entries, want %d)", len(got), len(want))
			}
			if !reflect.DeepEqual(startedPerNode(got), startedPerNode(want)) {
				t.Fatalf("per-node STARTED aggregation differs between streaming and buffered decode")
			}
		})
	}
}

func BenchmarkShardsDecode(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		data := buildCatShardsPayload(b, n)
		label := fmt.Sprintf("shards_%d_(%dKB)", n, len(data)/1024)

		// Previous behavior: decode the whole array into []ShardResponse, then aggregate.
		b.Run(label+"/buffered_decode_slice", func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			var sink float64
			for i := 0; i < b.N; i++ {
				var sfr []ShardResponse
				if err := json.NewDecoder(bytes.NewReader(data)).Decode(&sfr); err != nil {
					b.Fatal(err)
				}
				for _, v := range startedPerNode(sfr) {
					sink += v
				}
			}
			_ = sink
		})

		// New behavior: decode one shard at a time, aggregating as we go.
		b.Run(label+"/streaming", func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			var sink float64
			for i := 0; i < b.N; i++ {
				nodeShards := make(map[string]float64)
				if err := streamShards(bytes.NewReader(data), func(s ShardResponse) {
					if s.State == "STARTED" {
						nodeShards[s.Node]++
					}
				}); err != nil {
					b.Fatal(err)
				}
				for _, v := range nodeShards {
					sink += v
				}
			}
			_ = sink
		})
	}
}

func BenchmarkShardsDecodeHTTP(b *testing.B) {
	data := buildCatShardsPayload(b, 10000)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer ts.Close()
	client := ts.Client()

	b.Run("buffered_decode_slice", func(b *testing.B) {
		b.ReportAllocs()
		var sink float64
		for i := 0; i < b.N; i++ {
			resp, err := client.Get(ts.URL)
			if err != nil {
				b.Fatal(err)
			}
			var sfr []ShardResponse
			err = json.NewDecoder(resp.Body).Decode(&sfr)
			resp.Body.Close()
			if err != nil {
				b.Fatal(err)
			}
			for _, v := range startedPerNode(sfr) {
				sink += v
			}
		}
		_ = sink
	})

	b.Run("streaming", func(b *testing.B) {
		b.ReportAllocs()
		var sink float64
		for i := 0; i < b.N; i++ {
			resp, err := client.Get(ts.URL)
			if err != nil {
				b.Fatal(err)
			}
			nodeShards := make(map[string]float64)
			err = streamShards(resp.Body, func(s ShardResponse) {
				if s.State == "STARTED" {
					nodeShards[s.Node]++
				}
			})
			resp.Body.Close()
			if err != nil {
				b.Fatal(err)
			}
			for _, v := range nodeShards {
				sink += v
			}
		}
		_ = sink
	})
}

// retainedHeapBytes returns the increase in live heap attributable to produce()
// once its result is kept alive across a GC. It approximates the heap the
// collector holds while it works: the input payload is allocated by the caller
// (so it is live in both samples and cancels out), and only what produce()
// returns is still reachable at the second sample. Two GCs settle the baseline,
// a final GC reclaims the transient garbage, and runtime.KeepAlive prevents the
// result from being collected before it is measured.
//
// This measures post-GC RETAINED heap, not transient peak: it is a conservative
// lower bound on the buffered path's true peak, which additionally holds the
// read buffer and decode garbage. The direction is structural (buffered retains
// the full slice, streaming only the per-node map); the magnitude is reported.
// Sub-baseline GC jitter can make the delta negative when the retained set is
// tiny, so it is clamped to zero — the streaming arm is therefore a small
// near-noise-floor figure, not an exact byte count.
func retainedHeapBytes(produce func() any) uint64 {
	runtime.GC()
	runtime.GC()
	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)
	result := produce()
	runtime.GC()
	runtime.ReadMemStats(&m1)
	runtime.KeepAlive(result)
	if m1.HeapAlloc <= m0.HeapAlloc {
		return 0
	}
	return m1.HeapAlloc - m0.HeapAlloc
}

func BenchmarkShardsRetainedHeap(b *testing.B) {
	const n = 50000
	data := buildCatShardsPayload(b, n)

	// Buffered: the whole []ShardResponse stays live alongside the aggregation
	// map for the duration of the (real) emit loop.
	b.Run(fmt.Sprintf("shards_%d/buffered_decode_slice", n), func(b *testing.B) {
		var sum uint64
		for i := 0; i < b.N; i++ {
			sum += retainedHeapBytes(func() any {
				var sfr []ShardResponse
				if err := json.NewDecoder(bytes.NewReader(data)).Decode(&sfr); err != nil {
					b.Fatal(err)
				}
				return []any{sfr, startedPerNode(sfr)}
			})
		}
		b.ReportMetric(float64(sum)/float64(b.N)/(1<<20), "retained_MB")
	})

	// Streaming: only the small per-node map survives; every shard struct is
	// freed as soon as it is counted.
	b.Run(fmt.Sprintf("shards_%d/streaming", n), func(b *testing.B) {
		var sum uint64
		for i := 0; i < b.N; i++ {
			sum += retainedHeapBytes(func() any {
				nodeShards := make(map[string]float64)
				if err := streamShards(bytes.NewReader(data), func(s ShardResponse) {
					if s.State == "STARTED" {
						nodeShards[s.Node]++
					}
				}); err != nil {
					b.Fatal(err)
				}
				return nodeShards
			})
		}
		b.ReportMetric(float64(sum)/float64(b.N)/(1<<20), "retained_MB")
	})
}
