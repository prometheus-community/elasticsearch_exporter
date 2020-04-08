package collector

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestIndicesSettings(t *testing.T) {
	for _, ver := range testElasticsearchVersions {
		t.Run(ver, func(t *testing.T) {
			filename := fmt.Sprintf("../fixtures/indexsettings/%s.json", ver)
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
			c := NewIndicesSettings(log.NewNopLogger(), http.DefaultClient, u)
			nsr, err := c.fetchAndDecodeIndicesSettings()
			if err != nil {
				t.Fatalf("Failed to fetch or decode indices settings: %s", err)
			}
			t.Logf("All Indices Settings Response: %+v", nsr)
			// if nsr.Cluster.Routing.Allocation.Enabled != "ALL" {
			// 	t.Errorf("Wrong setting for cluster routing allocation enabled")
			// }
			// 5.X does not support the Settings.indexInfo.Blocks.ReadOnly attribute
			if strings.Split(ver, ".")[0] == "5" {
				return
			}
			var counter int
			for key, value := range nsr {
				if value.Settings.IndexInfo.Blocks.ReadOnly == "true" {
					counter++
					if key != "instagram" && key != "twitter" {
						t.Errorf("Wrong read_only index")
					}
				}
			}
			if counter != 2 {
				t.Errorf("Wrong number of read_only indexes")
			}
		})

	}
}
