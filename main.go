package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/justwatchcom/elasticsearch_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	var (
		listenAddress      = flag.String("web.listen-address", ":9108", "Address to listen on for web interface and telemetry.")
		metricsPath        = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		esURI              = flag.String("es.uri", "http://localhost:9200", "HTTP API address of an Elasticsearch node.")
		esTimeout          = flag.Duration("es.timeout", 5*time.Second, "Timeout for trying to get stats from Elasticsearch.")
		esAllNodes         = flag.Bool("es.all", false, "Export stats for all nodes in the cluster.")
		esExportIndices    = flag.Bool("es.indices", false, "Export stats for indices in the cluster.")
		esCA               = flag.String("es.ca", "", "Path to PEM file that conains trusted CAs for the Elasticsearch connection.")
		esClientPrivateKey = flag.String("es.client-private-key", "", "Path to PEM file that conains the private key for client auth when connecting to Elasticsearch.")
		esClientCert       = flag.String("es.client-cert", "", "Path to PEM file that conains the corresponding cert for the private key to connect to Elasticsearch.")
	)
	flag.Parse()

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	esURL, err := url.Parse(*esURI)
	if err != nil {
		level.Error(logger).Log(
			"msg", "failed to parse es.uri",
			"err", err,
		)
		os.Exit(1)
	}

	// returns nil if not provided and falls back to simple TCP.
	tlsConfig := createTLSConfig(*esCA, *esClientCert, *esClientPrivateKey)

	httpClient := &http.Client{
		Timeout: *esTimeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	prometheus.MustRegister(collector.NewClusterHealth(logger, httpClient, esURL))
	prometheus.MustRegister(collector.NewNodes(logger, httpClient, esURL, *esAllNodes))
	prometheus.MustRegister(collector.NewIndices(logger, httpClient, esURL, *esAllNodes, *esExportIndices))

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", IndexHandler(*metricsPath))

	level.Info(logger).Log(
		"msg", "starting elasticsearch_exporter",
		"addr", *listenAddress,
	)

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log(
			"msg", "http server quit",
			"err", err,
		)
	}
}

// IndexHandler returns a http handler with the correct metricsPath
func IndexHandler(metricsPath string) http.HandlerFunc {
	indexHTML := `
<html>
	<head>
		<title>Elasticsearch Exporter</title>
	</head>
	<body>
		<h1>Elasticsearch Exporter</h1>
		<p>
			<a href='%s'>Metrics</a>
		</p>
	</body>
</html>
`
	index := []byte(fmt.Sprintf(strings.TrimSpace(indexHTML), metricsPath))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	}
}
