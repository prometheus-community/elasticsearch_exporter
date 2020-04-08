package collector

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

func TestIndices(t *testing.T) {
	for _, ver := range testElasticsearchVersions {
		t.Run(ver, func(t *testing.T) {
			filename := fmt.Sprintf("../fixtures/indexstats/%s.json", ver)

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				content, err := ioutil.ReadFile(filename)
				if err != nil {
					t.Errorf("Failed to read fixture: %v", err)
				}
				fmt.Fprintf(w, "%s\n", content)
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}
			i := NewIndices(context.Background(), log.NewNopLogger(), http.DefaultClient, u, false, time.Nanosecond)
			stats, err := i.updater.fetchAndDecodeIndexStats()
			if err != nil {
				t.Fatalf("Failed to fetch or decode indices stats: %s", err)
			}
			// t.Logf("[%s] Index Response: %+v", ver, stats)
			if stats.Indices["foo_1"].Primaries.Docs.Count != 2 {
				t.Errorf("Wrong number of primary docs")
			}
			if stats.Indices["foo_1"].Primaries.Store.SizeInBytes == 0 {
				t.Errorf("Wrong number of primary store size in bytes")
			}
			if stats.Indices["foo_1"].Total.Store.SizeInBytes == 0 {
				t.Errorf("Wrong number of total store size in bytes")
			}
			if stats.Indices["foo_1"].Total.Indexing.IndexTimeInMillis == 0 {
				t.Errorf("Wrong indexing time recorded")
			}
			if stats.Indices["foo_1"].Total.Indexing.IndexTotal == 0 {
				t.Errorf("Wrong indexing total recorded")
			}
		})
	}
}
