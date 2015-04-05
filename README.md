# Elasticsearch Exporter

Export Elasticsearch service health to Prometheus.

To run it:

```bash
make
./elasticsearch_exporter [flags]
```

### Flags

```bash
./elasticsearch_exporter --help
```

* __`es.server`:__ Address (host and port) of the Elasticsearch node we should
    connect to. This could be a local node (`localhost:8500`, for instance), or
    the address of a remote Elasticsearch server.
* __`es.cluster_nodes`:__ If true, query stats for all nodes in the cluster,
    rather than just the node we connect to.
* __`web.listen-address`:__ Address to listen on for web interface and telemetry.
* __`web.telemetry-path`:__ Path under which to expose metrics.

