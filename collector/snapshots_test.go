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
		//"1.7.6":           {`{"test1":{"type":"fs","settings":{"location":"/tmp/test1"}}}`, `{"snapshots":[{"snapshot":"snapshot_1","version_id":1070699,"version":"1.7.6","indices":["foo_1","foo_2"],"state":"SUCCESS","start_time":"2018-09-04T09:09:02.427Z","start_time_in_millis":1536052142427,"end_time":"2018-09-04T09:09:02.755Z","end_time_in_millis":1536052142755,"duration_in_millis":328,"failures":[],"shards":{"total":10,"failed":0,"successful":10}}]}`},
		//"2.4.5":           {`{"test1":{"type":"fs","settings":{"location":"/tmp/test1"}}}`, `{"snapshots":[{"snapshot":"snapshot_1","version_id":2040599,"version":"2.4.5","indices":["foo_2","foo_1"],"state":"SUCCESS","start_time":"2018-09-04T09:25:25.818Z","start_time_in_millis":1536053125818,"end_time":"2018-09-04T09:25:26.326Z","end_time_in_millis":1536053126326,"duration_in_millis":508,"failures":[],"shards":{"total":10,"failed":0,"successful":10}}]}`},
		//"5.4.2":           {`{"test1":{"type":"fs","settings":{"location":"/tmp/test1"}}}`, `{"snapshots":[{"snapshot":"snapshot_1","uuid":"VZ_c_kKISAW8rpcqiwSg0w","version_id":5040299,"version":"5.4.2","indices":["foo_2","foo_1"],"state":"SUCCESS","start_time":"2018-09-04T09:29:13.971Z","start_time_in_millis":1536053353971,"end_time":"2018-09-04T09:29:14.477Z","end_time_in_millis":1536053354477,"duration_in_millis":506,"failures":[],"shards":{"total":10,"failed":0,"successful":10}}]}`},
		//"5.4.2-failed":    {`{"test1":{"type":"fs","settings":{"location":"/tmp/test1"}}}`, `{"snapshots":[{"snapshot":"snapshot_1","uuid":"VZ_c_kKISAW8rpcqiwSg0w","version_id":5040299,"version":"5.4.2","indices":["foo_2","foo_1"],"state":"SUCCESS","start_time":"2018-09-04T09:29:13.971Z","start_time_in_millis":1536053353971,"end_time":"2018-09-04T09:29:14.477Z","end_time_in_millis":1536053354477,"duration_in_millis":506,"failures":[{"index" : "index_name","index_uuid" : "index_name","shard_id" : 52,"reason" : "IndexShardSnapshotFailedException[error deleting index file [pending-index-5] during cleanup]; nested: NoSuchFileException[Blob [pending-index-5] does not exist]; ","node_id" : "pPm9jafyTjyMk0T5A101xA","status" : "INTERNAL_SERVER_ERROR"}],"shards":{"total":10,"failed":1,"successful":10}}]}`},
		"7.5.2":		   {`{"succeed":{"type":"fs","settings":{"location":"/tmp/succeed"}}}`, `{"snapshots":[{"snapshot":"visible","uuid":"VsEA3Y9GStOVKPy0-nAA2A","version_id":7050299,"version":"7.5.2","indices":["foo_1","foo_2"],"include_global_state":true,"state":"SUCCESS","start_time":"2020-04-22T16:18:22.700Z","start_time_in_millis":1587572302700,"end_time":"2020-04-22T16:18:23.100Z","end_time_in_millis":1587572303100,"duration_in_millis":400,"failures":[],"shards":{"total":2,"failed":0,"successful":2}}]}`},
		"7.5.2-repo-fail": {`{"fail":{"type":"fs","settings":{"location":"/tmp/fail"}}}`, `{"error":{"root_cause":[{"type":"repository_exception","reason":"[fail] Could not determine repository generation from root blobs"}],"type":"repository_exception","reason":"[fail] Could not determine repository generation from root blobs","caused_by":{"type":"access_denied_exception","reason":"/tmp/repos/fail"}},"status":500}`},
	}
	for ver, out := range tcs {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.RequestURI == "/_snapshot" {
				fmt.Fprint(w, out[0])
				return
			}
			fmt.Fprint(w, out[1])
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
		succeedRepoStats := stats["succeed"]
		//failRepoStats := stats["fail"]
		succeedSnapshotStats := succeedRepoStats.Snapshots[0]

		if len(succeedSnapshotStats.Indices) != 2 {
			t.Errorf("Bad number of snapshot indices")
		}
		if len(succeedSnapshotStats.Failures) != int(succeedSnapshotStats.Shards.Failed) {
			t.Errorf("Bad number of snapshot failures")
		}
		if succeedSnapshotStats.Shards.Total != 10 {
			t.Errorf("Bad number of snapshot shards total")
		}
		if succeedSnapshotStats.Shards.Successful != 10 {
			t.Errorf("Bad number of snapshot shards successful")
		}
		if len(succeedRepoStats.Snapshots) != 1 {
			t.Errorf("Bad number of repository snapshots")
		}
	}

}
