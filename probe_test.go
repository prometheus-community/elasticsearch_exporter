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
	"net/url"
	"testing"

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
