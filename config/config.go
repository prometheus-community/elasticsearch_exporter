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
	"net/url"
	"time"
)

type Config struct {
	ElasticsearchURL      string
	Timeout               time.Duration
	AllNodes              bool
	Node                  string
	ExportIndices         bool
	ExportIndicesMappings bool
	ExportIndexAliases    bool
	ExportShards          bool
	ClusterInfoInterval   time.Duration
	TLS                   TLSConfig
	AWS                   AWSConfig
	AWSEnabled            bool
	Username              string
	Password              string
	APIKey                string
	Collectors            map[string]bool
	TasksActions          string

	validated bool
}

const (
	DefaultElasticsearchURL      = ""
	DefaultTimeout               = 5 * time.Second
	DefaultAllNodes              = false
	DefaultNode                  = "_local"
	DefaultExportIndices         = false
	DefaultExportIndicesMappings = false
	DefaultExportIndexAliases    = true
	DefaultExportShards          = false
	DefaultClusterInfoInterval   = 5 * time.Minute
	DefaultTasksActions          = "indices:*"
)

const (
	CollectorClusterInfo     = "cluster-info"
	CollectorClusterSettings = "clustersettings"
	CollectorDataStream      = "data-stream"
	CollectorHealthReport    = "health-report"
	CollectorILM             = "ilm"
	CollectorIndicesSettings = "indices_settings"
	CollectorSnapshots       = "snapshots"
	CollectorSLM             = "slm"
	CollectorTasks           = "tasks"
)

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

func NewConfigWithDefaults() Config {
	return Config{
		ElasticsearchURL:      DefaultElasticsearchURL,
		Timeout:               DefaultTimeout,
		AllNodes:              DefaultAllNodes,
		Node:                  DefaultNode,
		ExportIndices:         DefaultExportIndices,
		ExportIndicesMappings: DefaultExportIndicesMappings,
		ExportIndexAliases:    DefaultExportIndexAliases,
		ExportShards:          DefaultExportShards,
		ClusterInfoInterval:   DefaultClusterInfoInterval,
		Collectors:            DefaultCollectorConfig(),
		TasksActions:          DefaultTasksActions,
	}
}

func DefaultCollectorConfig() map[string]bool {
	return map[string]bool{
		CollectorClusterInfo:     true,
		CollectorClusterSettings: false,
		CollectorDataStream:      false,
		CollectorHealthReport:    false,
		CollectorILM:             false,
		CollectorIndicesSettings: false,
		CollectorSnapshots:       false,
		CollectorSLM:             false,
		CollectorTasks:           false,
	}
}

func (c *Config) Validate() error {
	c.validated = false
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}
	if c.ClusterInfoInterval < 0 {
		return fmt.Errorf("cluster info interval must not be negative")
	}
	if c.Node == "" {
		return fmt.Errorf("node must not be empty")
	}
	if c.TasksActions == "" {
		return fmt.Errorf("tasks actions must not be empty")
	}
	if c.ElasticsearchURL != "" {
		u, err := url.Parse(c.ElasticsearchURL)
		if err != nil {
			return fmt.Errorf("elasticsearch URL is invalid: %w", err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("elasticsearch URL scheme must be http or https")
		}
		if u.Host == "" {
			return fmt.Errorf("elasticsearch URL host must not be empty")
		}
	}
	if (c.TLS.CertFile != "") != (c.TLS.KeyFile != "") {
		return fmt.Errorf("if providing client certificate, both cert_file and key_file must be specified")
	}
	defaults := DefaultCollectorConfig()
	for name := range c.Collectors {
		if name == "" {
			return fmt.Errorf("collector name must not be empty")
		}
		if _, ok := defaults[name]; !ok {
			return fmt.Errorf("unknown collector %q", name)
		}
	}
	c.validated = true
	return nil
}

func (c Config) Validated() bool {
	return c.validated
}
