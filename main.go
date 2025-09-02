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
	"context"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/prometheus-community/elasticsearch_exporter/collector"
	"github.com/prometheus-community/elasticsearch_exporter/config"
	"github.com/prometheus-community/elasticsearch_exporter/pkg/clusterinfo"
	"github.com/prometheus-community/elasticsearch_exporter/pkg/roundtripper"
)

const name = "elasticsearch_exporter"

type transportWithAPIKey struct {
	underlyingTransport http.RoundTripper
	apiKey              string
}

func (t *transportWithAPIKey) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("ApiKey %s", t.apiKey))
	return t.underlyingTransport.RoundTrip(req)
}

func main() {
	var (
		metricsPath = kingpin.Flag("web.telemetry-path",
			"Path under which to expose metrics.").
			Default("/metrics").String()
		toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9114")
		esURI        = kingpin.Flag("es.uri",
			"HTTP API address of an Elasticsearch node.").
			Default("").String()
		esTimeout = kingpin.Flag("es.timeout",
			"Timeout for trying to get stats from Elasticsearch.").
			Default("5s").Duration()
		esAllNodes = kingpin.Flag("es.all",
			"Export stats for all nodes in the cluster. If used, this flag will override the flag es.node.").
			Default("false").Bool()
		esNode = kingpin.Flag("es.node",
			"Node's name of which metrics should be exposed.").
			Default("_local").String()
		esExportIndices = kingpin.Flag("es.indices",
			"Export stats for indices in the cluster.").
			Default("false").Bool()
		esExportIndicesSettings = kingpin.Flag("es.indices_settings",
			"Export stats for settings of all indices of the cluster.").
			Default("false").Bool()
		esExportIndicesMappings = kingpin.Flag("es.indices_mappings",
			"Export stats for mappings of all indices of the cluster.").
			Default("false").Bool()
		esExportIndexAliases = kingpin.Flag("es.aliases",
			"Export informational alias metrics.").
			Default("true").Bool()
		esExportShards = kingpin.Flag("es.shards",
			"Export stats for shards in the cluster (implies --es.indices).").
			Default("false").Bool()
		esClusterInfoInterval = kingpin.Flag("es.clusterinfo.interval",
			"Cluster info update interval for the cluster label").
			Default("5m").Duration()
		esCA = kingpin.Flag("es.ca",
			"Path to PEM file that contains trusted Certificate Authorities for the Elasticsearch connection.").
			Default("").String()
		esClientPrivateKey = kingpin.Flag("es.client-private-key",
			"Path to PEM file that contains the private key for client auth when connecting to Elasticsearch.").
			Default("").String()
		esClientCert = kingpin.Flag("es.client-cert",
			"Path to PEM file that contains the corresponding cert for the private key to connect to Elasticsearch.").
			Default("").String()
		esInsecureSkipVerify = kingpin.Flag("es.ssl-skip-verify",
			"Skip SSL verification when connecting to Elasticsearch.").
			Default("false").Bool()
		logOutput = kingpin.Flag("log.output",
			"Sets the log output. Valid outputs are stdout and stderr").
			Default("stdout").String()
		awsRegion = kingpin.Flag("aws.region",
			"Region for AWS elasticsearch").
			Default("").String()
		awsRoleArn = kingpin.Flag("aws.role-arn",
			"Role ARN of an IAM role to assume.").
			Default("").String()
		configFile = kingpin.Flag("config.file", "Path to YAML configuration file.").Default("").String()
	)

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print(name))
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	// Load optional YAML config
	var cfg *config.Config
	if *configFile != "" {
		var cfgErr error
		cfg, cfgErr = config.LoadConfig(*configFile)
		if cfgErr != nil {
			// At this stage logger not yet created; fallback to stderr
			fmt.Fprintf(os.Stderr, "failed to load config file: %v\n", cfgErr)
			os.Exit(1)
		}
	}

	var w io.Writer
	switch strings.ToLower(*logOutput) {
	case "stderr":
		w = os.Stderr
	case "stdout":
		w = os.Stdout
	default:
		w = os.Stdout
	}
	promslogConfig.Writer = w
	logger := promslog.New(promslogConfig)

	// version metric
	prometheus.MustRegister(versioncollector.NewCollector(name))

	// Create a context that is cancelled on SIGKILL or SIGINT.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if *esURI != "" {
		esURL, err := url.Parse(*esURI)
		if err != nil {
			logger.Error("failed to parse es.uri", "err", err)
			os.Exit(1)
		}

		esUsername := os.Getenv("ES_USERNAME")
		esPassword := os.Getenv("ES_PASSWORD")

		if esUsername != "" && esPassword != "" {
			esURL.User = url.UserPassword(esUsername, esPassword)
		}

		// returns nil if not provided and falls back to simple TCP.
		tlsConfig := createTLSConfig(*esCA, *esClientCert, *esClientPrivateKey, *esInsecureSkipVerify)

		var httpTransport http.RoundTripper

		httpTransport = &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
		}

		esAPIKey := os.Getenv("ES_API_KEY")

		if esAPIKey != "" {
			httpTransport = &transportWithAPIKey{
				underlyingTransport: httpTransport,
				apiKey:              esAPIKey,
			}
		}

		httpClient := &http.Client{
			Timeout:   *esTimeout,
			Transport: httpTransport,
		}

		if *awsRegion != "" {
			var err error
			httpClient.Transport, err = roundtripper.NewAWSSigningTransport(httpTransport, *awsRegion, *awsRoleArn, logger)
			if err != nil {
				logger.Error("failed to create AWS transport", "err", err)
				os.Exit(1)
			}
		}

		// create the exporter
		exporter, err := collector.NewElasticsearchCollector(
			logger,
			[]string{},
			collector.WithElasticsearchURL(esURL),
			collector.WithHTTPClient(httpClient),
		)
		if err != nil {
			logger.Error("failed to create Elasticsearch collector", "err", err)
			os.Exit(1)
		}
		prometheus.MustRegister(exporter)

		// TODO(@sysadmind): Remove this when we have a better way to get the cluster name to down stream collectors.
		// cluster info retriever
		clusterInfoRetriever := clusterinfo.New(logger, httpClient, esURL, *esClusterInfoInterval)

		prometheus.MustRegister(collector.NewClusterHealth(logger, httpClient, esURL))
		prometheus.MustRegister(collector.NewNodes(logger, httpClient, esURL, *esAllNodes, *esNode))

		if *esExportIndices || *esExportShards {
			sC := collector.NewShards(logger, httpClient, esURL)
			prometheus.MustRegister(sC)
			iC := collector.NewIndices(logger, httpClient, esURL, *esExportShards, *esExportIndexAliases)
			prometheus.MustRegister(iC)
			if registerErr := clusterInfoRetriever.RegisterConsumer(iC); registerErr != nil {
				logger.Error("failed to register indices collector in cluster info")
				os.Exit(1)
			}
			if registerErr := clusterInfoRetriever.RegisterConsumer(sC); registerErr != nil {
				logger.Error("failed to register shards collector in cluster info")
				os.Exit(1)
			}
		}

		if *esExportIndicesSettings {
			prometheus.MustRegister(collector.NewIndicesSettings(logger, httpClient, esURL))
		}

		if *esExportIndicesMappings {
			prometheus.MustRegister(collector.NewIndicesMappings(logger, httpClient, esURL))
		}

		// start the cluster info retriever
		switch runErr := clusterInfoRetriever.Run(ctx); runErr {
		case nil:
			logger.Info("started cluster info retriever", "interval", (*esClusterInfoInterval).String())
		case clusterinfo.ErrInitialCallTimeout:
			logger.Info("initial cluster info call timed out")
		default:
			logger.Error("failed to run cluster info retriever", "err", runErr)
			os.Exit(1)
		}

		// register cluster info retriever as prometheus collector
		prometheus.MustRegister(clusterInfoRetriever)
	}

	http.HandleFunc(*metricsPath, func(w http.ResponseWriter, r *http.Request) {
		// /metrics endpoint is reserved for single-target mode only.
		// For per-scrape overrides use the dedicated /probe endpoint.
		promhttp.Handler().ServeHTTP(w, r)
	})

	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "Elasticsearch Exporter",
			Description: "Prometheus Exporter for Elasticsearch servers",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			logger.Error("error creating landing page", "err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	// health endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
	})

	// probe endpoint
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		origQuery := r.URL.Query()
		targetStr, am, valErr := validateProbeParams(cfg, origQuery)
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
		pemCA := *esCA
		pemCert := *esClientCert
		pemKey := *esClientPrivateKey
		insecure := *esInsecureSkipVerify

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
		var transport http.RoundTripper = &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		}

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
			Timeout:   *esTimeout,
			Transport: transport,
		}

		reg := prometheus.NewRegistry()

		// version metric
		reg.MustRegister(versioncollector.NewCollector(name))

		// Core exporter collector
		exp, err := collector.NewElasticsearchCollector(
			logger,
			[]string{},
			collector.WithElasticsearchURL(targetURL),
			collector.WithHTTPClient(probeClient),
		)
		if err != nil {
			http.Error(w, "failed to create exporter", http.StatusInternalServerError)
			return
		}
		reg.MustRegister(exp)
		// Basic additional collectors – reuse global CLI flags
		reg.MustRegister(collector.NewClusterHealth(logger, probeClient, targetURL))
		reg.MustRegister(collector.NewNodes(logger, probeClient, targetURL, *esAllNodes, *esNode))
		if *esExportIndices || *esExportShards {
			shardsC := collector.NewShards(logger, probeClient, targetURL)
			indicesC := collector.NewIndices(logger, probeClient, targetURL, *esExportShards, *esExportIndexAliases)
			reg.MustRegister(shardsC)
			reg.MustRegister(indicesC)
		}
		if *esExportIndicesSettings {
			reg.MustRegister(collector.NewIndicesSettings(logger, probeClient, targetURL))
		}
		if *esExportIndicesMappings {
			reg.MustRegister(collector.NewIndicesMappings(logger, probeClient, targetURL))
		}

		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})

	server := &http.Server{}
	go func() {
		if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
			logger.Error("http server quit", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	// create a context for graceful http server shutdown
	srvCtx, srvCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer srvCancel()
	_ = server.Shutdown(srvCtx)
}
