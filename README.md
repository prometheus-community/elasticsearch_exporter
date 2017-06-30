# Elasticsearch Exporter [![Build Status](https://travis-ci.org/justwatchcom/elasticsearch_exporter.svg?branch=master)](https://travis-ci.org/justwatchcom/elasticsearch_exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/justwatch/elasticsearch_exporter.svg?maxAge=604800)](https://hub.docker.com/r/justwatch/elasticsearch_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/justwatchcom/elasticsearch_exporter)](https://goreportcard.com/report/github.com/justwatchcom/elasticsearch_exporter)

Prometheus exporter for various metrics about ElasticSearch, written in Go.

### Installation

For pre-built binaries please take a look at the releases.  
https://github.com/justwatchcom/elasticsearch_exporter/releases

#### Docker

```bash
docker pull justwatch/elasticsearch_exporter:1.0.0-rc1
docker run --rm -p 9108:9108 justwatch/elasticsearch_exporter:1.0.0-rc1
```

Example `docker-compose.yml`:

```yaml
elasticsearch_exporter:
    image: justwatch/elasticsearch_exporter:1.0.0-rc1
    command:
     - '-es.uri=http://elasticsearch:9200'
    restart: always
    ports:
    - "127.0.0.1:9108:9108"
```

### Configuration

```bash
elasticsearch_exporter --help
```

| Argument              | Description |
| --------              | ----------- |
| es.uri                | Address (host and port) of the Elasticsearch node we should connect to. This could be a local node (`localhost:9200`, for instance), or the address of a remote Elasticsearch server.
| es.all                | If true, query stats for all nodes in the cluster, rather than just the node we connect to.
| es.timeout            | Timeout for trying to get stats from Elasticsearch. (ex: 20s) |
| es.ca                 | Path to PEM file that contains trusted CAs for the Elasticsearch connection.
| es.client-private-key | Path to PEM file that contains the private key for client auth when connecting to Elasticsearch.
| es.client-cert        | Path to PEM file that contains the corresponding cert for the private key to connect to Elasticsearch.
| web.listen-address    | Address to listen on for web interface and telemetry. |
| web.telemetry-path    | Path under which to expose metrics. |

### Metrics

