package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	yaml := `auth_modules:
  foo:
    type: userpass
    userpass:
      username: bar
      password: baz
    options:
      sslmode: disable
`
	tmp, err := os.CreateTemp(t.TempDir(), "cfg-*.yml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := tmp.WriteString(yaml); err != nil {
		t.Fatalf("write: %v", err)
	}
	tmp.Close()

	cfg, err := LoadConfig(tmp.Name())
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.AuthModules["foo"].UserPass.Username != "bar" {
		t.Fatalf("unexpected username: %s", cfg.AuthModules["foo"].UserPass.Username)
	}
}

// Additional test coverage for apikey based authentication modules.
// Technically, a module could specify both userpass and apikey configs, but
// the `type` field should dictate which credentials are considered valid by
// the application logic.
func TestLoadConfigAPIKey(t *testing.T) {
	yaml := `auth_modules:
  api_only:
    type: apikey
    apikey: secretkey123
  mixed:
    type: apikey
    apikey: anotherkey456
    userpass:
      username: should
      password: be_ignored
`
	tmp, err := os.CreateTemp(t.TempDir(), "cfg-*.yml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := tmp.WriteString(yaml); err != nil {
		t.Fatalf("write: %v", err)
	}
	tmp.Close()

	cfg, err := LoadConfig(tmp.Name())
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	am := cfg.AuthModules["api_only"]
	if am.Type != "apikey" {
		t.Fatalf("expected module type apikey, got %s", am.Type)
	}
	if am.APIKey != "secretkey123" {
		t.Fatalf("unexpected apikey value: %s", am.APIKey)
	}

	mixed := cfg.AuthModules["mixed"]
	if mixed.Type != "apikey" {
		t.Fatalf("expected mixed module type apikey, got %s", mixed.Type)
	}
	if mixed.APIKey != "anotherkey456" {
		t.Fatalf("unexpected mixed apikey value: %s", mixed.APIKey)
	}
	// The userpass credentials should still be parsed but are expected to be ignored
	// by the application when the type is apikey.
	if mixed.UserPass == nil {
		t.Fatalf("expected userpass section to be parsed for mixed module")
	}
}
