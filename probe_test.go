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

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/promslog"

	"github.com/prometheus-community/elasticsearch_exporter/config"
)

func TestValidateProbeParams(t *testing.T) {
	cfg := &config.Config{AuthModules: map[string]config.AuthModule{}}
	// missing target
	_, _, err := validateProbeParams(cfg, url.Values{})
	if err != errMissingTarget {
		t.Fatalf("expected missing target error, got %v", err)
	}

	// invalid target
	vals := url.Values{}
	vals.Set("target", "http://[::1")
	_, _, err = validateProbeParams(cfg, vals)
	if err == nil {
		t.Fatalf("expected invalid target error")
	}

	// invalid scheme
	vals = url.Values{}
	vals.Set("target", "ftp://example.com")
	_, _, err = validateProbeParams(cfg, vals)
	if err == nil {
		t.Fatalf("expected invalid target error for unsupported scheme")
	}

	// unknown module
	vals = url.Values{}
	vals.Set("target", "http://localhost:9200")
	vals.Set("auth_module", "foo")
	_, _, err = validateProbeParams(cfg, vals)
	if err != errModuleNotFound {
		t.Fatalf("expected module not found error, got %v", err)
	}

	// good path (userpass)
	cfg.AuthModules["foo"] = config.AuthModule{Type: "userpass", UserPass: &config.UserPassConfig{Username: "u", Password: "p"}}
	vals = url.Values{}
	vals.Set("target", "http://localhost:9200")
	vals.Set("auth_module", "foo")
	tgt, am, err := validateProbeParams(cfg, vals)
	if err != nil || am == nil || tgt == "" {
		t.Fatalf("expected success, got err=%v", err)
	}

	// good path (apikey) with both userpass and apikey set - apikey should be accepted
	cfg.AuthModules["api"] = config.AuthModule{
		Type:     "apikey",
		APIKey:   "mysecret",
		UserPass: &config.UserPassConfig{Username: "u", Password: "p"},
	}
	vals = url.Values{}
	vals.Set("target", "http://localhost:9200")
	vals.Set("auth_module", "api")
	_, am, err = validateProbeParams(cfg, vals)
	if err != nil {
		t.Fatalf("expected success for apikey module, got err=%v", err)
	}
	if am == nil || am.Type != "apikey" {
		t.Fatalf("expected apikey module, got %+v", am)
	}
	if am.APIKey != "mysecret" {
		t.Fatalf("unexpected apikey value: %s", am.APIKey)
	}

	// good path (aws)
	cfg.AuthModules["awsmod"] = config.AuthModule{
		Type: "aws",
		AWS: &config.AWSConfig{
			Region:  "us-east-1",
			RoleARN: "arn:aws:iam::123456789012:role/metrics",
		},
	}
	vals = url.Values{}
	vals.Set("target", "http://localhost:9200")
	vals.Set("auth_module", "awsmod")
	_, am, err = validateProbeParams(cfg, vals)
	if err != nil {
		t.Fatalf("expected success for aws module, got err=%v", err)
	}
	if am == nil || am.Type != "aws" {
		t.Fatalf("expected aws module, got %+v", am)
	}
	if am.AWS == nil || am.AWS.Region != "us-east-1" {
		t.Fatalf("unexpected aws config: %+v", am.AWS)
	}

	// invalid path (aws with empty region - rejected at config load; simulate here by passing nil cfg lookup)
	// No additional test needed as config.LoadConfig enforces region.

	// good path (tls)
	cfg.AuthModules["mtls"] = config.AuthModule{
		Type: "tls",
		TLS:  &config.TLSConfig{CAFile: "/dev/null", CertFile: "/dev/null", KeyFile: "/dev/null"},
	}
	vals = url.Values{}
	vals.Set("target", "http://localhost:9200")
	vals.Set("auth_module", "mtls")
	_, am, err = validateProbeParams(cfg, vals)
	if err != nil {
		t.Fatalf("expected success for tls module, got err=%v", err)
	}
	if am == nil || am.Type != "tls" {
		t.Fatalf("expected tls module, got %+v", am)
	}
}

func mockClusterRootJSON(clusterName, clusterUUID, version, luceneVersion string) string {
	return fmt.Sprintf(`{
  "name": "node-1",
  "cluster_name": "%s",
  "cluster_uuid": "%s",
  "version": {
    "number": "%s",
    "build_hash": "abc123",
    "build_date": "2021-05-28T17:40:59.346932922Z",
    "build_snapshot": false,
    "lucene_version": "%s"
  },
  "tagline": "You Know, for Search"
}`, clusterName, clusterUUID, version, luceneVersion)
}

func mockClusterHealthJSON(clusterName string) string {
	return fmt.Sprintf(`{
  "cluster_name": "%s",
  "status": "green",
  "timed_out": false,
  "number_of_nodes": 1,
  "number_of_data_nodes": 1,
  "active_primary_shards": 5,
  "active_shards": 5
}`, clusterName)
}

