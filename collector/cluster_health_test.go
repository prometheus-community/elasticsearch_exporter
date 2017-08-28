package collector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestClusterHealth(t *testing.T) {
	// Testcases created using:
	//  docker run -d -p 9200:9200 elasticsearch:VERSION-alpine
	//  curl -XPUT http://localhost:9200/twitter
	//  curl http://localhost:9200/_cluster/health
	tcs := map[string]string{
		"1.7.6": `{"cluster_name":"elasticsearch","status":"yellow","timed_out":false,"number_of_nodes":1,"number_of_data_nodes":1,"active_primary_shards":5,"active_shards":5,"relocating_shards":0,"initializing_shards":0,"unassigned_shards":5,"delayed_unassigned_shards":0,"number_of_pending_tasks":0,"number_of_in_flight_fetch":0}`,
		"2.4.5": `{"cluster_name":"elasticsearch","status":"yellow","timed_out":false,"number_of_nodes":1,"number_of_data_nodes":1,"active_primary_shards":5,"active_shards":5,"relocating_shards":0,"initializing_shards":0,"unassigned_shards":5,"delayed_unassigned_shards":0,"number_of_pending_tasks":0,"number_of_in_flight_fetch":0,"task_max_waiting_in_queue_millis":0,"active_shards_percent_as_number":50.0}`,
		"5.4.2": `{"cluster_name":"elasticsearch","status":"yellow","timed_out":false,"number_of_nodes":1,"number_of_data_nodes":1,"active_primary_shards":5,"active_shards":5,"relocating_shards":0,"initializing_shards":0,"unassigned_shards":5,"delayed_unassigned_shards":0,"number_of_pending_tasks":0,"number_of_in_flight_fetch":0,"task_max_waiting_in_queue_millis":0,"active_shards_percent_as_number":50.0}`,
	}
	for ver, out := range tcs {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, out)
		}))
		defer ts.Close()

		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatalf("Failed to parse URL: %s", err)
		}
		c := NewClusterHealth(log.NewNopLogger(), http.DefaultClient, u)
		chr, err := c.fetchAndDecodeClusterHealth()
		if err != nil {
			t.Fatalf("Failed to fetch or decode cluster health: %s", err)
		}
		t.Logf("[%s] Cluster Health Response: %+v", ver, chr)
		if chr.ClusterName != "elasticsearch" {
			t.Errorf("Invalid cluster health response")
		}
		if chr.Status != "yellow" {
			t.Errorf("Invalid cluster status")
		}
		if chr.TimedOut {
			t.Errorf("Check didn't time out")
		}
		if chr.NumberOfNodes != 1 {
			t.Errorf("Wrong number of nodes")
		}
		if chr.NumberOfDataNodes != 1 {
			t.Errorf("Wrong number of data nodes")
		}
	}
}
