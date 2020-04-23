package collector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestSnapshots(t *testing.T) {
	// Testcases created using $REPO/hack/snapshots_test_setup.sh

	tcs := map[string][]string{
		"7.5.2": {
			`{"succeed":{"type":"fs","settings":{"location":"/tmp/succeed"}},"fail":{"type":"fs","settings":{"location":"/tmp/fail"}}}`,
			`{"snapshots":[{"snapshot":"visible","uuid":"3IKCfpc1THyLr2AuMy1eng","version_id":7050299,"version":"7.5.2","indices":["foo_1","foo_2"],"include_global_state":true,"state":"SUCCESS","start_time":"2020-04-23T10:20:25.125Z","start_time_in_millis":1587637225125,"end_time":"2020-04-23T10:20:25.525Z","end_time_in_millis":1587637225525,"duration_in_millis":400,"failures":[],"shards":{"total":2,"failed":0,"successful":2}}]}`,
			`{"error":{"root_cause":[{"type":"repository_exception","reason":"[fail] Could not determine repository generation from root blobs"}],"type":"repository_exception","reason":"[fail] Could not determine repository generation from root blobs","caused_by":{"type":"access_denied_exception","reason":"/tmp/fail"}},"status":500}`,
		},
	}
	for ver, out := range tcs {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			case "/_snapshot":
				fmt.Fprint(w, out[0])
				return
			case "/_snapshot/succeed/_all":
				fmt.Fprint(w, out[1])
				return
			case "/_snapshot/fail/_all":
				fmt.Fprint(w, out[2])
				return
			}
		}))
		defer ts.Close()

		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatalf("Failed to parse URL: %s", err)
		}
		s := NewSnapshots(log.NewNopLogger(), http.DefaultClient, u)
		stats, err := s.fetchAndDecodeSnapshotsStats()
		if err != nil {
			t.Fatalf("Failed to fetch or decode snapshots stats: %s", err)
		}
		t.Logf("[%s] Snapshots Response: %+v", ver, stats)

		for repo, snapshotResponse := range stats {
			if repo == "fail" {
				if snapshotResponse.Snapshots != nil{
					t.Errorf("Returning non-nil Snapshots response for inaccessible repo")
				}
				continue
			} else {
				if len(snapshotResponse.Snapshots[0].Indices) != 2 {
					t.Errorf("Bad number of snapshot indices")
				}
				if len(snapshotResponse.Snapshots[0].Failures) != int(snapshotResponse.Snapshots[0].Shards.Failed) {
					t.Errorf("Bad number of snapshot failures")
				}
				if snapshotResponse.Snapshots[0].Shards.Total != 2 {
					t.Errorf("Bad number of snapshot shards total")
				}
				if snapshotResponse.Snapshots[0].Shards.Successful != 2 {
					t.Errorf("Bad number of snapshot shards successful")
				}
				if len(snapshotResponse.Snapshots) != 1 {
					t.Errorf("Bad number of repository snapshots")
				}
			}
		}
	}

}
