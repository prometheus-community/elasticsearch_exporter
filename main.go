package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

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

func main() {
	var (
		listenAddress      = flag.String("web.listen-address", ":9108", "Address to listen on for web interface and telemetry.")
		metricsPath        = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		esURI              = flag.String("es.uri", "http://localhost:9200", "HTTP API address of an Elasticsearch node.")
		esTimeout          = flag.Duration("es.timeout", 5*time.Second, "Timeout for trying to get stats from Elasticsearch.")
		esAllNodes         = flag.Bool("es.all", false, "Export stats for all nodes in the cluster.")
		esCA               = flag.String("es.ca", "", "Path to PEM file that conains trusted CAs for the Elasticsearch connection.")
		esClientPrivateKey = flag.String("es.client-private-key", "", "Path to PEM file that conains the private key for client auth when connecting to Elasticsearch.")
		esClientCert       = flag.String("es.client-cert", "", "Path to PEM file that conains the corresponding cert for the private key to connect to Elasticsearch.")
	)
	flag.Parse()

	nodesStatsURI := *esURI + "/_nodes/_local/stats"
	if *esAllNodes {
		nodesStatsURI = *esURI + "/_nodes/stats"
	}
	clusterHealthURI := *esURI + "/_cluster/health"

	exporter := NewExporter(nodesStatsURI, clusterHealthURI, *esTimeout, *esAllNodes, createElasticSearchTLSConfig(*esCA, *esClientCert, *esClientPrivateKey))
	prometheus.MustRegister(exporter)

	log.Println("Starting Server:", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(indexHTML, *metricsPath)))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func createElasticSearchTLSConfig(pemFile, pemCertFile, pemPrivateKeyFile string) *tls.Config {
	if len(pemFile) <= 0 {
		return nil
	}
	rootCerts, err := loadCertificatesFrom(pemFile)
	if err != nil {
		log.Fatalf("Couldn't load root certificate from %s. Got %s.", pemFile, err)
	}
	if len(pemCertFile) > 0 && len(pemPrivateKeyFile) > 0 {
		clientPrivateKey, err := loadPrivateKeyFrom(pemCertFile, pemPrivateKeyFile)
		if err != nil {
			log.Fatalf("Couldn't setup client authentication. Got %s.", err)
		}
		return &tls.Config{
			RootCAs:      rootCerts,
			Certificates: []tls.Certificate{*clientPrivateKey},
		}
	}
	return &tls.Config{
		RootCAs: rootCerts,
	}
}

func loadCertificatesFrom(pemFile string) (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(pemFile)
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
