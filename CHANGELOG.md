## master / unreleased

BREAKING CHANGES:

The flag `--es.slm` has been renamed to `--collector.slm`.

The logging system has been replaced with log/slog from the stdlib. This change is being made across the prometheus ecosystem. The logging output has changed, but the messages and levels remain the same. The `ts` label for the timestamp has bewen replaced with `time`, the accuracy is less, and the timezone is not forced to UTC. The `caller` field has been replaced by the `source` field, which now includes the full path to the source file. The `level` field now exposes the log level in capital letters.

* [CHANGE] Rename --es.slm to --collector.slm #932
* [CHANGE] Replace logging system #942

## 1.8.0 / 2024-09-14

* [FEATURE] Add tasks action collector. Enable using `--collector.tasks.actions`. #778
* [FEATURE] Add additional nodes metrics for indexing pressure monitoring. #904

## 1.7.0 / 2023-12-02

BREAKING CHANGES:

The flag `--es.snapshots` has been renamed to `--collector.snapshots`.

* [CHANGE] Rename --es.snapshots to --collector.snapshots #789
* [CHANGE] Add cluster label to `elasticsearch_node_shards_total` metric #639
* [FEATURE] Add watermark metrics #611
* [FEATURE] Add `elasticsearch_indices_settings_creation_timestamp_seconds` metric #816

## 1.6.0 / 2023-06-22

BREAKING CHANGES:

The flag `--es.cluster_settings` has been renamed to `--collector.clustersettings`.

* [CHANGE] Rename --es.cluster_settings to --collector.clustersettings
* [FEATURE] Add ILM metrics #513
* [ENHANCEMENT] Add ElasticCloud node roles to role label #652
* [ENHANCEMENT] Add ability to use AWS IAM role for authentication #653
* [ENHANCEMENT] Add metric for index replica count #483
* [BUGFIX] Set elasticsearch_clusterinfo_version_info guage to 1 #728
* [BUGFIX] Fix index field counts with nested fields #675


## 1.5.0 / 2022-07-28

* [FEATURE] Add metrics collection for data stream statistics #592
* [FEATURE] Support for AWS Elasticsearch using AWS SDK v2 #597
* [BUGFIX] Fix cluster settings collection when max_shards_per_node is manually set. #603

## 1.4.0 / 2022-06-29

* [BREAKING] Remove ENV var support for most non-sensitive options. #518
* [BREAKING] Rename elasticsearch_process_cpu_time_seconds_sum to elasticsearch_process_cpu_seconds_total #520
* [FEATURE] Add metric for index aliases #563
* [FEATURE] Add metric for number of shards on a node #535
* [FEATURE] Add metrics for SLM (snapshot lifecycle management) #558
* [FEATURE] Add metric for JVM uptime #537
* [FEATURE] Add metrics for current searches and current indexing documents #485
* [BUGFIX] Remove the elasticsearch_process_cpu_time_seconds_sum metric as it was never used #498

## 1.3.0 / 2021-10-21

* [FEATURE] Add support for passing elasticsearch credentials via the ES_USERNAME and ES_PASSWORD environment varialbes #461
* [FEATURE] Add support for API keys for elasticsearch authentication (Elastic cloud) #459
* [BUGFIX] Fix index stats when shards are unavailable #445

## 1.2.1 / 2021-06-29

* [BUGFIX] Fixed elasticsearch 7.13 node stats metrics #439
* [BUGFIX] Fixed snapshot stats metrics for some snapshot repository types #442

## 1.2.0 / 2021-06-10

This release marks the first release under the prometheus-community organization.

* [FEATURE] Added elasticsearch_clustersettings_stats_max_shards_per_node metric. #277
* [FEATURE] Added elasticsearch_indices_shards_store_size_in_bytes metric. #292
* [FEATURE] Added --es.indices_mappings flag to scrape elasticsearch index mapping stats and elasticsearch_indices_mappings_stats collector. #411
* [FEATURE] Added elasticsearch_snapshot_stats_latest_snapshot_timestamp_seconds metric. #318
* [ENHANCEMENT] Added support for reloading the tls client certificate in case it changes on disk. #414
* [BUGFIX] Fixed the elasticsearch_indices_shards_docs metric name. #291

## 1.1.0

repeating the breaking changes introduced in 1.1.0rc1:
* [BREAKING] uses the registered exporter port 9114 instead of 9118. If you need to stick to the old port, you can specify the listen port with --web.listen-address
* [BREAKING] commandline flags are now POSIX flags with double dashes --

