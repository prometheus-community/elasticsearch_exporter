package config

import (
	"os"

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
	Options  map[string]string `yaml:"options,omitempty"`
}

type UserPassConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// LoadConfig reads and parses YAML config file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
