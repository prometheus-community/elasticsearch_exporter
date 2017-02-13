package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "elasticsearch"
	indexHTML = `
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
	</html>`
)

func getESVersion(esURI *string) string {
	resp, _ := http.Get(*esURI)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var clusterInfo struct {
		Version struct {
			Number string
		}
	}
	json.Unmarshal(body, &clusterInfo)
	return clusterInfo.Version.Number
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9108", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		esURI         = flag.String("es.uri", "http://localhost:9200", "HTTP API address of an Elasticsearch node.")
		esTimeout     = flag.Duration("es.timeout", 5*time.Second, "Timeout for trying to get stats from Elasticsearch.")
		esAllNodes    = flag.Bool("es.all", false, "Export stats for all nodes in the cluster.")
	)
	flag.Parse()

	esVersion := getESVersion(esURI)

	if *esAllNodes {
		*esURI = *esURI + "/_nodes/stats"
	} else {
		*esURI = *esURI + "/_nodes/_local/stats"
	}

	exporter := NewExporter(*esURI, *esTimeout, *esAllNodes, esVersion)
	prometheus.MustRegister(exporter)

	log.Println("Starting Server:", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(indexHTML, *metricsPath)))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
