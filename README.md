# Elasticsearch Exporter

[![CircleCI](https://circleci.com/gh/prometheus-community/elasticsearch_exporter.svg?style=svg)](https://circleci.com/gh/prometheus-community/elasticsearch_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/prometheus-community/elasticsearch_exporter)](https://goreportcard.com/report/github.com/prometheus-community/elasticsearch_exporter)

Prometheus exporter for various metrics about Elasticsearch and OpenSearch, written in Go.

## Supported Versions

We support all currently supported versions of Elasticsearch and OpenSearch. This project will make reasonable attempts to maintain compatibility with previous versions but considerations will be made for code maintainability and favoring supported versions. Where Elasticsearch and OpenSearch diverge, this project will make reasonable attempts to maintain compatibility with both. Some collectors may only be compatible with one or the other.

### Installation

For pre-built binaries please take a look at the releases.
<https://github.com/prometheus-community/elasticsearch_exporter/releases>

#### Docker

```bash
docker pull quay.io/prometheuscommunity/elasticsearch-exporter:latest
docker run --rm -p 9114:9114 quay.io/prometheuscommunity/elasticsearch-exporter:latest
```

Example `docker-compose.yml`:

```yaml
elasticsearch_exporter:
    image: quay.io/prometheuscommunity/elasticsearch-exporter:latest
    command:
     - '--es.uri=http://elasticsearch:9200'
    restart: always
    ports:
    - "127.0.0.1:9114:9114"
```

#### Kubernetes

You can find a helm chart in the prometheus-community charts repository at <https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus-elasticsearch-exporter>

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install [RELEASE_NAME] prometheus-community/prometheus-elasticsearch-exporter
```

### Configuration

**NOTE:** The exporter fetches information from an Elasticsearch cluster on every scrape, therefore having a too short scrape interval can impose load on ES master nodes, particularly if you run with `--es.all` and `--es.indices`. We suggest you measure how long fetching `/_nodes/stats` and `/_all/_stats` takes for your ES cluster to determine whether your scraping interval is too short. As a last resort, you can scrape this exporter using a dedicated job with its own scraping interval.

Below is the command line options summary:

```bash
elasticsearch_exporter --help
```

| Argument                | Introduced in Version | Description                                                                                                                                                                                                                                                                                                                                                                           | Default     |
| ----------------------- | --------------------- |---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| ----------- |
| collector.clustersettings| 1.6.0                | If true, query stats for cluster settings (As of v1.6.0, this flag has replaced "es.cluster_settings").                                                                                                                                                                                                                                                                                | false |
| es.uri                  | 1.0.2                 | Address (host and port) of the Elasticsearch node we should connect to **when running in single-target mode**. Leave empty (the default) when you want to run the exporter only as a multi-target `/probe` endpoint. When basic auth is needed, specify as: `<proto>://<user>:<password>@<host>:<port>`. E.G., `http://admin:pass@localhost:9200`. Special characters in the user credentials need to be URL-encoded. | "" |
| es.all                  | 1.0.2                 | If true, query stats for all nodes in the cluster, rather than just the node we connect to.                                                                                                                                                                                                                                                                                           | false |
| es.indices              | 1.0.2                 | If true, query stats for all indices in the cluster.                                                                                                                                                                                                                                                                                                                                  | false |
| es.indices_settings     | 1.0.4rc1              | If true, query settings stats for all indices in the cluster.                                                                                                                                                                                                                                                                                                                         | false |
| es.indices_mappings     | 1.2.0                 | If true, query stats for mappings of all indices of the cluster.                                                                                                                                                                                                                                                                                                                      | false |
| es.aliases              | 1.0.4rc1              | If true, include informational aliases metrics.                                                                                                                                                                                                                                                                                                                                       | true |
| es.ilm                  | 1.6.0                 | If true, query index lifecycle policies for indices in the cluster.
| es.shards               | 1.0.3rc1              | If true, query stats for all indices in the cluster, including shard-level stats (implies `es.indices=true`).                                                                                                                                                                                                                                                                         | false |
| collector.snapshots     | 1.0.4rc1              | If true, query stats for the cluster snapshots. (As of v1.7.0, this flag has replaced "es.snapshots").                                                                                                                                                                                                                                                                                | false |
| collector.health-report | 1.10.0                 | If true, query the health report (requires elasticsearch 8.7.0 or later)                                                                                                                                                                                                                                                                                                              | false |
| es.slm                  |                       | If true, query stats for SLM.                                                                                                                                                                                                                                                                                                                                                         | false |
| es.data_stream          |                       | If true, query state for Data Steams.                                                                                                                                                                                                                                                                                                                                                 | false |
| es.timeout              | 1.0.2                 | Timeout for trying to get stats from Elasticsearch. (ex: 20s)                                                                                                                                                                                                                                                                                                                         | 5s |
| es.ca                   | 1.0.2                 | Path to PEM file that contains trusted Certificate Authorities for the Elasticsearch connection.                                                                                                                                                                                                                                                                                      | |
| es.client-private-key   | 1.0.2                 | Path to PEM file that contains the private key for client auth when connecting to Elasticsearch.                                                                                                                                                                                                                                                                                      | |
| es.client-cert          | 1.0.2                 | Path to PEM file that contains the corresponding cert for the private key to connect to Elasticsearch.                                                                                                                                                                                                                                                                                | |
| es.clusterinfo.interval | 1.1.0rc1              | Cluster info update interval for the cluster label                                                                                                                                                                                                                                                                                                                                    | 5m |
| es.ssl-skip-verify      | 1.0.4rc1              | Skip SSL verification when connecting to Elasticsearch.                                                                                                                                                                                                                                                                                                                               | false |
| web.listen-address      | 1.0.2                 | Address to listen on for web interface and telemetry.                                                                                                                                                                                                                                                                                                                                 | :9114 |
| web.telemetry-path      | 1.0.2                 | Path under which to expose metrics.                                                                                                                                                                                                                                                                                                                                                   | /metrics |
| aws.region              | 1.5.0                 | Region for AWS elasticsearch                                                                                                                                                                                                                                                                                                                                                          | |
| aws.role-arn            | 1.6.0                 | Role ARN of an IAM role to assume.                                                                                                                                                                                                                                                                                                                                                    | |
| config.file             | 1.10.0                 | Path to a YAML configuration file that defines `auth_modules:` used by the `/probe` multi-target endpoint. Leave unset when not using multi-target mode.                                                                                                                                                                                                                              | |
| version                 | 1.0.2                 | Show version info on stdout and exit.                                                                                                                                                                                                                                                                                                                                                 | |

Commandline parameters start with a single `-` for versions less than `1.1.0rc1`.
For versions greater than `1.1.0rc1`, commandline parameters are specified with `--`.

The API key used to connect can be set with the `ES_API_KEY` environment variable.

#### Logging

Logging by the exporter is handled by the `log/slog` package. The output format can be customized with the `--log.format` flag which defaults to logfmt. The log level can be set with the `--log.level` flag which defaults to info. The output can be set to either stdout (default) or stderr with the `--log.output` flag.

#### Elasticsearch 7.x security privileges

Username and password can be passed either directly in the URI or through the `ES_USERNAME` and `ES_PASSWORD` environment variables.
Specifying those two environment variables will override authentication passed in the URI (if any).

ES 7.x supports RBACs. The following security privileges are required for the elasticsearch_exporter.

Setting | Privilege Required | Description
:---- | :---- | :----
collector.clustersettings| `cluster` `monitor` |
exporter defaults | `cluster` `monitor` | All cluster read-only operations, like cluster health and state, hot threads, node info, node and cluster stats, and pending cluster tasks. |
es.indices | `indices` `monitor` (per index or `*`) | All actions that are required for monitoring (recovery, segments info, index stats and status)
es.indices_settings | `indices` `monitor` (per index or `*`) |
es.indices_mappings | `indices` `view_index_metadata` (per index or `*`) |
es.shards | not sure if `indices` or `cluster` `monitor` or both |
collector.snapshots | `cluster:admin/snapshot/status` and `cluster:admin/repository/get` | [ES Forum Post](https://discuss.elastic.co/t/permissions-for-backup-user-with-x-pack/88057)
es.slm | `manage_slm`
es.data_stream | `monitor` or `manage` (per index or `*`) |

Further Information

- [Built in Users](https://www.elastic.co/guide/en/elastic-stack-overview/7.3/built-in-users.html)
- [Defining Roles](https://www.elastic.co/guide/en/elastic-stack-overview/7.3/defining-roles.html)
- [Privileges](https://www.elastic.co/guide/en/elastic-stack-overview/7.3/security-privileges.html)

### Multi-Target Scraping (beta)

From v2.X the exporter exposes `/probe` allowing one running instance to scrape many clusters.

Supported `auth_module` types:

| type       | YAML fields                                                       | Injected into request                                                                 |
| ---------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| `userpass` | `userpass.username`, `userpass.password`, optional `options:` map | Sets HTTP basic-auth header, appends `options` as query parameters                    |
| `apikey`   | `apikey:` Base64 API-Key string, optional `options:` map          | Adds `Authorization: ApiKey …` header, appends `options`                              |
| `aws`      | `aws.region`, optional `aws.role_arn`, optional `options:` map    | Uses AWS SigV4 signing transport for HTTP(S) requests, appends `options`              |
| `tls`      | `tls.ca_file`, `tls.cert_file`, `tls.key_file`                    | Uses client certificate authentication via TLS; cannot be mixed with other auth types |

Example config:

```yaml
# exporter-config.yml
auth_modules:
  prod_basic:
    type: userpass
    userpass:
      username: metrics
      password: s3cr3t

  staging_key:
    type: apikey
    apikey: "bXk6YXBpa2V5Ig=="  # base64 id:key
    options:
      sslmode: disable
```

Run exporter:

```bash
./elasticsearch_exporter --config.file=exporter-config.yml
```

Prometheus scrape_config:

```yaml
- job_name: es
  metrics_path: /probe
  params:
    auth_module: [staging_key]
  static_configs:
    - targets: ["https://es-stage:9200"]
  relabel_configs:
    - source_labels: [__address__]
      target_label: __param_target
    - source_labels: [__param_target]
      target_label: instance
    - target_label: __address__
      replacement: exporter:9114
```

Notes:
- `/metrics` serves a single, process-wide registry and is intended for single-target mode.
- `/probe` creates a fresh registry per scrape for the given `target` allowing multi-target scraping.
- Any `options:` under an auth module will be appended as URL query parameters to the target URL.
- The `tls` auth module (client certificate authentication) is intended for self‑managed Elasticsearch/OpenSearch deployments. Amazon OpenSearch Service typically authenticates at the domain edge with IAM/SigV4 and does not support client certificate authentication; use the `aws` auth module instead when scraping Amazon OpenSearch Service domains.

### Metrics

See the [metrics documentation](metrics.md)

### Alerts & Recording Rules

We provide examples for [Prometheus](http://prometheus.io) [alerts and recording rules](examples/prometheus/elasticsearch.rules) as well as an [Grafana](http://www.grafana.org) [Dashboard](examples/grafana/dashboard.json) and a [Kubernetes](http://kubernetes.io) [Deployment](examples/kubernetes/deployment.yml).

The example dashboard needs the [node_exporter](https://github.com/prometheus/node_exporter) installed. In order to select the nodes that belong to the Elasticsearch cluster, we rely on a label `cluster`.
Depending on your setup, it can derived from the platform metadata:

For example on [GCE](https://cloud.google.com)

```
- source_labels: [__meta_gce_metadata_Cluster]
  separator: ;
  regex: (.*)
  target_label: cluster
  replacement: ${1}
  action: replace
```

Please refer to the [Prometheus SD documentation](https://prometheus.io/docs/operating/configuration/) to see which metadata labels can be used to create the `cluster` label.

## Credit & License

`elasticsearch_exporter` is maintained by the [Prometheus Community](https://www.prometheus.io/community/).

`elasticsearch_exporter` was then maintained by the nice folks from [JustWatch](https://www.justwatch.com/).
Then transferred this repository to the Prometheus Community in May 2021.

This package was originally created and maintained by [Eric Richardson](https://github.com/ewr),
who transferred this repository to us in January 2017.

Please refer to the Git commit log for a complete list of contributors.

## Contributing

We welcome any contributions. Please fork the project on GitHub and open
Pull Requests for any proposed changes.

Please note that we will not merge any changes that encourage insecure
behaviour. If in doubt please open an Issue first to discuss your proposal.
