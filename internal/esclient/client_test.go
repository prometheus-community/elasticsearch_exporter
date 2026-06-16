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

package esclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/common/promslog"

	"github.com/prometheus-community/elasticsearch_exporter/config"
)

func TestNewBuildsClientAndAppliesBasicAuth(t *testing.T) {
	cfg := config.NewConfigWithDefaults()
	cfg.ElasticsearchURL = "http://example.com:9200"
	cfg.Username = "user"
	cfg.Password = "pass"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate config: %v", err)
	}

	client, err := New(cfg, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if got := client.URL.User.Username(); got != "user" {
		t.Fatalf("unexpected username: %s", got)
	}
	password, ok := client.URL.User.Password()
	if !ok || password != "pass" {
		t.Fatalf("unexpected password: %s", password)
	}
}

func TestNewAppliesAPIKeyTransport(t *testing.T) {
	var got string
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
	}))
	defer server.Close()

	cfg := config.NewConfigWithDefaults()
	cfg.ElasticsearchURL = server.URL
	cfg.APIKey = "secret"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate config: %v", err)
	}
	client, err := New(cfg, promslog.NewNopLogger())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	resp, err := client.HTTP.Get(server.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if got != "ApiKey secret" {
		t.Fatalf("unexpected authorization header: %s", got)
	}
}

func TestTLSConfigReturnsErrors(t *testing.T) {
	_, err := TLSConfig(config.TLSConfig{CAFile: "/path/does/not/exist"})
	if err == nil {
		t.Fatalf("expected CA loading error")
	}
}
