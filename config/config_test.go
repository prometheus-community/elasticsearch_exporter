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
	return f.Name()
}

func TestLoadConfigTLSValid(t *testing.T) {
	ca := mustTempFile(t)
	cert := mustTempFile(t)
	key := mustTempFile(t)
	yaml := `auth_modules:
  secure:
    type: userpass
    userpass:
      username: foo
      password: bar
    tls:
      ca_file: ` + ca + `
      cert_file: ` + cert + `
      key_file: ` + key + `
`
	tmp, _ := os.CreateTemp(t.TempDir(), "cfg-*.yml")
	tmp.WriteString(yaml)
	tmp.Close()
	if _, err := LoadConfig(tmp.Name()); err != nil {
		t.Fatalf("expected config to load, got %v", err)
	}
}

func TestLoadConfigTLSMissingKey(t *testing.T) {
	cert := mustTempFile(t)
	yaml := `auth_modules:
  badtls:
    type: userpass
    userpass:
      username: foo
      password: bar
    tls:
      cert_file: ` + cert + `
`
	tmp, _ := os.CreateTemp(t.TempDir(), "cfg-*.yml")
	tmp.WriteString(yaml)
	tmp.Close()
	if _, err := LoadConfig(tmp.Name()); err == nil {
		t.Fatalf("expected validation error for missing key_file")
	}
}

func TestLoadConfigValidationErrors(t *testing.T) {
	badPath := "/path/does/not/exist"
	key := mustTempFile(t)
	cases := []struct {
		name string
		yaml string
	}{
		{
			"tlsMissingCert",
			`auth_modules:
  bad:
    type: userpass
    userpass: {username: u, password: p}
    tls: {key_file: ` + key + `}`,
		},
		{
			"tlsBadCAPath",
			`auth_modules:
  bad:
    type: userpass
    userpass: {username: u, password: p}
    tls: {ca_file: ` + badPath + `}`,
		},
		{
			"awsNoRegion",
			`auth_modules:
  bad:
    type: aws
    aws: {}`,
		},
		{
			"unsupportedType",
			`auth_modules:
  bad:
    type: foobar`,
		},
	}
	for _, c := range cases {
		tmp, _ := os.CreateTemp(t.TempDir(), "cfg-*.yml")
		tmp.WriteString(c.yaml)
		tmp.Close()
		if _, err := LoadConfig(tmp.Name()); err == nil {
			t.Fatalf("%s: expected validation error, got nil", c.name)
		}
	}
}