|Name                                                        |Type       |Cardinality   |Help
|----                                                        |----       |-----------   |----
| elasticsearch_breakers_estimated_size_bytes                | gauge     | 4            | Estimated size in bytes of breaker
| elasticsearch_breakers_limit_size_bytes                    | gauge     | 4            | Limit size in bytes for breaker
| elasticsearch_breakers_tripped                             | gauge     | 4            | tripped for breaker
| elasticsearch_cluster_health_active_primary_shards         | gauge     | 1            | The number of primary shards in your cluster. This is an aggregate total across all indices.
| elasticsearch_cluster_health_active_shards                 | gauge     | 1            | Aggregate total of all shards across all indices, which includes replica shards.
| elasticsearch_cluster_health_delayed_unassigned_shards     | gauge     | 1            | Shards delayed to reduce reallocation overhead
| elasticsearch_cluster_health_initializing_shards           | gauge     | 1            | Count of shards that are being freshly created.
| elasticsearch_cluster_health_number_of_data_nodes          | gauge     | 1            | Number of data nodes in the cluster.
| elasticsearch_cluster_health_number_of_in_flight_fetch     | gauge     | 1            | The number of ongoing shard info requests.
| elasticsearch_cluster_health_number_of_nodes               | gauge     | 1            | Number of nodes in the cluster.
| elasticsearch_cluster_health_number_of_pending_tasks       | gauge     | 1            | Cluster level changes which have not yet been executed
| elasticsearch_cluster_health_relocating_shards             | gauge     | 1            | The number of shards that are currently moving from one node to another node.
| elasticsearch_cluster_health_status                        | gauge     | 3            | Whether all primary and replica shards are allocated.
| elasticsearch_cluster_health_timed_out                     | gauge     | 1            | Number of cluster health checks timed out
| elasticsearch_cluster_health_unassigned_shards             | gauge     | 1            | The number of shards that exist in the cluster state, but cannot be found in the cluster itself.
| elasticsearch_filesystem_data_available_bytes              | gauge     | 1            | Available space on block device in bytes
| elasticsearch_filesystem_data_free_bytes                   | gauge     | 1            | Free space on block device in bytes
| elasticsearch_filesystem_data_size_bytes                   | gauge     | 1            | Size of block device in bytes
| elasticsearch_indices_docs                                 | gauge     | 1            | Count of documents on this node
| elasticsearch_indices_docs_deleted                         | gauge     | 1            | Count of deleted documents on this node
| elasticsearch_indices_fielddata_evictions                  | counter   | 1            | Evictions from field data
| elasticsearch_indices_fielddata_memory_size_bytes          | gauge     | 1            | Field data cache memory usage in bytes
| elasticsearch_indices_filter_cache_evictions               | counter   | 1            | Evictions from filter cache
| elasticsearch_indices_filter_cache_memory_size_bytes       | gauge     | 1            | Filter cache memory usage in bytes
| elasticsearch_indices_flush_time_seconds                   | counter   | 1            | Cumulative flush time in seconds
| elasticsearch_indices_flush_total                          | counter   | 1            | Total flushes
| elasticsearch_indices_get_exists_time_seconds              | counter   | 1            | Total time get exists in seconds
| elasticsearch_indices_get_exists_total                     | counter   | 1            | Total get exists operations
| elasticsearch_indices_get_missing_time_seconds             | counter   | 1            | Total time of get missing in seconds
| elasticsearch_indices_get_missing_total                    | counter   | 1            | Total get missing
| elasticsearch_indices_get_time_seconds                     | counter   | 1            | Total get time in seconds
| elasticsearch_indices_get_total                            | counter   | 1            | Total get
| elasticsearch_indices_indexing_delete_time_seconds_total   | counter   | 1            | Total time indexing delete in seconds
| elasticsearch_indices_indexing_delete_total                | counter   | 1            | Total indexing deletes
| elasticsearch_indices_indexing_index_time_seconds_total    | counter   | 1            | Total index calls
| elasticsearch_indices_indexing_index_total                 | counter   | 1            | Cumulative index time in seconds
| elasticsearch_indices_merges_docs_total                    | counter   | 1            | Cumulative docs merged
| elasticsearch_indices_merges_total                         | counter   | 1            | Total merges
| elasticsearch_indices_merges_total_size_bytes_total        | counter   | 1            | Total merge size in bytes
| elasticsearch_indices_merges_total_time_seconds_total      | counter   | 1            | Total time spent merging in seconds
| elasticsearch_indices_query_cache_evictions                | counter   | 1            | Evictions from query cache
| elasticsearch_indices_query_cache_memory_size_bytes        | gauge     | 1            | Query cache memory usage in bytes
| elasticsearch_indices_refresh_time_seconds_total           | counter   | 1            | Total refreshes
| elasticsearch_indices_refresh_total                        | counter   | 1            | Total time spent refreshing in seconds
| elasticsearch_indices_request_cache_evictions              | counter   | 1            | Evictions from request cache
| elasticsearch_indices_request_cache_memory_size_bytes      | gauge     | 1            | Request cache memory usage in bytes
| elasticsearch_indices_search_fetch_time_seconds            | counter   | 1            | Total search fetch time in seconds
| elasticsearch_indices_search_fetch_total                   | counter   | 1            | Total number of fetches
| elasticsearch_indices_search_query_time_seconds            | counter   | 1            | Total search query time in seconds
| elasticsearch_indices_search_query_total                   | counter   | 1            | Total number of queries
| elasticsearch_indices_segments_count                       | gauge     | 1            | Count of index segments on this node
| elasticsearch_indices_segments_memory_bytes                | gauge     | 1            | Current memory size of segments in bytes
| elasticsearch_indices_store_size_bytes                     | gauge     | 1            | Current size of stored index data in bytes
| elasticsearch_indices_store_throttle_time_seconds_total    | counter   | 1            | Throttle time for index store in seconds
| elasticsearch_indices_translog_operations                  | counter   | 1            | Total translog operations
| elasticsearch_indices_translog_size_in_bytes               | counter   | 1            | Total translog size in bytes
| elasticsearch_jvm_gc_collection_seconds_count              | counter   | 2            | Count of JVM GC runs
| elasticsearch_jvm_gc_collection_seconds_sum                | counter   | 2            | GC run time in seconds
| elasticsearch_jvm_memory_committed_bytes                   | gauge     | 2            | JVM memory currently committed by area
| elasticsearch_jvm_memory_max_bytes                         | gauge     | 1            | JVM memory max
| elasticsearch_jvm_memory_used_bytes                        | gauge     | 2            | JVM memory currently used by area
| elasticsearch_process_cpu_percent                          | gauge     | 1            | Percent CPU used by process
| elasticsearch_process_cpu_time_seconds_sum                 | counter   | 3            | Process CPU time in seconds
| elasticsearch_process_mem_resident_size_bytes              | gauge     | 1            | Resident memory in use by process in bytes
| elasticsearch_process_mem_share_size_bytes                 | gauge     | 1            | Shared memory in use by process in bytes
| elasticsearch_process_mem_virtual_size_bytes               | gauge     | 1            | Total virtual memory used in bytes
| elasticsearch_process_open_files_count                     | gauge     | 1            | Open file descriptors
| elasticsearch_thread_pool_active_count                     | gauge     | 14           | Thread Pool threads active
| elasticsearch_thread_pool_completed_count                  | counter   | 14           | Thread Pool operations completed
| elasticsearch_thread_pool_largest_count                    | gauge     | 14           | Thread Pool largest threads count
| elasticsearch_thread_pool_queue_count                      | gauge     | 14           | Thread Pool operations queued
| elasticsearch_thread_pool_rejected_count                   | counter   | 14           | Thread Pool operations rejected
| elasticsearch_thread_pool_threads_count                    | gauge     | 14           | Thread Pool current threads count
| elasticsearch_transport_rx_packets_total                   | counter   | 1            | Count of packets received
| elasticsearch_transport_rx_size_bytes_total                | counter   | 1            | Total number of bytes received
| elasticsearch_transport_tx_packets_total                   | counter   | 1            | Count of packets sent
| elasticsearch_transport_tx_size_bytes_total                | counter   | 1            | Total number of bytes sent

### Alerts & Recording Rules

We provide examples for [Prometheus](http://prometheus.io) [alerts and recording rules](examples/prometheus/elasticsearch.rules) as well as an [Grafana](http://www.grafana.org) [Dashboard](examples/grafana/dashboard.json) and a [Kubernetes](http://kubernetes.io) [Deployment](examples/kubernetes/deployment.yml).

## Credit & License

`elasticsearch_exporter` is maintained by the nice folks from [JustWatch](https://www.justwatch.com/)
and licensed under the terms of the Apache license.

This package was originally created and mainted by [Eric Richardson](https://github.com/ewr),
who transferred this repository to us in January 2017.

Maintainers of this repository:

* Matthias Loibl <matthias.loibl@justwatch.com> @metalmatze
* Dominik Schulz <dominik.schulz@justwatch.com> @dominikschulz

Please refer to the Git commit log for a complete list of contributors.

## Contributing

We welcome any contributions. Please fork the project on GitHub and open
Pull Requests for any proposed changes.

Please note that we will not merge any changes that encourage insecure
behaviour. If in doubt please open an Issue first to discuss your proposal.
