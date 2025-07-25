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
