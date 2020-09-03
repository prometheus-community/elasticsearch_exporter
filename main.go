package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/justwatchcom/elasticsearch_exporter/collector"
	"github.com/justwatchcom/elasticsearch_exporter/pkg/clusterinfo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		Name          = "elasticsearch_exporter"
		listenAddress = kingpin.Flag("web.listen-address",
			"Address to listen on for web interface and telemetry.").
			Default(":9114").Envar("WEB_LISTEN_ADDRESS").String()
		metricsPath = kingpin.Flag("web.telemetry-path",
			"Path under which to expose metrics.").
			Default("/metrics").Envar("WEB_TELEMETRY_PATH").String()
		esURI = kingpin.Flag("es.uri",
			"HTTP API address of an Elasticsearch node.").
			Default("http://localhost:9200").Envar("ES_URI").String()
		esTimeout = kingpin.Flag("es.timeout",
			"Timeout for trying to get stats from Elasticsearch.").
			Default("5s").Envar("ES_TIMEOUT").Duration()
		esAllNodes = kingpin.Flag("es.all",
			"Export stats for all nodes in the cluster. If used, this flag will override the flag es.node.").
			Default("false").Envar("ES_ALL").Bool()
		esNode = kingpin.Flag("es.node",
			"Node's name of which metrics should be exposed.").
			Default("_local").Envar("ES_NODE").String()
		esNodesInterval = kingpin.Flag("es.nodes.interval",
			"Node stats metrics update interval").
			Default("30s").Envar("ES_NODES_INTERVAL").Duration()
		esExportNodesHTTP = kingpin.Flag("es.nodehttp",
			"Export stats for node HTTP in the cluster.").
			Default("false").Envar("ES_NODE_HTTP").Bool()
		esExportIndices = kingpin.Flag("es.indices",
			"Export stats for indices in the cluster.").
			Default("false").Envar("ES_INDICES").Bool()
		esIndicesInterval = kingpin.Flag("es.indices.interval",
			"Node stats metrics update interval").
			Default("30s").Envar("ES_INDICES_INTERVAL").Duration()
		esExportIndicesSettings = kingpin.Flag("es.indices_settings",
			"Export stats for settings of all indices of the cluster.").
			Default("false").Envar("ES_INDICES_SETTINGS").Bool()
		esExportClusterSettings = kingpin.Flag("es.cluster_settings",
			"Export stats for cluster settings.").
			Default("false").Envar("ES_CLUSTER_SETTINGS").Bool()
		esExportShards = kingpin.Flag("es.shards",
			"Export stats for shards in the cluster (implies --es.indices).").
			Default("false").Envar("ES_SHARDS").Bool()
		esExportSnapshots = kingpin.Flag("es.snapshots",
			"Export stats for the cluster snapshots.").
			Default("false").Envar("ES_SNAPSHOTS").Bool()
		esSnapshotsInterval = kingpin.Flag("es.snapshots.interval",
			"Snapshots metrics update interval").
			Default("0s").Envar("ES_SNAPSHOTS_INTERVAL").Duration()
		esClusterInfoInterval = kingpin.Flag("es.clusterinfo.interval",
			"Cluster info update interval for the cluster label").
			Default("5m").Envar("ES_CLUSTERINFO_INTERVAL").Duration()
		esCA = kingpin.Flag("es.ca",
			"Path to PEM file that contains trusted Certificate Authorities for the Elasticsearch connection.").
			Default("").Envar("ES_CA").String()
		esClientPrivateKey = kingpin.Flag("es.client-private-key",
			"Path to PEM file that contains the private key for client auth when connecting to Elasticsearch.").
			Default("").Envar("ES_CLIENT_PRIVATE_KEY").String()
		esClientCert = kingpin.Flag("es.client-cert",
			"Path to PEM file that contains the corresponding cert for the private key to connect to Elasticsearch.").
			Default("").Envar("ES_CLIENT_CERT").String()
		esInsecureSkipVerify = kingpin.Flag("es.ssl-skip-verify",
			"Skip SSL verification when connecting to Elasticsearch.").
			Default("false").Envar("ES_SSL_SKIP_VERIFY").Bool()
		logLevel = kingpin.Flag("log.level",
			"Sets the loglevel. Valid levels are debug, info, warn, error").
			Default("info").Envar("LOG_LEVEL").String()
		logFormat = kingpin.Flag("log.format",
			"Sets the log format. Valid formats are json and logfmt").
			Default("logfmt").Envar("LOG_FMT").String()
		logOutput = kingpin.Flag("log.output",
			"Sets the log output. Valid outputs are stdout and stderr").
			Default("stdout").Envar("LOG_OUTPUT").String()
	)

	kingpin.Version(version.Print(Name))
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	logger := getLogger(*logLevel, *logOutput, *logFormat)

	// create a context that is cancelled on SIGKILL
	ctx, cancel := context.WithCancel(context.Background())

	esURL, err := url.Parse(*esURI)
	if err != nil {
		_ = level.Error(logger).Log(
			"msg", "failed to parse es.uri",
			"err", err,
		)
		os.Exit(1)
	}

	// returns nil if not provided and falls back to simple TCP.
	tlsConfig := createTLSConfig(*esCA, *esClientCert, *esClientPrivateKey, *esInsecureSkipVerify)
	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
	}
	httpClient := &http.Client{
		Timeout:   *esTimeout,
		Transport: httpTransport,
	}

	// version metric
	versionMetric := version.NewCollector(Name)
	prometheus.MustRegister(versionMetric)

	// cluster info retriever
	clusterInfoRetriever := clusterinfo.New(logger, httpClient, esURL, *esClusterInfoInterval)

	prometheus.MustRegister(collector.NewClusterHealth(logger, httpClient, esURL))
	prometheus.MustRegister(collector.NewNodes(ctx, logger, createClient(httpTransport, *esTimeout, *esNodesInterval*2), esURL, *esAllNodes, *esNode, *esNodesInterval))

	if *esExportNodesHTTP {
		prometheus.MustRegister(collector.NewNodesHTTP(logger, httpClient, esURL))
	}

	if *esExportIndices || *esExportShards {
		client := createClient(httpTransport, *esTimeout, *esIndicesInterval*2)
		iC := collector.NewIndices(ctx, logger, client, esURL, *esExportShards, *esIndicesInterval)
		prometheus.MustRegister(iC)
		if registerErr := clusterInfoRetriever.RegisterConsumer(iC); registerErr != nil {
			_ = level.Error(logger).Log("msg", "failed to register indices collector in cluster info")
			os.Exit(1)
		}
	}

	if *esExportSnapshots {
		client := createClient(httpTransport, *esTimeout, *esSnapshotsInterval)
		prometheus.MustRegister(collector.NewSnapshots(ctx, logger, client, esURL, *esSnapshotsInterval))
	}

	if *esExportClusterSettings {
		prometheus.MustRegister(collector.NewClusterSettings(logger, httpClient, esURL))
	}

	if *esExportIndicesSettings {
		prometheus.MustRegister(collector.NewIndicesSettings(logger, httpClient, esURL))
	}

	// create a http server
	server := &http.Server{}

	// start the cluster info retriever
	switch runErr := clusterInfoRetriever.Run(ctx); runErr {
	case nil:
		_ = level.Info(logger).Log(
			"msg", "started cluster info retriever",
			"interval", (*esClusterInfoInterval).String(),
		)
	case clusterinfo.ErrInitialCallTimeout:
		_ = level.Info(logger).Log("msg", "initial cluster info call timed out")
	default:
		_ = level.Error(logger).Log("msg", "failed to run cluster info retriever", "err", err)
		os.Exit(1)
	}

	// register cluster info retriever as prometheus collector
	prometheus.MustRegister(clusterInfoRetriever)

	mux := http.DefaultServeMux
	mux.Handle(*metricsPath, promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err = w.Write([]byte(`<html>
			<head><title>Elasticsearch Exporter</title></head>
			<body>
			<h1>Elasticsearch Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			_ = level.Error(logger).Log(
				"msg", "failed handling writer",
				"err", err,
			)
		}
	})

	// health endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
	})

	server.Handler = mux
	server.Addr = *listenAddress

	_ = level.Info(logger).Log(
		"msg", "starting elasticsearch_exporter",
		"addr", *listenAddress,
	)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			_ = level.Error(logger).Log(
				"msg", "http server quit",
				"err", err,
			)
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// create a context for graceful http server shutdown
	srvCtx, srvCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer srvCancel()
	<-c
	_ = level.Info(logger).Log("msg", "shutting down")
	_ = server.Shutdown(srvCtx)
	cancel()
}

func createClient(transport *http.Transport, globalTimeout, timeout time.Duration) *http.Client {
	var c = &http.Client{
		Timeout:   globalTimeout,
		Transport: transport,
	}
	if timeout != 0 {
		c.Timeout = timeout
	}
	return c
}
