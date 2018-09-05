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
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine  -Des.path.repo="/tmp" (1.7.6, 2.4.5)
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine  -E path.repo="/tmp" (5.4.2)
	//  curl -XPUT http://localhost:9200/foo_1/type1/1 -d '{"title":"abc","content":"hello"}'
	//  curl -XPUT http://localhost:9200/foo_1/type1/2 -d '{"title":"def","content":"world"}'
	//  curl -XPUT http://localhost:9200/foo_2/type1/1 -d '{"title":"abc001","content":"hello001"}'
	//  curl -XPUT http://localhost:9200/foo_2/type1/2 -d '{"title":"def002","content":"world002"}'
	//  curl -XPUT http://localhost:9200/foo_2/type1/3 -d '{"title":"def003","content":"world003"}'
	//  curl -XPUT http://localhost:9200/_snapshot/test1 -d '{"type": "fs","settings":{"location": "/tmp/test1"}}'
	//  curl -XPUT "http://localhost:9200/_snapshot/test1/snapshot_1?wait_for_completion=true"
	//  curl http://localhost:9200/_snapshot/
	//  curl http://localhost:9200/_snapshot/test1/_all

	tcs := map[string][]string{
		"1.7.6": []string{`{"test1":{"type":"fs","settings":{"location":"/tmp/test1"}}}`, `{"snapshots":[{"snapshot":"snapshot_1","version_id":1070699,"version":"1.7.6","indices":["foo_1","foo_2"],"state":"SUCCESS","start_time":"2018-09-04T09:09:02.427Z","start_time_in_millis":1536052142427,"end_time":"2018-09-04T09:09:02.755Z","end_time_in_millis":1536052142755,"duration_in_millis":328,"failures":[],"shards":{"total":10,"failed":0,"successful":10}}]}`},
		"2.4.5": []string{`{"test1":{"type":"fs","settings":{"location":"/tmp/test1"}}}`, `{"snapshots":[{"snapshot":"snapshot_1","version_id":2040599,"version":"2.4.5","indices":["foo_2","foo_1"],"state":"SUCCESS","start_time":"2018-09-04T09:25:25.818Z","start_time_in_millis":1536053125818,"end_time":"2018-09-04T09:25:26.326Z","end_time_in_millis":1536053126326,"duration_in_millis":508,"failures":[],"shards":{"total":10,"failed":0,"successful":10}}]}`},
		"5.4.2": []string{`{"test1":{"type":"fs","settings":{"location":"/tmp/test1"}}}`, `{"snapshots":[{"snapshot":"snapshot_1","uuid":"VZ_c_kKISAW8rpcqiwSg0w","version_id":5040299,"version":"5.4.2","indices":["foo_2","foo_1"],"state":"SUCCESS","start_time":"2018-09-04T09:29:13.971Z","start_time_in_millis":1536053353971,"end_time":"2018-09-04T09:29:14.477Z","end_time_in_millis":1536053354477,"duration_in_millis":506,"failures":[],"shards":{"total":10,"failed":0,"successful":10}}]}`},
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
		repositoryStats := stats["test1"]
		snapshotStats := repositoryStats.Snapshots[0]

		if len(snapshotStats.Indices) != 2 {
			t.Errorf("Bad number of snapshot indices")
		}
		if len(snapshotStats.Failures) != 0 {
			t.Errorf("Bad number of snapshot failures")
		}
		if snapshotStats.Shards.Total != 10 {
			t.Errorf("Bad number of snapshot shards total")
		}
		if snapshotStats.Shards.Failed != 0 {
			t.Errorf("Bad number of snapshot shards failed")
		}
		if snapshotStats.Shards.Successful != 10 {
			t.Errorf("Bad number of snapshot shards successful")
		}
		if len(repositoryStats.Snapshots) != 1 {
			t.Errorf("Bad number of repository snapshots")
		}
	}

}
