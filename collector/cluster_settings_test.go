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

func TestClusterSettingsStats(t *testing.T) {
	for _, ver := range testElasticsearchVersions {
		for _, fpart := range []string{"", "-updated"} {
			name := fmt.Sprintf("%s%s", ver, fpart)
			t.Run(name, func(t *testing.T) {
				filename := fmt.Sprintf("../fixtures/clustersettings/%s.json", name)
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
				c := NewClusterSettings(log.NewNopLogger(), http.DefaultClient, u)
				nsr, err := c.fetchAndDecodeClusterSettingsStats()
				if err != nil {
					t.Fatalf("Failed to fetch or decode cluster settings stats: %s", err)
				}
				t.Logf("Cluster Settings Stats Response: %+v", nsr)

				// Version 6.X+ will use the lowercase result when responding
				if strings.ToLower(nsr.Cluster.Routing.Allocation.Enabled) != "all" {
					t.Errorf("Wrong setting for cluster routing allocation enabled, expected ALL, got %v", nsr.Cluster.Routing.Allocation.Enabled)
				}
				if strings.Split(ver, ".")[0] == "5" {
					if nsr.Cluster.MaxShardsPerNode != "" {
						t.Errorf("MaxShardsPerNode should be empty on older releases")
					}
				} else {
					if nsr.Cluster.MaxShardsPerNode != "1000" {
						t.Errorf("Wrong value for MaxShardsPerNode")
					}
				}
			})

		}
	}
}
