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

package main

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus-community/elasticsearch_exporter/cluster"
	"github.com/prometheus-community/elasticsearch_exporter/collector"
	"github.com/prometheus-community/elasticsearch_exporter/config"
	"github.com/prometheus-community/elasticsearch_exporter/pkg/roundtripper"
)

var (
	errMissingTarget     = errors.New("missing target parameter")
	errInvalidTarget     = errors.New("invalid target parameter")
	errModuleNotFound    = errors.New("auth_module not found")
	errUnsupportedModule = errors.New("unsupported auth_module type")
)

// validateProbeParams performs upfront validation of the query parameters.
// It returns the target string (as given), the resolved AuthModule (optional), or an error.
func validateProbeParams(cfg *config.Config, q url.Values) (string, *config.AuthModule, error) {
	target := q.Get("target")
	if target == "" {
		return "", nil, errMissingTarget
	}

	// If the target does not contain an URL scheme, default to http.
	// This allows users to pass "host:port" without the "http://" prefix.
	if !strings.Contains(target, "://") {
		target = "http://" + target
	}

	u, err := url.Parse(target)
	if err != nil {
		return "", nil, errInvalidTarget
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", nil, errInvalidTarget
	}

	modu := q.Get("auth_module")
	if modu == "" {
		return target, nil, nil // no auth module requested
	}
	if cfg == nil {
		return "", nil, errModuleNotFound
	}
	am, ok := cfg.AuthModules[modu]
	if !ok {
		return "", nil, errModuleNotFound
	}
	switch strings.ToLower(am.Type) {
	case "userpass":
		return target, &am, nil
	case "apikey":
		return target, &am, nil
	case "aws":
		// Accept module even if region omitted; environment resolver can provide it.
		return target, &am, nil
	case "tls":
		// TLS auth type is valid; detailed TLS validation is performed during config load.
		return target, &am, nil
	default:
		return "", nil, errUnsupportedModule
	}
}

type probeHandlerOptions struct {
	cfg                     *config.Config
	logger                  *slog.Logger
	esCA                    string
	esClientCert            string
	esClientPrivateKey      string
	esInsecureSkipVerify    bool
	esTimeout               time.Duration
	esClusterInfoInterval   time.Duration
	esAllNodes              bool
	esNode                  string
	esExportIndices         bool
	esExportShards          bool
	esExportIndicesSettings bool
	esExportIndicesMappings bool
	esExportIndexAliases    bool
}

func newProbeHandler(opts probeHandlerOptions) http.HandlerFunc {
	logger := opts.logger
	if logger == nil {
		logger = slog.Default()
	}
	timeout := opts.esTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	clusterInfoInterval := opts.esClusterInfoInterval
	if clusterInfoInterval == 0 {
		clusterInfoInterval = 5 * time.Minute
	}
	node := opts.esNode
	if node == "" {
		node = "_local"
	}

	return func(w http.ResponseWriter, r *http.Request) {
		origQuery := r.URL.Query()
		targetStr, am, valErr := validateProbeParams(opts.cfg, origQuery)
		if valErr != nil {
			http.Error(w, valErr.Error(), http.StatusBadRequest)
			return
		}
		targetURL, _ := url.Parse(targetStr)
		if am != nil {
			// Apply userpass credentials only if the module type is explicitly set to userpass.
			if strings.EqualFold(am.Type, "userpass") && am.UserPass != nil {
				targetURL.User = url.UserPassword(am.UserPass.Username, am.UserPass.Password)
			}
			if len(am.Options) > 0 {
				q := targetURL.Query()
				for k, v := range am.Options {
					q.Set(k, v)
				}
				targetURL.RawQuery = q.Encode()
			}
		}

		// Build a dedicated HTTP client for this probe request (reuse TLS opts, timeout, etc.).
		pemCA := opts.esCA
		pemCert := opts.esClientCert
		pemKey := opts.esClientPrivateKey
		insecure := opts.esInsecureSkipVerify

		// Apply TLS configuration from auth module if provided (for transport security)
		// This matches single-target behavior where TLS settings are always applied
		if am != nil && am.TLS != nil {
			// Override with module-specific TLS settings
			if am.TLS.CAFile != "" {
				pemCA = am.TLS.CAFile
			}
			if am.TLS.CertFile != "" {
				pemCert = am.TLS.CertFile
			}
			if am.TLS.KeyFile != "" {
				pemKey = am.TLS.KeyFile
			}
			if am.TLS.InsecureSkipVerify {
				insecure = true
			}
		}
		tlsCfg := createTLSConfig(pemCA, pemCert, pemKey, insecure)
		baseTransport := &http.Transport{
			TLSClientConfig:   tlsCfg,
			Proxy:             http.ProxyFromEnvironment,
			ForceAttemptHTTP2: true,
			DisableKeepAlives: true,
		}
		var transport http.RoundTripper = baseTransport

		// inject authentication based on auth_module type
		if am != nil {
			switch strings.ToLower(am.Type) {
			case "apikey":
				if am.APIKey != "" {
					transport = &transportWithAPIKey{
						underlyingTransport: transport,
						apiKey:              am.APIKey,
					}
				}
			case "aws":
				var region string
				if am.AWS.Region != "" {
					region = am.AWS.Region
				}
				var err error
				transport, err = roundtripper.NewAWSSigningTransport(transport, region, am.AWS.RoleARN, logger)
				if err != nil {
					http.Error(w, "failed to create AWS signing transport", http.StatusInternalServerError)
					return
				}
			case "tls":
				// No additional auth wrapper needed - client certificates in TLS config handle authentication
			case "userpass":
				// Already handled above by setting targetURL.User
			}
		}
		probeClient := &http.Client{
			Timeout:   timeout,
			Transport: transport,
		}
		// Close idle connections when handler completes to prevent resource leaks.
		defer baseTransport.CloseIdleConnections()

		reg := prometheus.NewRegistry()

		// version metric
		reg.MustRegister(versioncollector.NewCollector(name))

		// Per-probe cluster info provider. NewInfoProvider is lazy (fetches on
		// first GetInfo, then caches), so constructing one per request starts no
		// background goroutine and is safe to discard when the handler returns.
		infoProvider := cluster.NewInfoProvider(logger, probeClient, targetURL, clusterInfoInterval)

		// Core exporter collector
		exp, err := collector.NewElasticsearchCollector(
			logger,
			[]string{},
			collector.WithElasticsearchURL(targetURL),
			collector.WithHTTPClient(probeClient),
			collector.WithClusterInfoProvider(infoProvider),
		)
		if err != nil {
			http.Error(w, "failed to create exporter", http.StatusInternalServerError)
			return
		}
		reg.MustRegister(exp)
		// Basic additional collectors – reuse global CLI flags
		reg.MustRegister(collector.NewClusterHealth(logger, probeClient, targetURL))
		reg.MustRegister(collector.NewNodes(logger, probeClient, targetURL, opts.esAllNodes, node))
		if opts.esExportIndices || opts.esExportShards {
			shardsC := collector.NewShards(logger, probeClient, targetURL)
			indicesC := collector.NewIndices(logger, probeClient, targetURL, opts.esExportShards, opts.esExportIndexAliases)
			reg.MustRegister(shardsC)
			reg.MustRegister(indicesC)
		}
		if opts.esExportIndicesSettings {
			reg.MustRegister(collector.NewIndicesSettings(logger, probeClient, targetURL))
		}
		if opts.esExportIndicesMappings {
			reg.MustRegister(collector.NewIndicesMappings(logger, probeClient, targetURL))
		}

		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}
}
