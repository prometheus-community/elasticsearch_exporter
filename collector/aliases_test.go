package collector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestAliases(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine
	//  curl -XPUT http://localhost:9200/foo_1
	//  curl -XPUT http://localhost:9200/foo_2
	//  curl -XPOST http://localhost:9200/_aliases -d '{"actions":[{"add":{"index":"foo_1","alias":"foo"}},{"add":{"index":"foo_1","alias":"fooA"}},{"add":{"index":"foo_2","alias":"fooB"}}]}'
	//  curl http://localhost:9200/_aliases
	ta := map[string]string{
		"1.7.6": `{"foo_1":{"aliases":{"foo":{},"fooA":{}}},"foo_2":{"aliases":{"fooB":{}}}}`,
		"2.4.5": `{"foo_2":{"aliases":{"fooB":{}}},"foo_1":{"aliases":{"foo":{},"fooA":{}}}}`,
		"5.4.2": `{".monitoring-alerts-2":{"aliases":{}},".monitoring-data-2":{"aliases":{}},".watches":{"aliases":{}},".monitoring-es-2-2017.08.29":{"aliases":{}},".triggered_watches":{"aliases":{}},"foo_1":{"aliases":{"foo":{},"fooA":{}}},".watcher-history-3-2017.08.29":{"aliases":{}},"foo_2":{"aliases":{"fooB":{}}}}`,
		"5.5.2": `{".watches":{"aliases":{}},".monitoring-alerts-6":{"aliases":{}},"foo_2":{"aliases":{"fooB":{}}},".monitoring-es-6-2017.08.29":{"aliases":{}},".watcher-history-3-2017.08.29":{"aliases":{}},".triggered_watches":{"aliases":{}},"foo_1":{"aliases":{"foo":{},"fooA":{}}}}`,
	}
	for ver, out := range ta {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, out)
		}))
		defer ts.Close()

		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatalf("Failed to parse URL: %s", err)
		}
		a := NewAliases(log.NewNopLogger(), http.DefaultClient, u)
		stats, err := a.fetchAndDecodeAliasStats()
		if err != nil {
			t.Fatalf("Failed to fetch or decode indices stats: %s", err)
		}
		t.Logf("[%s] Index Response: %+v", ver, stats)
		if len(stats["foo_1"]["aliases"]) != 2 {
			t.Errorf("Wrong number of aliases")
		}
		if len(stats["foo_2"]["aliases"]) != 1 {
			t.Errorf("Wrong number of aliases")
		}
	}
}
