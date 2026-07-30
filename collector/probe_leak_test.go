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
	"net/http"
	"net/url"
	"runtime"
	"testing"
	"time"

	"github.com/prometheus/common/promslog"
)

// settleGoroutines polls the goroutine count a few times so background
// receive-loop goroutines that are exiting have a chance to unwind before we
// take a measurement.
func settleGoroutines() int {
	var n int
	for i := 0; i < 50; i++ {
		runtime.Gosched()
		time.Sleep(2 * time.Millisecond)
		n = runtime.NumGoroutine()
	}
	return n
}

// TestProbeCollectorsCloseNoGoroutineLeak guards against the /probe goroutine
// leak: NewShards and NewIndices each start a cluster info receive loop that
// only exits when their channel is closed. In /probe mode a fresh collector is
// built per scrape and nothing feeds or closes that channel, so without Close()
// every scrape would leak two goroutines. Creating and closing many collectors
// must return the goroutine count to its baseline.
func TestProbeCollectorsCloseNoGoroutineLeak(t *testing.T) {
	u, err := url.Parse("http://localhost:9200")
	if err != nil {
		t.Fatal(err)
	}

	before := settleGoroutines()

	const n = 50
	for i := 0; i < n; i++ {
		s := NewShards(promslog.NewNopLogger(), http.DefaultClient, u)
		idx := NewIndices(promslog.NewNopLogger(), http.DefaultClient, u, false, false)
		s.Close()
		idx.Close()
		// Close must be idempotent and safe to call more than once.
		s.Close()
		idx.Close()
	}

	after := settleGoroutines()

	// Each iteration leaks 2 goroutines without the fix (2*n == 100 here), so a
	// small tolerance distinguishes cleanly from the leaking behaviour.
	if after > before+10 {
		t.Fatalf("goroutine leak: created and closed %d Shards+Indices collectors, goroutines before=%d after=%d", n, before, after)
	}
}
