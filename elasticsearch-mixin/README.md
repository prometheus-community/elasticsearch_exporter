# Elasticsearch Exporter Mixin

This is a mixin for the elasticsearch_exporter to define dashboards, alerts, and monitoring queries for use with this exporter.

Good example of upstream mixin for reference: https://github.com/kubernetes-monitoring/kubernetes-mixin

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

### Run the build
```bash
./scripts/compile-mixin.sh
```
