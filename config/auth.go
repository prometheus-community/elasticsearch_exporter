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

// AuthConfig represents the YAML configuration file structure for /probe auth modules.
type AuthConfig struct {
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

// validate ensures every auth module has the required fields according to its type.
func (c *AuthConfig) validate() error {
	for name, am := range c.AuthModules {
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
			if am.TLS == nil {
				return fmt.Errorf("auth_module %s type tls requires tls configuration section", name)
			}
			if am.TLS.CertFile == "" || am.TLS.KeyFile == "" {
				return fmt.Errorf("auth_module %s type tls requires cert_file and key_file for client certificate authentication", name)
			}
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

		if am.TLS != nil {
			if strings.ToLower(am.Type) != "tls" && (am.TLS.CertFile != "") != (am.TLS.KeyFile != "") {
				return fmt.Errorf("auth_module %s: if providing client certificate, both cert_file and key_file must be specified", name)
			}
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

// LoadAuthConfig reads, parses, and validates the YAML auth module config file.
func LoadAuthConfig(path string) (*AuthConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AuthConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if cfg.AuthModules == nil {
		cfg.AuthModules = map[string]AuthModule{}
	}
	return &cfg, nil
}
