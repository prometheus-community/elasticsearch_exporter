// Copyright 2021 The Prometheus Authors
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
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"context"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/elasticsearch_exporter/collector"
	"github.com/prometheus-community/elasticsearch_exporter/pkg/clusterinfo"
	"github.com/prometheus-community/elasticsearch_exporter/pkg/roundtripper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
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
			Default("http://localhost:9200").String()
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
		esExportILM = kingpin.Flag("es.ilm",
			"Export index lifecycle politics for indices in the cluster.").
			Default("false").Bool()
		esExportShards = kingpin.Flag("es.shards",
			"Export stats for shards in the cluster (implies --es.indices).").
			Default("false").Bool()
		esExportSLM = kingpin.Flag("es.slm",
			"Export stats for SLM snapshots.").
			Default("false").Bool()
		esExportDataStream = kingpin.Flag("es.data_stream",
			"Export stas for Data Streams.").
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
		logLevel = kingpin.Flag("log.level",
			"Sets the loglevel. Valid levels are debug, info, warn, error").
			Default("info").String()
		logFormat = kingpin.Flag("log.format",
			"Sets the log format. Valid formats are json and logfmt").
			Default("logfmt").String()
		logOutput = kingpin.Flag("log.output",
			"Sets the log output. Valid outputs are stdout and stderr").
			Default("stdout").String()
		awsRegion = kingpin.Flag("aws.region",
			"Region for AWS elasticsearch").
			Default("").String()
		awsRoleArn = kingpin.Flag("aws.role-arn",
			"Role ARN of an IAM role to assume.").
			Default("").String()
	)

	kingpin.Version(version.Print(name))
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	logger := getLogger(*logLevel, *logOutput, *logFormat)

	esURL, err := url.Parse(*esURI)
	if err != nil {
		level.Error(logger).Log(
			"msg", "failed to parse es.uri",
			"err", err,
		)
		os.Exit(1)
	}

	esUsername := os.Getenv("ES_USERNAME")
	esPassword := os.Getenv("ES_PASSWORD")

	if esUsername != "" && esPassword != "" {
		esURL.User = url.UserPassword(esUsername, esPassword)
	}

	clusterRetrieverMap := make(map[string]*clusterinfo.Retriever)

	// Create a context that is cancelled on SIGKILL or SIGINT.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	probePath := "/probe"

	http.Handle(*metricsPath, promhttp.Handler())
	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "Elasticsearch Exporter",
			Description: "Prometheus Exporter for Elasticsearch servers",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address:     *metricsPath,
					Text:        "Metrics",
					Description: "Metrics endpoint exposing elasticsearch-exporter metrics in the Prometheus exposition format.",
				},
				{
					Address:     probePath,
					Text:        "Probe",
					Description: "Probe endpoint for testing the exporter against a specific Elasticsearch instance.",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		target := params.Get("target")

		if target != "" {
			targetURL, err := url.Parse(target)
			if err != nil {
				level.Error(logger).Log("invalid target", err)
				http.Error(w, "invalid target", http.StatusBadRequest)
				return
			}

			targetUsername := os.Getenv("ES_USERNAME")
			targetPassword := os.Getenv("ES_PASSWORD")

			authModule := params.Get("auth_module")
			if authModule != "" {
				authModule = strings.ToUpper(authModule)
				targetUsername = os.Getenv(fmt.Sprintf("ES_%s_USERNAME", authModule))
				targetPassword = os.Getenv(fmt.Sprintf("ES_%s_PASSWORD", authModule))
			}

			if targetUsername != "" && targetPassword != "" {
				targetURL.User = url.UserPassword(targetUsername, targetPassword)
			}

			esURL = targetURL
		}

		registry := prometheus.NewRegistry()
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
			httpClient.Transport, err = roundtripper.NewAWSSigningTransport(httpTransport, *awsRegion, *awsRoleArn, logger)
			if err != nil {
				level.Error(logger).Log("msg", "failed to create AWS transport", "err", err)
				http.Error(w, "failed to create AWS transport", http.StatusInternalServerError)
				return
			}
		}

		// version metric
		registry.MustRegister(version.NewCollector(name))

		// create the exporter
		exporter, err := collector.NewElasticsearchCollector(
			logger,
			[]string{},
			collector.WithElasticsearchURL(esURL),
			collector.WithHTTPClient(httpClient),
		)
		if err != nil {
			level.Error(logger).Log("msg", "failed to create Elasticsearch collector", "err", err)
			http.Error(w, "failed to create Elasticsearch collector", http.StatusInternalServerError)
			return
		}
		registry.MustRegister(exporter)

		// TODO(@sysadmind): Remove this when we have a better way to get the cluster name to down stream collectors.
		// cluster info retriever
		clusterInfoRetriever, ok := clusterRetrieverMap[target]
		if !ok {
			clusterInfoRetriever = clusterinfo.New(logger, httpClient, esURL, *esClusterInfoInterval)
			clusterRetrieverMap[target] = clusterInfoRetriever

			if *esExportIndices || *esExportShards {
				sC := collector.NewShards(logger, httpClient, esURL)
				registry.MustRegister(sC)
				iC := collector.NewIndices(logger, httpClient, esURL, *esExportShards, *esExportIndexAliases)
				registry.MustRegister(iC)
				if registerErr := clusterInfoRetriever.RegisterConsumer(iC); registerErr != nil {
					level.Error(logger).Log("msg", "failed to register indices collector in cluster info", registerErr)
					http.Error(w, "failed to register indices collector in cluster info", http.StatusInternalServerError)
					return
				}
				if registerErr := clusterInfoRetriever.RegisterConsumer(sC); registerErr != nil {
					level.Error(logger).Log("msg", "failed to register shards collector in cluster info", registerErr)
					http.Error(w, "failed to register shards collector in cluster info", http.StatusInternalServerError)
					return
				}
			}

			// start the cluster info retriever
			switch runErr := clusterInfoRetriever.Run(ctx); runErr {
			case nil:
				level.Info(logger).Log(
					"msg", fmt.Sprintf("[%s]started cluster info retriever", esURL.Host),
					"interval", (*esClusterInfoInterval).String(),
				)
			case clusterinfo.ErrInitialCallTimeout:
				level.Info(logger).Log("msg", fmt.Sprintf("[%s]initial cluster info call timed out", esURL.Host))
			default:
				level.Error(logger).Log("msg", fmt.Sprintf("[%s]failed to run cluster info retriever", esURL.Host), "err", err)
			}
		}

		registry.MustRegister(collector.NewClusterHealth(logger, httpClient, esURL))
		registry.MustRegister(collector.NewNodes(logger, httpClient, esURL, *esAllNodes, *esNode))

		if *esExportSLM {
			registry.MustRegister(collector.NewSLM(logger, httpClient, esURL))
		}

		if *esExportDataStream {
			registry.MustRegister(collector.NewDataStream(logger, httpClient, esURL))
		}

		if *esExportIndicesSettings {
			registry.MustRegister(collector.NewIndicesSettings(logger, httpClient, esURL))
		}

		if *esExportIndicesMappings {
			registry.MustRegister(collector.NewIndicesMappings(logger, httpClient, esURL))
		}

		if *esExportILM {
			registry.MustRegister(collector.NewIlmStatus(logger, httpClient, esURL))
			registry.MustRegister(collector.NewIlmIndicies(logger, httpClient, esURL))
		}

		// register cluster info retriever as prometheus collector
		registry.MustRegister(clusterInfoRetriever)

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	})

	// health endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
	})

	server := &http.Server{}
	go func() {
		if err = web.ListenAndServe(server, toolkitFlags, logger); err != nil {
			level.Error(logger).Log("msg", "http server quit", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	level.Info(logger).Log("msg", "shutting down")
	// create a context for graceful http server shutdown
	srvCtx, srvCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer srvCancel()
	_ = server.Shutdown(srvCtx)
}
