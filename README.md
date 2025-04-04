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
| es.uri                  | 1.0.2                 | Address (host and port) of the Elasticsearch node we should connect to. This could be a local node (`localhost:9200`, for instance), or the address of a remote Elasticsearch server. When basic auth is needed, specify as: `<proto>://<user>:<password>@<host>:<port>`. E.G., `http://admin:pass@localhost:9200`. Special characters in the user credentials need to be URL-encoded. | <http://localhost:9200> |
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

### Metrics

| Name                                                                 | Type       | Cardinality | Help                                                                                                |
|----------------------------------------------------------------------|------------|-------------|-----------------------------------------------------------------------------------------------------|
| elasticsearch_breakers_estimated_size_bytes                          | gauge      | 4           | Estimated size in bytes of breaker                                                                  |
| elasticsearch_breakers_limit_size_bytes                              | gauge      | 4           | Limit size in bytes for breaker                                                                     |
| elasticsearch_breakers_tripped                                       | counter    | 4           | tripped for breaker                                                                                 |
| elasticsearch_cluster_health_active_primary_shards                   | gauge      | 1           | The number of primary shards in your cluster. This is an aggregate total across all indices.        |
| elasticsearch_cluster_health_active_shards                           | gauge      | 1           | Aggregate total of all shards across all indices, which includes replica shards.                    |
| elasticsearch_cluster_health_delayed_unassigned_shards               | gauge      | 1           | Shards delayed to reduce reallocation overhead                                                      |
| elasticsearch_cluster_health_initializing_shards                     | gauge      | 1           | Count of shards that are being freshly created.                                                     |
| elasticsearch_cluster_health_number_of_data_nodes                    | gauge      | 1           | Number of data nodes in the cluster.                                                                |
| elasticsearch_cluster_health_number_of_in_flight_fetch               | gauge      | 1           | The number of ongoing shard info requests.                                                          |
| elasticsearch_cluster_health_number_of_nodes                         | gauge      | 1           | Number of nodes in the cluster.                                                                     |
| elasticsearch_cluster_health_number_of_pending_tasks                 | gauge      | 1           | Cluster level changes which have not yet been executed                                              |
| elasticsearch_cluster_health_task_max_waiting_in_queue_millis        | gauge      | 1           | Max time in millis that a task is waiting in queue.                                                 |
| elasticsearch_cluster_health_relocating_shards                       | gauge      | 1           | The number of shards that are currently moving from one node to another node.                       |
| elasticsearch_cluster_health_status                                  | gauge      | 3           | Whether all primary and replica shards are allocated.                                               |
| elasticsearch_cluster_health_unassigned_shards                       | gauge      | 1           | The number of shards that exist in the cluster state, but cannot be found in the cluster itself.    |
| elasticsearch_clustersettings_stats_max_shards_per_node              | gauge      | 0           | Current maximum number of shards per node setting.                                                  |
| elasticsearch_clustersettings_allocation_threshold_enabled           | gauge      | 0           | Is disk allocation decider enabled.                                                                 |
| elasticsearch_clustersettings_allocation_watermark_flood_stage_bytes | gauge      | 0           | Flood stage watermark as in bytes.                                                                  |
| elasticsearch_clustersettings_allocation_watermark_high_bytes        | gauge      | 0           | High watermark for disk usage in bytes.                                                             |
| elasticsearch_clustersettings_allocation_watermark_low_bytes         | gauge      | 0           | Low watermark for disk usage in bytes.                                                              |
| elasticsearch_clustersettings_allocation_watermark_flood_stage_ratio | gauge      | 0           | Flood stage watermark as a ratio.                                                                   |
| elasticsearch_clustersettings_allocation_watermark_high_ratio        | gauge      | 0           | High watermark for disk usage as a ratio.                                                           |
| elasticsearch_clustersettings_allocation_watermark_low_ratio         | gauge      | 0           | Low watermark for disk usage as a ratio.                                                            |
| elasticsearch_filesystem_data_available_bytes                        | gauge      | 1           | Available space on block device in bytes                                                            |
| elasticsearch_filesystem_data_free_bytes                             | gauge      | 1           | Free space on block device in bytes                                                                 |
| elasticsearch_filesystem_data_size_bytes                             | gauge      | 1           | Size of block device in bytes                                                                       |
| elasticsearch_filesystem_io_stats_device_operations_count            | gauge      | 1           | Count of disk operations                                                                            |
| elasticsearch_filesystem_io_stats_device_read_operations_count       | gauge      | 1           | Count of disk read operations                                                                       |
| elasticsearch_filesystem_io_stats_device_write_operations_count      | gauge      | 1           | Count of disk write operations                                                                      |
| elasticsearch_filesystem_io_stats_device_read_size_kilobytes_sum     | gauge      | 1           | Total kilobytes read from disk                                                                      |
| elasticsearch_filesystem_io_stats_device_write_size_kilobytes_sum    | gauge      | 1           | Total kilobytes written to disk                                                                     |
| elasticsearch_ilm_status                                             | gauge      | 1           | Current status of ILM. Status can be `STOPPED`, `RUNNING`, `STOPPING`.                              |
| elasticsearch_ilm_index_status                                       | gauge      | 4           | Status of ILM policy for index                                                                      |
| elasticsearch_indices_active_queries                                 | gauge      | 1           | The number of currently active queries                                                              |
| elasticsearch_indices_docs                                           | gauge      | 1           | Count of documents on this node                                                                     |
| elasticsearch_indices_docs_deleted                                   | gauge      | 1           | Count of deleted documents on this node                                                             |
| elasticsearch_indices_deleted_docs_primary                           | gauge      | 1           | Count of deleted documents with only primary shards                                                 |
| elasticsearch_indices_docs_primary                                   | gauge      | 1           | Count of documents with only primary shards on all nodes                                            |
| elasticsearch_indices_docs_total                                     | gauge      |             | Count of documents with shards on all nodes                                                         |
| elasticsearch_indices_fielddata_evictions                            | counter    | 1           | Evictions from field data                                                                           |
| elasticsearch_indices_fielddata_memory_size_bytes                    | gauge      | 1           | Field data cache memory usage in bytes                                                              |
| elasticsearch_indices_filter_cache_evictions                         | counter    | 1           | Evictions from filter cache                                                                         |
| elasticsearch_indices_filter_cache_memory_size_bytes                 | gauge      | 1           | Filter cache memory usage in bytes                                                                  |
| elasticsearch_indices_flush_time_seconds                             | counter    | 1           | Cumulative flush time in seconds                                                                    |
| elasticsearch_indices_flush_total                                    | counter    | 1           | Total flushes                                                                                       |
| elasticsearch_indices_get_exists_time_seconds                        | counter    | 1           | Total time get exists in seconds                                                                    |
| elasticsearch_indices_get_exists_total                               | counter    | 1           | Total get exists operations                                                                         |
| elasticsearch_indices_get_missing_time_seconds                       | counter    | 1           | Total time of get missing in seconds                                                                |
| elasticsearch_indices_get_missing_total                              | counter    | 1           | Total get missing                                                                                   |
| elasticsearch_indices_get_time_seconds                               | counter    | 1           | Total get time in seconds                                                                           |
| elasticsearch_indices_get_total                                      | counter    | 1           | Total get                                                                                           |
| elasticsearch_indices_indexing_delete_time_seconds_total             | counter    | 1           | Total time indexing delete in seconds                                                               |
| elasticsearch_indices_indexing_delete_total                          | counter    | 1           | Total indexing deletes                                                                              |
| elasticsearch_indices_index_current                                  | gauge      | 1           | The number of documents currently being indexed to an index                                         |
| elasticsearch_indices_indexing_index_time_seconds_total              | counter    | 1           | Cumulative index time in seconds                                                                    |
| elasticsearch_indices_indexing_index_total                           | counter    | 1           | Total index calls                                                                                   |
| elasticsearch_indices_mappings_stats_fields                          | gauge      | 1           | Count of fields currently mapped by index                                                           |
| elasticsearch_indices_mappings_stats_json_parse_failures_total       | counter    | 0           | Number of errors while parsing JSON                                                                 |
| elasticsearch_indices_mappings_stats_scrapes_total                   | counter    | 0           | Current total Elasticsearch Indices Mappings scrapes                                                |
| elasticsearch_indices_mappings_stats_up                              | gauge      | 0           | Was the last scrape of the Elasticsearch Indices Mappings endpoint successful                       |
| elasticsearch_indices_merges_docs_total                              | counter    | 1           | Cumulative docs merged                                                                              |
| elasticsearch_indices_merges_total                                   | counter    | 1           | Total merges                                                                                        |
| elasticsearch_indices_merges_total_size_bytes_total                  | counter    | 1           | Total merge size in bytes                                                                           |
| elasticsearch_indices_merges_total_time_seconds_total                | counter    | 1           | Total time spent merging in seconds                                                                 |
| elasticsearch_indices_query_cache_cache_total                        | counter    | 1           | Count of query cache                                                                                |
| elasticsearch_indices_query_cache_cache_size                         | gauge      | 1           | Size of query cache                                                                                 |
| elasticsearch_indices_query_cache_count                              | counter    | 2           | Count of query cache hit/miss                                                                       |
| elasticsearch_indices_query_cache_evictions                          | counter    | 1           | Evictions from query cache                                                                          |
| elasticsearch_indices_query_cache_memory_size_bytes                  | gauge      | 1           | Query cache memory usage in bytes                                                                   |
| elasticsearch_indices_query_cache_total                              | counter    | 1           | Size of query cache total                                                                           |
| elasticsearch_indices_refresh_time_seconds_total                     | counter    | 1           | Total time spent refreshing in seconds                                                              |
| elasticsearch_indices_refresh_total                                  | counter    | 1           | Total refreshes                                                                                     |
| elasticsearch_indices_request_cache_count                            | counter    | 2           | Count of request cache hit/miss                                                                     |
| elasticsearch_indices_request_cache_evictions                        | counter    | 1           | Evictions from request cache                                                                        |
| elasticsearch_indices_request_cache_memory_size_bytes                | gauge      | 1           | Request cache memory usage in bytes                                                                 |
| elasticsearch_indices_search_fetch_time_seconds                      | counter    | 1           | Total search fetch time in seconds                                                                  |
| elasticsearch_indices_search_fetch_total                             | counter    | 1           | Total number of fetches                                                                             |
| elasticsearch_indices_search_query_time_seconds                      | counter    | 1           | Total search query time in seconds                                                                  |
| elasticsearch_indices_search_query_total                             | counter    | 1           | Total number of queries                                                                             |
| elasticsearch_indices_segments_count                                 | gauge      | 1           | Count of index segments on this node                                                                |
| elasticsearch_indices_segments_memory_bytes                          | gauge      | 1           | Current memory size of segments in bytes                                                            |
| elasticsearch_indices_settings_creation_timestamp_seconds            | gauge      | 1           | Timestamp of the index creation in seconds                                                                     |
| elasticsearch_indices_settings_stats_read_only_indices               | gauge      | 1           | Count of indices that have read_only_allow_delete=true                                              |
| elasticsearch_indices_settings_total_fields                          | gauge      |             | Index setting value for index.mapping.total_fields.limit (total allowable mapped fields in a index) |
| elasticsearch_indices_settings_replicas                              | gauge      |             | Index setting value for index.replicas                                                              |
| elasticsearch_indices_shards_docs                                    | gauge      | 3           | Count of documents on this shard                                                                    |
| elasticsearch_indices_shards_docs_deleted                            | gauge      | 3           | Count of deleted documents on each shard                                                            |
| elasticsearch_indices_store_size_bytes                               | gauge      | 1           | Current size of stored index data in bytes                                                          |
| elasticsearch_indices_store_size_bytes_primary                       | gauge      |             | Current size of stored index data in bytes with only primary shards on all nodes                    |
| elasticsearch_indices_store_size_bytes_total                         | gauge      |             | Current size of stored index data in bytes with all shards on all nodes                             |
| elasticsearch_indices_store_throttle_time_seconds_total              | counter    | 1           | Throttle time for index store in seconds                                                            |
| elasticsearch_indices_translog_operations                            | counter    | 1           | Total translog operations                                                                           |
| elasticsearch_indices_translog_size_in_bytes                         | counter    | 1           | Total translog size in bytes                                                                        |
| elasticsearch_indices_warmer_time_seconds_total                      | counter    | 1           | Total warmer time in seconds                                                                        |
| elasticsearch_indices_warmer_total                                   | counter    | 1           | Total warmer count                                                                                  |
| elasticsearch_jvm_gc_collection_seconds_count                        | counter    | 2           | Count of JVM GC runs                                                                                |
| elasticsearch_jvm_gc_collection_seconds_sum                          | counter    | 2           | GC run time in seconds                                                                              |
| elasticsearch_jvm_memory_committed_bytes                             | gauge      | 2           | JVM memory currently committed by area                                                              |
| elasticsearch_jvm_memory_max_bytes                                   | gauge      | 1           | JVM memory max                                                                                      |
| elasticsearch_jvm_memory_used_bytes                                  | gauge      | 2           | JVM memory currently used by area                                                                   |
| elasticsearch_jvm_memory_pool_used_bytes                             | gauge      | 3           | JVM memory currently used by pool                                                                   |
| elasticsearch_jvm_memory_pool_max_bytes                              | counter    | 3           | JVM memory max by pool                                                                              |
| elasticsearch_jvm_memory_pool_peak_used_bytes                        | counter    | 3           | JVM memory peak used by pool                                                                        |
| elasticsearch_jvm_memory_pool_peak_max_bytes                         | counter    | 3           | JVM memory peak max by pool                                                                         |
| elasticsearch_os_cpu_percent                                         | gauge      | 1           | Percent CPU used by the OS                                                                          |
| elasticsearch_os_load1                                               | gauge      | 1           | Shortterm load average                                                                              |
| elasticsearch_os_load5                                               | gauge      | 1           | Midterm load average                                                                                |
| elasticsearch_os_load15                                              | gauge      | 1           | Longterm load average                                                                               |
| elasticsearch_process_cpu_percent                                    | gauge      | 1           | Percent CPU used by process                                                                         |
| elasticsearch_process_cpu_seconds_total                              | counter    | 1           | Process CPU time in seconds                                                                         |
| elasticsearch_process_mem_resident_size_bytes                        | gauge      | 1           | Resident memory in use by process in bytes                                                          |
| elasticsearch_process_mem_share_size_bytes                           | gauge      | 1           | Shared memory in use by process in bytes                                                            |
| elasticsearch_process_mem_virtual_size_bytes                         | gauge      | 1           | Total virtual memory used in bytes                                                                  |
| elasticsearch_process_open_files_count                               | gauge      | 1           | Open file descriptors                                                                               |
| elasticsearch_snapshot_stats_number_of_snapshots                     | gauge      | 1           | Total number of snapshots                                                                           |
| elasticsearch_snapshot_stats_oldest_snapshot_timestamp               | gauge      | 1           | Oldest snapshot timestamp                                                                           |
| elasticsearch_snapshot_stats_snapshot_start_time_timestamp           | gauge      | 1           | Last snapshot start timestamp                                                                       |
| elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds       | gauge      | 1           | Timestamp of the latest SUCCESS or PARTIAL snapshot                                                 |
| elasticsearch_snapshot_stats_snapshot_end_time_timestamp             | gauge      | 1           | Last snapshot end timestamp                                                                         |
| elasticsearch_snapshot_stats_snapshot_number_of_failures             | gauge      | 1           | Last snapshot number of failures                                                                    |
| elasticsearch_snapshot_stats_snapshot_number_of_indices              | gauge      | 1           | Last snapshot number of indices                                                                     |
| elasticsearch_snapshot_stats_snapshot_failed_shards                  | gauge      | 1           | Last snapshot failed shards                                                                         |
| elasticsearch_snapshot_stats_snapshot_successful_shards              | gauge      | 1           | Last snapshot successful shards                                                                     |
| elasticsearch_snapshot_stats_snapshot_total_shards                   | gauge      | 1           | Last snapshot total shard                                                                           |
| elasticsearch_thread_pool_active_count                               | gauge      | 14          | Thread Pool threads active                                                                          |
| elasticsearch_thread_pool_completed_count                            | counter    | 14          | Thread Pool operations completed                                                                    |
| elasticsearch_thread_pool_largest_count                              | gauge      | 14          | Thread Pool largest threads count                                                                   |
| elasticsearch_thread_pool_queue_count                                | gauge      | 14          | Thread Pool operations queued                                                                       |
| elasticsearch_thread_pool_rejected_count                             | counter    | 14          | Thread Pool operations rejected                                                                     |
| elasticsearch_thread_pool_threads_count                              | gauge      | 14          | Thread Pool current threads count                                                                   |
| elasticsearch_transport_rx_packets_total                             | counter    | 1           | Count of packets received                                                                           |
| elasticsearch_transport_rx_size_bytes_total                          | counter    | 1           | Total number of bytes received                                                                      |
| elasticsearch_transport_tx_packets_total                             | counter    | 1           | Count of packets sent                                                                               |
| elasticsearch_transport_tx_size_bytes_total                          | counter    | 1           | Total number of bytes sent                                                                          |
| elasticsearch_clusterinfo_last_retrieval_success_ts                  | gauge      | 1           | Timestamp of the last successful cluster info retrieval                                             |
| elasticsearch_clusterinfo_up                                         | gauge      | 1           | Up metric for the cluster info collector                                                            |
| elasticsearch_clusterinfo_version_info                               | gauge      | 6           | Constant metric with ES version information as labels                                               |
| elasticsearch_slm_stats_up                                           | gauge      | 0           | Up metric for SLM collector                                                                         |
| elasticsearch_slm_stats_total_scrapes                                | counter    | 0           | Number of scrapes for SLM collector                                                                 |
| elasticsearch_slm_stats_json_parse_failures                          | counter    | 0           | JSON parse failures for SLM collector                                                               |
| elasticsearch_slm_stats_retention_runs_total                         | counter    | 0           | Total retention runs                                                                                |
| elasticsearch_slm_stats_retention_failed_total                       | counter    | 0           | Total failed retention runs                                                                         |
| elasticsearch_slm_stats_retention_timed_out_total                    | counter    | 0           | Total retention run timeouts                                                                        |
| elasticsearch_slm_stats_retention_deletion_time_seconds              | gauge      | 0           | Retention run deletion time                                                                         |
| elasticsearch_slm_stats_total_snapshots_taken_total                  | counter    | 0           | Total snapshots taken                                                                               |
| elasticsearch_slm_stats_total_snapshots_failed_total                 | counter    | 0           | Total snapshots failed                                                                              |
| elasticsearch_slm_stats_total_snapshots_deleted_total                | counter    | 0           | Total snapshots deleted                                                                             |
| elasticsearch_slm_stats_total_snapshots_failed_total                 | counter    | 0           | Total snapshots failed                                                                              |
| elasticsearch_slm_stats_snapshots_taken_total                        | counter    | 1           | Snapshots taken by policy                                                                           |
| elasticsearch_slm_stats_snapshots_failed_total                       | counter    | 1           | Snapshots failed by policy                                                                          |
| elasticsearch_slm_stats_snapshots_deleted_total                      | counter    | 1           | Snapshots deleted by policy                                                                         |
| elasticsearch_slm_stats_snapshot_deletion_failures_total             | counter    | 1           | Snapshot deletion failures by policy                                                                |
| elasticsearch_slm_stats_operation_mode                               | gauge      | 1           | SLM operation mode (Running, stopping, stopped)                                                     |
| elasticsearch_data_stream_stats_up                                   | gauge      | 0           | Up metric for Data Stream collection                                                                |
| elasticsearch_data_stream_stats_total_scrapes                        | counter    | 0           | Total scrapes for Data Stream stats                                                                 |
| elasticsearch_data_stream_stats_json_parse_failures                  | counter    | 0           | Number of parsing failures for Data Stream stats                                                    |
| elasticsearch_data_stream_backing_indices_total                      | gauge      | 1           | Number of backing indices for Data Stream                                                           |
| elasticsearch_data_stream_store_size_bytes                           | gauge      | 1           | Current size of data stream backing indices in bytes                                                |
| elasticsearch_health_report_creating_primaries                       | gauge      | 1           | The number of creating primary shards                                                               |
| elasticsearch_health_report_creating_replicas                        | gauge      | 1           | The number of creating replica shards                                                               |
| elasticsearch_health_report_data_stream_lifecycle_status             | gauge      | 2           | Data stream lifecycle status                                                                        |
| elasticsearch_health_report_disk_status                              | gauge      | 2           | disk status                                                                                         |
| elasticsearch_health_report_ilm_policies                             | gauge      | 1           | The number of ILM Policies                                                                          |
| elasticsearch_health_report_ilm_stagnating_indices                   | gauge      | 1           | The number of stagnating indices                                                                    |
| elasticsearch_health_report_ilm_status                               | gauge      | 2           | ILM status                                                                                          |
| elasticsearch_health_report_initializing_primaries                   | gauge      | 1           | The number of initializing primary shards                                                           |
| elasticsearch_health_report_initializing_replicas                    | gauge      | 1           | The number of initializing replica shards                                                           |
| elasticsearch_health_report_master_is_stable_status                  | gauge      | 2           | Master is stable status                                                                             |
| elasticsearch_health_report_max_shards_in_cluster_data               | gauge      | 1           | The number of maximum shards in a cluster                                                           |
| elasticsearch_health_report_max_shards_in_cluster_frozen             | gauge      | 1           | The number of maximum frozen shards in a cluster                                                    |
| elasticsearch_health_report_repository_integrity_status              | gauge      | 2           | Repository integrity status                                                                         |
| elasticsearch_health_report_restarting_primaries                     | gauge      | 1           | The number of restarting primary shards                                                             |
| elasticsearch_health_report_restarting_replicas                      | gauge      | 1           | The number of restarting replica shards                                                             |
| elasticsearch_health_report_shards_availabilty_status                | gauge      | 2           | Shards availabilty status                                                                           |
| elasticsearch_health_report_shards_capacity_status                   | gauge      | 2           | Shards capacity status                                                                              |
| elasticsearch_health_report_slm_policies                             | gauge      | 1           | The number of SLM policies                                                                          |
| elasticsearch_health_report_slm_status                               | gauge      | 2           | SLM status                                                                                          |
| elasticsearch_health_report_started_primaries                        | gauge      | 1           | The number of started primary shards                                                                |
| elasticsearch_health_report_started_replicas                         | gauge      | 1           | The number of started replica shards                                                                |
| elasticsearch_health_report_status                                   | gauge      | 2           | Overall cluster status                                                                              |
| elasticsearch_health_report_total_repositories                       | gauge      | 1           | The number snapshot repositories                                                                    |
| elasticsearch_health_report_unassigned_primaries                     | gauge      | 1           | The number of unassigned primary shards                                                             |
| elasticsearch_health_report_unassigned_replicas                      | gauge      | 1           | The number of unassigned replica shards                                                             |

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
