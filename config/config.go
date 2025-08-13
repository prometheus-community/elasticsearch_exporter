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
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Config represents the YAML configuration file structure.
type Config struct {
	AuthModules map[string]AuthModule `yaml:"auth_modules"`
}

type AuthModule struct {
	Type     string            `yaml:"type"`
	UserPass *UserPassConfig   `yaml:"userpass,omitempty"`
	APIKey   string            `yaml:"apikey,omitempty"`
	AWS      *AWSConfig        `yaml:"aws,omitempty"`
	TLS      *TLSConfig        `yaml:"tls,omitempty"`
	Options  map[string]string `yaml:"options,omitempty"`
}

// AWSConfig contains settings for SigV4 authentication.
type AWSConfig struct {
	Region  string `yaml:"region,omitempty"`
	RoleARN string `yaml:"role_arn,omitempty"`
}

// TLSConfig allows per-target TLS options.
type TLSConfig struct {
	CAFile             string `yaml:"ca_file,omitempty"`
	CertFile           string `yaml:"cert_file,omitempty"`
	KeyFile            string `yaml:"key_file,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify,omitempty"`
}

type UserPassConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// validate ensures every auth module has the required fields according to its type.
func (c *Config) validate() error {
	for name, am := range c.AuthModules {
		// Validate fields based on auth type
		switch strings.ToLower(am.Type) {
		case "userpass":
			if am.UserPass == nil || am.UserPass.Username == "" || am.UserPass.Password == "" {
				return fmt.Errorf("auth_module %s type userpass requires username and password", name)
			}
		case "apikey":
			if am.APIKey == "" {
				return fmt.Errorf("auth_module %s type apikey requires apikey", name)
			}
		case "aws":
			// No strict validation: region can come from environment/defaults; role_arn is optional.
		case "tls":
			// TLS auth type means client certificate authentication only (no other auth)
			if am.TLS == nil {
				return fmt.Errorf("auth_module %s type tls requires tls configuration section", name)
			}
			if am.TLS.CertFile == "" || am.TLS.KeyFile == "" {
				return fmt.Errorf("auth_module %s type tls requires cert_file and key_file for client certificate authentication", name)
			}
			// Validate that other auth fields are not set when using TLS auth type
			if am.UserPass != nil {
				return fmt.Errorf("auth_module %s type tls cannot have userpass configuration", name)
			}
			if am.APIKey != "" {
				return fmt.Errorf("auth_module %s type tls cannot have apikey", name)
			}
			if am.AWS != nil {
				return fmt.Errorf("auth_module %s type tls cannot have aws configuration", name)
			}
		default:
			return fmt.Errorf("auth_module %s has unsupported type %s", name, am.Type)
		}

		// Validate TLS configuration (optional for all auth types, provides transport security)
		if am.TLS != nil {
			// For cert-based auth (type: tls), cert and key are required
			// For other auth types, TLS config is optional and used for transport security
			if strings.ToLower(am.Type) != "tls" {
				// For non-TLS auth types, if cert/key are provided, both must be present
				if (am.TLS.CertFile != "") != (am.TLS.KeyFile != "") {
					return fmt.Errorf("auth_module %s: if providing client certificate, both cert_file and key_file must be specified", name)
				}
			}

			// Validate file accessibility
			for fileType, path := range map[string]string{
				"ca_file":   am.TLS.CAFile,
				"cert_file": am.TLS.CertFile,
				"key_file":  am.TLS.KeyFile,
			} {
				if path == "" {
					continue
				}
				if _, err := os.Stat(path); err != nil {
					return fmt.Errorf("auth_module %s: %s '%s' not accessible: %w", name, fileType, path, err)
				}
			}
		}
	}
	return nil
}

// LoadConfig reads, parses, and validates the YAML config file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