func mockNodesJSON(clusterName string) string {
	return fmt.Sprintf(`{
  "cluster_name": "%s",
  "nodes": {}
}`, clusterName)
}

func TestProbeHandlerMultiTarget(t *testing.T) {
	_, _ = kingpin.CommandLine.Parse([]string{})

	serverA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/":
			_, _ = w.Write([]byte(mockClusterRootJSON("cluster-A", "uuid-cluster-A", "7.13.1", "8.8.2")))
		case "/_cluster/health":
			_, _ = w.Write([]byte(mockClusterHealthJSON("cluster-A")))
		case "/_nodes/_local/stats", "/_nodes/stats":
			_, _ = w.Write([]byte(mockNodesJSON("cluster-A")))
		default:
			http.NotFound(w, r)
		}
	}))
	defer serverA.Close()

	serverB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/":
			_, _ = w.Write([]byte(mockClusterRootJSON("cluster-B", "uuid-cluster-B", "8.10.0", "9.7.0")))
		case "/_cluster/health":
			_, _ = w.Write([]byte(mockClusterHealthJSON("cluster-B")))
		case "/_nodes/_local/stats", "/_nodes/stats":
			_, _ = w.Write([]byte(mockNodesJSON("cluster-B")))
		default:
			http.NotFound(w, r)
		}
	}))
	defer serverB.Close()

	logger := promslog.NewNopLogger()
	handler := newProbeHandler(probeHandlerOptions{
		logger:                logger,
		esTimeout:             5 * time.Second,
		esClusterInfoInterval: 5 * time.Minute,
		esNode:                "_local",
	})

	// 1. Probe target A
	reqA := httptest.NewRequest(http.MethodGet, "/probe?target="+url.QueryEscape(serverA.URL), nil)
	recA := httptest.NewRecorder()
	handler.ServeHTTP(recA, reqA)

	if recA.Code != http.StatusOK {
		t.Fatalf("target A probe returned status %d: %s", recA.Code, recA.Body.String())
	}
	bodyA := recA.Body.String()
	if !strings.Contains(bodyA, `cluster="cluster-A"`) {
		t.Errorf("target A response should contain cluster=\"cluster-A\", got:\n%s", bodyA)
	}
	if !strings.Contains(bodyA, `cluster_uuid="uuid-cluster-A"`) {
		t.Errorf("target A response should contain cluster_uuid=\"uuid-cluster-A\", got:\n%s", bodyA)
	}
	if strings.Contains(bodyA, "cluster-B") || strings.Contains(bodyA, "uuid-cluster-B") {
		t.Errorf("target A response should not contain target B metadata, got:\n%s", bodyA)
	}

	// 2. Probe target B (must return target B metadata, not reuse target A)
	reqB := httptest.NewRequest(http.MethodGet, "/probe?target="+url.QueryEscape(serverB.URL), nil)
	recB := httptest.NewRecorder()
	handler.ServeHTTP(recB, reqB)

	if recB.Code != http.StatusOK {
		t.Fatalf("target B probe returned status %d: %s", recB.Code, recB.Body.String())
	}
	bodyB := recB.Body.String()
	if !strings.Contains(bodyB, `cluster="cluster-B"`) {
		t.Errorf("target B response should contain cluster=\"cluster-B\", got:\n%s", bodyB)
	}
	if !strings.Contains(bodyB, `cluster_uuid="uuid-cluster-B"`) {
		t.Errorf("target B response should contain cluster_uuid=\"uuid-cluster-B\", got:\n%s", bodyB)
	}
	if strings.Contains(bodyB, "cluster-A") || strings.Contains(bodyB, "uuid-cluster-A") {
		t.Errorf("target B response should not contain target A metadata, got:\n%s", bodyB)
	}

	// 3. Probe target A again (must still return target A metadata, no bleeding)
	reqA2 := httptest.NewRequest(http.MethodGet, "/probe?target="+url.QueryEscape(serverA.URL), nil)
	recA2 := httptest.NewRecorder()
	handler.ServeHTTP(recA2, reqA2)

	if recA2.Code != http.StatusOK {
		t.Fatalf("target A second probe returned status %d: %s", recA2.Code, recA2.Body.String())
	}
	bodyA2 := recA2.Body.String()
	if !strings.Contains(bodyA2, `cluster="cluster-A"`) {
		t.Errorf("target A second probe response should contain cluster=\"cluster-A\", got:\n%s", bodyA2)
	}
	if !strings.Contains(bodyA2, `cluster_uuid="uuid-cluster-A"`) {
		t.Errorf("target A second probe response should contain cluster_uuid=\"uuid-cluster-A\", got:\n%s", bodyA2)
	}
	if strings.Contains(bodyA2, "cluster-B") || strings.Contains(bodyA2, "uuid-cluster-B") {
		t.Errorf("target A second probe response should not contain target B metadata, got:\n%s", bodyA2)
	}
}
