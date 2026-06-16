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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	prometheuscollectors "github.com/prometheus/client_golang/prometheus/collectors"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/prometheus-community/elasticsearch_exporter/collector"
	"github.com/prometheus-community/elasticsearch_exporter/config"
)

const name = "elasticsearch_exporter"

func main() {
	var (
		metricsPath = kingpin.Flag("web.telemetry-path",
			"Path under which to expose metrics.").
			Default("/metrics").String()
		toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9114")
		esURI        = kingpin.Flag("es.uri",
			"HTTP API address of an Elasticsearch node.").
			Default(config.DefaultElasticsearchURL).String()
		esTimeout = kingpin.Flag("es.timeout",
			"Timeout for trying to get stats from Elasticsearch.").
			Default(config.DefaultTimeout.String()).Duration()
		esAllNodes = kingpin.Flag("es.all",
			"Export stats for all nodes in the cluster. If used, this flag will override the flag es.node.").
			Default(strconv.FormatBool(config.DefaultAllNodes)).Bool()
		esNode = kingpin.Flag("es.node",
			"Node's name of which metrics should be exposed.").
			Default(config.DefaultNode).String()
		esExportIndices = kingpin.Flag("es.indices",
			"Export stats for indices in the cluster.").
			Default(strconv.FormatBool(config.DefaultExportIndices)).Bool()
		esExportIndicesSettings = kingpin.Flag("es.indices_settings",
			"Export stats for settings of all indices of the cluster.").
			Default(strconv.FormatBool(config.DefaultCollectorConfig()[config.CollectorIndicesSettings])).Bool()
		esExportIndicesMappings = kingpin.Flag("es.indices_mappings",
			"Export stats for mappings of all indices of the cluster.").
			Default(strconv.FormatBool(config.DefaultExportIndicesMappings)).Bool()
		esExportIndexAliases = kingpin.Flag("es.aliases",
			"Export informational alias metrics.").
			Default(strconv.FormatBool(config.DefaultExportIndexAliases)).Bool()
		esExportShards = kingpin.Flag("es.shards",
			"Export stats for shards in the cluster (implies --es.indices).").
			Default(strconv.FormatBool(config.DefaultExportShards)).Bool()
		esClusterInfoInterval = kingpin.Flag("es.clusterinfo.interval",
			"Cluster info update interval for the cluster label").
			Default(config.DefaultClusterInfoInterval.String()).Duration()
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
		configFile     = kingpin.Flag("config.file", "Path to YAML configuration file.").Default("").String()
		collectorFlags = newCollectorFlags()
		tasksActions   = kingpin.Flag("tasks.actions", "Filter on task actions. Used in same way as Task API actions param.").Default(config.DefaultTasksActions).String()
	)

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print(name))
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	// Load optional YAML auth module config.
	authCfg := &config.AuthConfig{AuthModules: map[string]config.AuthModule{}}
	if *configFile != "" {
		loadedCfg, cfgErr := config.LoadAuthConfig(*configFile)
		if cfgErr != nil {
			// At this stage logger not yet created; fallback to stderr
			fmt.Fprintf(os.Stderr, "failed to load config file: %v\n", cfgErr)
			os.Exit(1)
		}
		authCfg = loadedCfg
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

	// Create a context that is cancelled on SIGKILL or SIGINT.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	collectorStates := collectorFlags.states()
	if *esExportIndicesSettings {
		collectorStates[config.CollectorIndicesSettings] = true
	}
	baseCfg := buildConfig(
		*esURI,
		*esTimeout,
		*esAllNodes,
		*esNode,
		*esExportIndices,
		*esExportIndicesMappings,
		*esExportIndexAliases,
		*esExportShards,
		*esClusterInfoInterval,
		*esCA,
		*esClientCert,
		*esClientPrivateKey,
		*esInsecureSkipVerify,
		*awsRegion,
		*awsRoleArn,
		collectorStates,
		*tasksActions,
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		prometheuscollectors.NewGoCollector(),
		prometheuscollectors.NewProcessCollector(prometheuscollectors.ProcessCollectorOpts{}),
		versioncollector.NewCollector(name),
	)
	var runtime *collector.Runtime
	if baseCfg.ElasticsearchURL != "" {
		singleTargetCfg := baseCfg
		if username, password := os.Getenv("ES_USERNAME"), os.Getenv("ES_PASSWORD"); username != "" && password != "" {
			singleTargetCfg.Username = username
			singleTargetCfg.Password = password
		}
		singleTargetCfg.APIKey = os.Getenv("ES_API_KEY")
		if err := singleTargetCfg.Validate(); err != nil {
			logger.Error("invalid single-target configuration", "err", err)
			os.Exit(1)
		}
		var err error
		runtime, err = collector.NewRuntime(ctx, logger, singleTargetCfg)
		if err != nil {
			logger.Error("failed to create runtime", "err", err)
			os.Exit(1)
		}
		defer func() {
			if err := runtime.Close(); err != nil {
				logger.Error("failed to close runtime", "err", err)
			}
		}()
		if err := runtime.Start(ctx); err != nil {
			logger.Error("failed to start runtime", "err", err)
			os.Exit(1)
		}
		collectors, err := runtime.Collectors()
		if err != nil {
			logger.Error("failed to build collectors", "err", err)
			os.Exit(1)
		}
		for _, c := range collectors {
			registry.MustRegister(c)
		}
	}

	http.HandleFunc(*metricsPath, func(w http.ResponseWriter, r *http.Request) {
		// /metrics endpoint is reserved for single-target mode only.
		// For per-scrape overrides use the dedicated /probe endpoint.
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
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
		targetStr, am, valErr := validateProbeParams(authCfg, origQuery)
		if valErr != nil {
			http.Error(w, valErr.Error(), http.StatusBadRequest)
			return
		}
		probeCfg := baseCfg
		probeCfg.ElasticsearchURL = targetStr
		if err := applyProbeAuthModule(&probeCfg, am); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := probeCfg.Validate(); err != nil {
			logger.Error("invalid probe configuration", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		probeRuntime, err := collector.NewRuntime(r.Context(), logger.With("target", targetStr), probeCfg)
		if err != nil {
			logger.Error("failed to create probe runtime", "err", err)
			http.Error(w, "failed to create exporter", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := probeRuntime.Close(); err != nil {
				logger.Error("failed to close probe runtime", "err", err)
			}
		}()
		reg := prometheus.NewRegistry()
		reg.MustRegister(versioncollector.NewCollector(name))
		collectors, err := probeRuntime.Collectors()
		if err != nil {
			logger.Error("failed to build probe collectors", "err", err)
			http.Error(w, "failed to build collectors", http.StatusInternalServerError)
			return
		}
		for _, c := range collectors {
			reg.MustRegister(c)
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

type collectorFlagSet map[string]*bool

func newCollectorFlags() collectorFlagSet {
	defaults := config.DefaultCollectorConfig()
	names := make([]string, 0, len(defaults))
	for name := range defaults {
		names = append(names, name)
	}
	sort.Strings(names)

	flags := make(collectorFlagSet, len(defaults))
	for _, name := range names {
		helpDefaultState := "disabled"
		if defaults[name] {
			helpDefaultState = "enabled"
		}
		flags[name] = kingpin.Flag(
			"collector."+name,
			fmt.Sprintf("Enable the %s collector (default: %s).", name, helpDefaultState),
		).Default(fmt.Sprintf("%v", defaults[name])).Bool()
	}
	return flags
}

func (flags collectorFlagSet) states() map[string]bool {
	states := make(map[string]bool, len(flags))
	for name, value := range flags {
		states[name] = *value
	}
	return states
}

func buildConfig(esURI string, timeout time.Duration, allNodes bool, node string, exportIndices bool, exportIndicesMappings bool, exportIndexAliases bool, exportShards bool, clusterInfoInterval time.Duration, caFile string, certFile string, keyFile string, insecureSkipVerify bool, awsRegion string, awsRoleArn string, collectors map[string]bool, tasksActions string) config.Config {
	cfg := config.NewConfigWithDefaults()
	cfg.ElasticsearchURL = esURI
	cfg.Timeout = timeout
	cfg.AllNodes = allNodes
	cfg.Node = node
	cfg.ExportIndices = exportIndices
	cfg.ExportIndicesMappings = exportIndicesMappings
	cfg.ExportIndexAliases = exportIndexAliases
	cfg.ExportShards = exportShards
	cfg.ClusterInfoInterval = clusterInfoInterval
	cfg.TLS = config.TLSConfig{
		CAFile:             caFile,
		CertFile:           certFile,
		KeyFile:            keyFile,
		InsecureSkipVerify: insecureSkipVerify,
	}
	cfg.AWS = config.AWSConfig{
		Region:  awsRegion,
		RoleARN: awsRoleArn,
	}
	cfg.AWSEnabled = awsRegion != ""
	cfg.Collectors = collectors
	cfg.TasksActions = tasksActions
	return cfg
}

func applyProbeAuthModule(cfg *config.Config, module *config.AuthModule) error {
	if module == nil {
		return nil
	}
	targetURL, err := url.Parse(cfg.ElasticsearchURL)
	if err != nil {
		return err
	}
	for k, v := range module.Options {
		q := targetURL.Query()
		q.Set(k, v)
		targetURL.RawQuery = q.Encode()
	}
	cfg.ElasticsearchURL = targetURL.String()
	if module.TLS != nil {
		if module.TLS.CAFile != "" {
			cfg.TLS.CAFile = module.TLS.CAFile
		}
		if module.TLS.CertFile != "" {
			cfg.TLS.CertFile = module.TLS.CertFile
		}
		if module.TLS.KeyFile != "" {
			cfg.TLS.KeyFile = module.TLS.KeyFile
		}
		if module.TLS.InsecureSkipVerify {
			cfg.TLS.InsecureSkipVerify = true
		}
	}
	switch strings.ToLower(module.Type) {
	case "userpass":
		if module.UserPass != nil {
			cfg.Username = module.UserPass.Username
			cfg.Password = module.UserPass.Password
		}
	case "apikey":
		cfg.APIKey = module.APIKey
	case "aws":
		if module.AWS != nil {
			cfg.AWS = *module.AWS
		}
		cfg.AWSEnabled = true
	case "tls":
	default:
		return fmt.Errorf("unsupported auth_module type %s", module.Type)
	}
	return nil
}