new changes in 1.1.0:
* [FEATURE] add checksum promu command to Makefile
* [FEATURE] add healthz handler
* [BUGFIX] json parse error if the snapshot json contains failures (#269)
* [BUGFIX] Remove credentials from URL in clusterinfo metrics
* [FEATURE] Add indices_segment_term_vectors_memory_bytes_{primary,total} metrics
* [FEATURE] Add indices_segments_{points,term_vectors,version_map}_memory_in_bytes metrics
* [BUGFIX] Kubernetes yml file fixes
* [FEATURE] Add index_stats_query_cache_caches_total metric
* [FEATURE] Rename query_cache_cache_count metric to query_cache_cache_total
* [BUGFIX] Change type for indices_query_cache_cache_count metric to counter
* [BUGFIX]/ [BREAKING] Add _total prefix to indices_warmer_time_seconds metric
* [FEATURE] Add indices_warmer_{time_seconds,total} metrics
* [BUGFIX] exporter doesn't exit 1 if port is already in use (#241)
* [BUGFIX] parse clusterinfo.build_date as string, not time.Time
* [BUGFIX] Various Documentation Fixes
* [FEATURE] add node_roles metric (#207)
* [FEATURE] Extend nodes metrics. added indices.merges.current_size
build fix: remove unnecessary conversion
* [FEATURE] Extend nodes metrics. added overhead of circuit breakers
* [BUGFIX] fix nodes metrics name indices.query_cache_miss_count, indices.request_cache_miss_count
* [FEATURE] Extend nodes search metrics. added scroll_total, scroll_time
* [FEATURE] Extend indices.indexing nodes metrics. added is_throttled, throttle_time
* [FEATURE]/ [BUGFIX] #212 remove misleading metric

## 1.1.0rc1

* [BREAKING] uses the registered exporter port 9114 instead of 9118. If you need to stick to the old port, you can specify the listen port with --web.listen-address
* [BREAKING] commandline flags are now POSIX flags with double dashes --
* [FEATURE] new collector for snapshot metrics
* [FEATURE] added os memory stats metrics
* [FEATURE] enable querying ES via proxy
* [FEATURE] new collector for cluster settings
* [FEATURE] new collector for indices settings
* [FEATURE] cluster info collector. The collector periodically queries the / endpoints and provides the other collectors with a semi up-to-date cluster label
*
* [FEATURE]/ [BUGFIX] grafana dashboard improvements and fixes
* [BUGFIX] Fixed createTLSConfig function. Return full tls configuration when ca, crt, key and insecure flag are set
*
* [INTERNAL] added code linting to build pipeline

## 1.0.4rc1

* [DOCUMENTATION] documentation updates
* [FEATURE] add more index metrics
* [FEATURE] add filesystem metrics
* [FEATURE] add jvm buffer pool metrics
* [FEATURE] add support for using the exporter behind reverse proxy (URL-prefixing)
* [ENHANCEMENT] add linting to build chain and make project lint clean

## 1.0.3rc1

* [BUGFIX] update prometheus alerting rule example to v2 format
* [ENHANCEMENT] Add formatting option for logger
* [ENHANCEMENT] Add shard-level document count (can be toggled out)
* [ENHANCEMENT] Add OS CPU usage metric
* [ENHANCEMENT] Add skip-ssl-verify option
* [ENHANCEMENT] Add node-level current merge metrics

## 1.0.2 / 2018-01-09

* [ENHANCEMENT] Add index metrics [#85] [#116] [#118]
* [ENHANCEMENT] Add cache metrics [#88]
* [ENHANCEMENT] Add documentation for the example dashboard [#84]
* [ENHANCEMNET] Expose load averages [#113]
* [BUGFIX] Fix role detection [#105] [#110]
* [BUGFIX] Fix indexing calls and time metrics [#83]

## 1.0.1 / 2017-07-24

* [ENHANCEMENT] Add exporter instrumentation [#78]
* [BUGFIX] Exclude basic auth credentials from log [#71]
* [BUGFIX] Fix missing node store size metric

## 1.0.0 / 2017-07-03

* [ENHANCEMENT] Rewrite the codebase to reduce redundancy and improve extensibility [#65]
* [ENHANCEMENT] Add examples for Grafana and Prometheus [#66]
* [BREAKING] Removed several duplicate or redundant metrics [#65]
