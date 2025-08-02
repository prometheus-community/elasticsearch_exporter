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
	Options  map[string]string `yaml:"options,omitempty"`
}

// AWSConfig contains settings for SigV4 authentication.
type AWSConfig struct {
	Region  string `yaml:"region"`
	RoleARN string `yaml:"role_arn,omitempty"`
}

type UserPassConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// validate ensures every auth module has the required fields according to its type.
func (c *Config) validate() error {
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
			if am.AWS == nil || am.AWS.Region == "" {
				return fmt.Errorf("auth_module %s type aws requires region", name)
			}
		default:
			return fmt.Errorf("auth_module %s has unsupported type %s", name, am.Type)
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
