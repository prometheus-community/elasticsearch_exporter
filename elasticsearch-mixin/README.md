# Elasticsearch Exporter Mixin

This is a mixin for the elasticsearch_exporter to define dashboards, alerts, and monitoring queries for use with this exporter.

Good example of upstream mixin for reference: https://github.com/kubernetes-monitoring/kubernetes-mixin


docker-compose
- docker-compose exec  elasticsearch bash
  - bin/elasticsearch-reset-password -u elastic -f
- login to grafana
- add prometheus datasource (http://prometheus:9090)
- http://127.0.0.1:3000
- http://127.0.0.1:9090/targets?search=
- http://127.0.0.1:9114/metrics

## Development

### JSONNET
https://jsonnet.org/

```go install github.com/google/go-jsonnet/cmd/jsonnet@latest```

### JSONNET BUNDLER
jsonnet bundler is a package manager for jsonnet

https://github.com/jsonnet-bundler/jsonnet-bundler

```go install -a github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@latest```

### Grafonnet
Grafana libraries for jsonnet: https://grafana.github.io/grafonnet/

```jb install github.com/grafana/grafonnet/gen/grafonnet-latest@main```

validate
go install github.com/grafana/dashboard-linter@latest
