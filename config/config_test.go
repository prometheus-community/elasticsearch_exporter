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
func TestLoadConfigPositiveVariants(t *testing.T) {
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
		if _, err := LoadConfig(tmp.Name()); err != nil {
			t.Fatalf("%s: expected success, got %v", c.name, err)
		}
	}
}

// ---------------------------- Negative cases ----------------------------
func TestLoadConfigNegativeVariants(t *testing.T) {
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
		if _, err := LoadConfig(tmp.Name()); err == nil {
			t.Fatalf("%s: expected validation error, got none", c.name)
		}
	}
}
