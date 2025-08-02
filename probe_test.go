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
}
