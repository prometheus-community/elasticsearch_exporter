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
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"context"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/elasticsearch_exporter/collector"
	"github.com/prometheus-community/elasticsearch_exporter/pkg/clusterinfo"
	"github.com/prometheus-community/elasticsearch_exporter/pkg/roundtripper"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
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
			"Export index lifecycle policies for indices in the cluster.").
			Default("false").Bool()
		esExportShards = kingpin.Flag("es.shards",
			"Export stats for shards in the cluster (implies --es.indices).").
			Default("false").Bool()
		esExportDataStream = kingpin.Flag("es.data_stream",
			"Export stats for Data Streams.").
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
	)

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print(name))
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

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
		httpClient.Transport, err = roundtripper.NewAWSSigningTransport(httpTransport, *awsRegion, *awsRoleArn, logger)
		if err != nil {
			logger.Error("failed to create AWS transport", "err", err)
			os.Exit(1)
		}
	}

	// version metric
	prometheus.MustRegister(versioncollector.NewCollector(name))

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

	if *esExportDataStream {
		prometheus.MustRegister(collector.NewDataStream(logger, httpClient, esURL))
	}

	if *esExportIndicesSettings {
		prometheus.MustRegister(collector.NewIndicesSettings(logger, httpClient, esURL))
	}

	if *esExportIndicesMappings {
		prometheus.MustRegister(collector.NewIndicesMappings(logger, httpClient, esURL))
	}

	if *esExportILM {
		prometheus.MustRegister(collector.NewIlmStatus(logger, httpClient, esURL))
		prometheus.MustRegister(collector.NewIlmIndicies(logger, httpClient, esURL))
	}

	// Create a context that is cancelled on SIGKILL or SIGINT.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// start the cluster info retriever
	switch runErr := clusterInfoRetriever.Run(ctx); runErr {
	case nil:
		logger.Info("started cluster info retriever", "interval", (*esClusterInfoInterval).String())
	case clusterinfo.ErrInitialCallTimeout:
		logger.Info("initial cluster info call timed out")
	default:
		logger.Error("failed to run cluster info retriever", "err", err)
		os.Exit(1)
	}

	// register cluster info retriever as prometheus collector
	prometheus.MustRegister(clusterInfoRetriever)

	http.Handle(*metricsPath, promhttp.Handler())
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

	server := &http.Server{}
	go func() {
		if err = web.ListenAndServe(server, toolkitFlags, logger); err != nil {
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
