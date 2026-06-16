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

package config

import (
	"os"
	"testing"
	"time"
)

func mustTempFile(t *testing.T) string {
	f, err := os.CreateTemp(t.TempDir(), "pem-*.crt")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	f.Close()
	// Ensure temp file is removed even if created outside of test's TempDir semantics change
	path := f.Name()
	t.Cleanup(func() { _ = os.Remove(path) })
	return path
}

// ---------------------------- Positive cases ----------------------------
func TestLoadAuthConfigPositiveVariants(t *testing.T) {
	ca := mustTempFile(t)
	cert := mustTempFile(t)
	key := mustTempFile(t)

	positive := []struct {
		name string
		yaml string
	}{{
		"userpass",
		`auth_modules:
  basic:
    type: userpass
    userpass:
      username: u
      password: p`,
	}, {
		"userpass-with-tls",
		`auth_modules:
  basic:
    type: userpass
    userpass:
      username: u
      password: p
    tls:
      ca_file: ` + ca + `
      insecure_skip_verify: true`,
	}, {
		"apikey",
		`auth_modules:
  key:
    type: apikey
    apikey: ZXhhbXBsZQ==`,
	}, {
		"apikey-with-tls",
		`auth_modules:
  key:
    type: apikey
    apikey: ZXhhbXBsZQ==
    tls:
      ca_file: ` + ca + `
      cert_file: ` + cert + `
      key_file: ` + key + ``,
	}, {
		"aws-with-tls",
		`auth_modules:
  awsmod:
    type: aws
    aws:
      region: us-east-1
    tls:
      insecure_skip_verify: true`,
	}, {
		"tls-only",
		`auth_modules:
  pki:
    type: tls
    tls:
      ca_file: ` + ca + `
      cert_file: ` + cert + `
      key_file: ` + key + ``,
	}}

	for _, c := range positive {
		tmp, _ := os.CreateTemp(t.TempDir(), "cfg-*.yml")
		_, _ = tmp.WriteString(c.yaml)
		_ = tmp.Close()
		t.Cleanup(func() { _ = os.Remove(tmp.Name()) })
		if _, err := LoadAuthConfig(tmp.Name()); err != nil {
			t.Fatalf("%s: expected success, got %v", c.name, err)
		}
	}
}

// ---------------------------- Negative cases ----------------------------
func TestLoadAuthConfigNegativeVariants(t *testing.T) {
	cert := mustTempFile(t)
	key := mustTempFile(t)

	negative := []struct {
		name string
		yaml string
	}{{
		"userpassMissingPassword",
		`auth_modules:
  bad:
    type: userpass
    userpass: {username: u}`,
	}, {
		"tlsMissingCert",
		`auth_modules:
  bad:
    type: tls
    tls: {key_file: ` + key + `}`,
	}, {
		"tlsMissingKey",
		`auth_modules:
  bad:
    type: tls
    tls: {cert_file: ` + cert + `}`,
	}, {
		"tlsMissingConfig",
		`auth_modules:
  bad:
    type: tls`,
	}, {
		"tlsWithUserpass",
		`auth_modules:
  bad:
    type: tls
    tls: {cert_file: ` + cert + `, key_file: ` + key + `}
    userpass: {username: u, password: p}`,
	}, {
		"tlsWithAPIKey",
		`auth_modules:
  bad:
    type: tls
    tls: {cert_file: ` + cert + `, key_file: ` + key + `}
    apikey: ZXhhbXBsZQ==`,
	}, {
		"tlsWithAWS",
		`auth_modules:
  bad:
    type: tls
    tls: {cert_file: ` + cert + `, key_file: ` + key + `}
    aws: {region: us-east-1}`,
	}, {
		"tlsIncompleteCert",
		`auth_modules:
  bad:
    type: apikey
    apikey: ZXhhbXBsZQ==
    tls: {cert_file: ` + cert + `}`,
	}, {
		"unsupportedType",
		`auth_modules:
  bad:
    type: foobar`,
	}}

	for _, c := range negative {
		tmp, _ := os.CreateTemp(t.TempDir(), "cfg-*.yml")
		_, _ = tmp.WriteString(c.yaml)
		_ = tmp.Close()
		t.Cleanup(func() { _ = os.Remove(tmp.Name()) })
		if _, err := LoadAuthConfig(tmp.Name()); err == nil {
			t.Fatalf("%s: expected validation error, got none", c.name)
		}
	}
}

func TestNewConfigWithDefaults(t *testing.T) {
	cfg := NewConfigWithDefaults()
	if cfg.ElasticsearchURL != DefaultElasticsearchURL {
		t.Fatalf("unexpected elasticsearch URL: %s", cfg.ElasticsearchURL)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Fatalf("unexpected timeout: %s", cfg.Timeout)
	}
	if cfg.AllNodes != DefaultAllNodes {
		t.Fatalf("unexpected all nodes setting: %v", cfg.AllNodes)
	}
	if cfg.Node != DefaultNode {
		t.Fatalf("unexpected node: %s", cfg.Node)
	}
	if cfg.ExportIndices != DefaultExportIndices {
		t.Fatalf("unexpected export indices setting: %v", cfg.ExportIndices)
	}
	if cfg.ExportIndicesMappings != DefaultExportIndicesMappings {
		t.Fatalf("unexpected export indices mappings setting: %v", cfg.ExportIndicesMappings)
	}
	if cfg.ExportIndexAliases != DefaultExportIndexAliases {
		t.Fatalf("unexpected export index aliases setting: %v", cfg.ExportIndexAliases)
	}
	if cfg.ExportShards != DefaultExportShards {
		t.Fatalf("unexpected export shards setting: %v", cfg.ExportShards)
	}
	if cfg.ClusterInfoInterval != DefaultClusterInfoInterval {
		t.Fatalf("unexpected cluster info interval: %s", cfg.ClusterInfoInterval)
	}
	if cfg.TasksActions != DefaultTasksActions {
		t.Fatalf("unexpected tasks actions: %s", cfg.TasksActions)
	}
	if got := cfg.Collectors[CollectorClusterInfo]; !got {
		t.Fatalf("cluster-info collector should be enabled by default")
	}
	if got := cfg.Collectors[CollectorTasks]; got {
		t.Fatalf("tasks collector should be disabled by default")
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := NewConfigWithDefaults()
	cfg.ElasticsearchURL = "http://localhost:9200"
	if cfg.Validated() {
		t.Fatalf("new config should not start validated")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
	if !cfg.Validated() {
		t.Fatalf("expected config to be marked validated")
	}
}

func TestConfigValidateRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		update func(*Config)
	}{{
		name: "invalid url",
		update: func(c *Config) {
			c.ElasticsearchURL = "ftp://localhost:9200"
		},
	}, {
		name: "invalid timeout",
		update: func(c *Config) {
			c.Timeout = 0
		},
	}, {
		name: "unknown collector",
		update: func(c *Config) {
			c.Collectors["unknown"] = true
		},
	}, {
		name: "incomplete tls client cert",
		update: func(c *Config) {
			c.TLS.CertFile = "client.pem"
		},
	}, {
		name: "empty tasks actions",
		update: func(c *Config) {
			c.TasksActions = ""
		},
	}, {
		name: "negative cluster info interval",
		update: func(c *Config) {
			c.ClusterInfoInterval = -time.Second
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfigWithDefaults()
			cfg.ElasticsearchURL = "http://localhost:9200"
			tt.update(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatalf("expected validation error")
			}
			if cfg.Validated() {
				t.Fatalf("invalid config should not be marked validated")
			}
		})
	}
}
