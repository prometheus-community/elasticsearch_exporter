package main

import (
	"net/http"
	"net/url"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/justwatchcom/elasticsearch_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		Name          = "elasticsearch_exporter"
		listenAddress = kingpin.Flag("web.listen-address",
			"Address to listen on for web interface and telemetry.").
			Default(":9108").String()
		metricsPath = kingpin.Flag("web.telemetry-path",
			"Path under which to expose metrics.").
			Default("/metrics").String()
		esURI = kingpin.Flag("es.uri",
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
		esExportIndicesSettings) = kingpin.Flag("es.indices_settings"),
			"Export stats for settings of all indices of the cluster.").
			Default("false").Bool()
		esExportClusterSettings = kingpin.Flag("es.cluster_settings",
			"Export stats for cluster settings.").
			Default("false").Bool()
		esExportShards = kingpin.Flag("es.shards",
			"Export stats for shards in the cluster (implies --es.indices).").
			Default("false").Bool()
		esExportSnapshots = kingpin.Flag("es.snapshots",
			"Export stats for the cluster snapshots.").
			Default("false").Bool()
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
			Default("infor").String()
		logFormat = kingpin.Flag("log.format",
			"Sets the log format. Valid formats are json and logfmt").
			Default("logfmt").String()
		logOutput = kingpin.Flag("log.output",
			"Sets the log output. Valid outputs are stdout and stderr").
			Default("stdout").String()
>>>>>>> Switch from flag to kingpin fixes #173
	)

	kingpin.Version(version.Print(Name))
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	logger := getLogger(*logLevel, *logOutput, *logFormat)

	esURIEnv, ok := os.LookupEnv("ES_URI")
	if ok {
		*esURI = esURIEnv
	}
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

	httpClient := &http.Client{
		Timeout: *esTimeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
		},
	}

	// version metric
	versionMetric := version.NewCollector(Name)
	prometheus.MustRegister(versionMetric)
	prometheus.MustRegister(collector.NewClusterHealth(logger, httpClient, esURL))
	prometheus.MustRegister(collector.NewNodes(logger, httpClient, esURL, *esAllNodes, *esNode))
	if *esExportIndices || *esExportShards {
		prometheus.MustRegister(collector.NewIndices(logger, httpClient, esURL, *esExportShards))
	}
	if *esExportSnapshots {
		prometheus.MustRegister(collector.NewSnapshots(logger, httpClient, esURL))
	}
	if *esExportClusterSettings {
		prometheus.MustRegister(collector.NewClusterSettings(logger, httpClient, esURL))
	}
	if *esExportIndicesSettings {
		prometheus.MustRegister(collector.NewIndicesSettings(logger, httpClient, esURL))
	}
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	_ = level.Info(logger).Log(
		"msg", "starting elasticsearch_exporter",
		"addr", *listenAddress,
	)

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		_ = level.Error(logger).Log(
			"msg", "http server quit",
			"err", err,
		)
	}
}
