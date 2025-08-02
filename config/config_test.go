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

// Additional test coverage for apikey and aws based authentication modules.
func TestLoadConfigAPIKey(t *testing.T) {
	yaml := `auth_modules:
  api_only:
    type: apikey
    apikey: secretkey123
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
	if am.Type != "apikey" || am.APIKey != "secretkey123" {
		t.Fatalf("unexpected apikey module: %+v", am)
	}
}

func TestLoadConfigAWS(t *testing.T) {
	yaml := `auth_modules:
  awsmod:
    type: aws
    aws:
      region: us-east-1
      role_arn: arn:aws:iam::123456789012:role/metrics
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

	awsMod := cfg.AuthModules["awsmod"]
	if awsMod.Type != "aws" || awsMod.AWS == nil || awsMod.AWS.Region != "us-east-1" {
		t.Fatalf("unexpected aws module: %+v", awsMod)
	}
}

func TestLoadConfigInvalidUserPass(t *testing.T) {
	// missing userpass section for type=userpass
	yaml := `auth_modules:
  bad:
    type: userpass
`
	tmp, err := os.CreateTemp(t.TempDir(), "cfg-*.yml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := tmp.WriteString(yaml); err != nil {
		t.Fatalf("write: %v", err)
	}
	tmp.Close()

	if _, err := LoadConfig(tmp.Name()); err == nil {
		t.Fatalf("expected validation error for missing credentials, got nil")
	}
}
