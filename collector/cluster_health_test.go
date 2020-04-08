package collector

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestClusterHealth(t *testing.T) {
	for _, ver := range testElasticsearchVersions {
		t.Run(ver, func(t *testing.T) {
			filename := fmt.Sprintf("../fixtures/clusterhealth/%s.json", ver)
			f, _ := os.Open(filename)
			defer f.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(w, f)
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
			if ver != "1.7.6" {
				if chr.TaskMaxWaitingInQueueMillis != 12 {
					t.Errorf("Wrong task max waiting time in millis")
				}
			}
		})
	}
}
