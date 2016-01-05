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

* __`es.uri`:__ Address (host and port) of the Elasticsearch node we should
    connect to. This could be a local node (`localhost:8500`, for instance), or
    the address of a remote Elasticsearch server.
* __`es.all`:__ If true, query stats for all nodes in the cluster,
    rather than just the node we connect to.
* __`es.timeout`:__ Timeout for trying to get stats from Elasticsearch. (ex: 20s)
* __`web.listen-address`:__ Address to listen on for web interface and telemetry.
* __`web.telemetry-path`:__ Path under which to expose metrics.

__NOTE:__ We support pulling stats for all nodes at once, but in production
this is unlikely to be the way you actually want to run the system. It is much
better to run an exporter on each Elasticsearch node to remove a single point
of failure and improve the connection between operation and reporting.

### Elasticsearch 2.0

Parts of the node stats struct changed for Elasticsearch 2.0. For the moment
we'll attempt to report important values for both.

* `indices.filter_cache` becomes `indices.query_cache`
* `indices.query_cache` becomes `indices.request_cache`
* `process.cpu` lost `user` and `sys` time, so we're now reporting `total`
* Added `process.cpu.max_file_descriptors`
* 
