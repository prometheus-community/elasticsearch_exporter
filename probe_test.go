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

	// good path
	cfg.AuthModules["foo"] = config.AuthModule{Type: "userpass", UserPass: &config.UserPassConfig{Username: "u", Password: "p"}}
	vals = url.Values{}
	vals.Set("target", "http://localhost:9200")
	vals.Set("auth_module", "foo")
	tgt, am, err := validateProbeParams(cfg, vals)
	if err != nil || am == nil || tgt == "" {
		t.Fatalf("expected success, got err=%v", err)
	}
}
