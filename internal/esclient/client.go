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

package esclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/prometheus-community/elasticsearch_exporter/config"
	"github.com/prometheus-community/elasticsearch_exporter/pkg/roundtripper"
)

type transportWithAPIKey struct {
	underlyingTransport http.RoundTripper
	apiKey              string
}

func (t *transportWithAPIKey) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("ApiKey %s", t.apiKey))
	return t.underlyingTransport.RoundTrip(req)
}

type Client struct {
	URL       *url.URL
	HTTP      *http.Client
	Transport *http.Transport
}

func New(cfg config.Config, logger *slog.Logger) (*Client, error) {
	if cfg.ElasticsearchURL == "" {
		return nil, fmt.Errorf("elasticsearch URL must not be empty")
	}
	esURL, err := url.Parse(cfg.ElasticsearchURL)
	if err != nil {
		return nil, fmt.Errorf("parse elasticsearch URL: %w", err)
	}
	if cfg.Username != "" && cfg.Password != "" {
		esURL.User = url.UserPassword(cfg.Username, cfg.Password)
	}

	tlsConfig, err := TLSConfig(cfg.TLS)
	if err != nil {
		return nil, err
	}
	baseTransport := &http.Transport{
		TLSClientConfig:   tlsConfig,
		Proxy:             http.ProxyFromEnvironment,
		ForceAttemptHTTP2: true,
	}
	var transport http.RoundTripper = baseTransport
	if cfg.APIKey != "" {
		transport = &transportWithAPIKey{
			underlyingTransport: transport,
			apiKey:              cfg.APIKey,
		}
	}
	if cfg.AWSEnabled {
		transport, err = roundtripper.NewAWSSigningTransport(transport, cfg.AWS.Region, cfg.AWS.RoleARN, logger)
		if err != nil {
			return nil, fmt.Errorf("create AWS signing transport: %w", err)
		}
	}

	return &Client{
		URL: esURL,
		HTTP: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		Transport: baseTransport,
	}, nil
}

func TLSConfig(cfg config.TLSConfig) (*tls.Config, error) {
	tlsConfig := tls.Config{}
	if cfg.InsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}
	if cfg.CAFile != "" {
		rootCerts, err := loadCertificatesFrom(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("load root certificate from %s: %w", cfg.CAFile, err)
		}
		tlsConfig.RootCAs = rootCerts
	}
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		if _, err := loadPrivateKeyFrom(cfg.CertFile, cfg.KeyFile); err != nil {
			return nil, fmt.Errorf("setup client authentication: %w", err)
		}
		tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return loadPrivateKeyFrom(cfg.CertFile, cfg.KeyFile)
		}
	}
	return &tlsConfig, nil
}

func loadCertificatesFrom(pemFile string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}
	certificates := x509.NewCertPool()
	certificates.AppendCertsFromPEM(caCert)
	return certificates, nil
}

func loadPrivateKeyFrom(pemCertFile, pemPrivateKeyFile string) (*tls.Certificate, error) {
	privateKey, err := tls.LoadX509KeyPair(pemCertFile, pemPrivateKeyFile)
	if err != nil {
		return nil, err
	}
	return &privateKey, nil
}
