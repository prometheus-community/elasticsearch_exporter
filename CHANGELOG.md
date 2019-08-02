## 1.0.3rc1

- [BUGFIX] update prometheus alerting rule example to v2 format
- [ENHANCEMENT] Add formatting option for logger
- [ENHANCEMENT] Add shard-level document count (can be toggled out)
- [ENHANCEMENT] Add OS CPU usage metric
- [ENHANCEMENT] Add skip-ssl-verify option
- [ENHANCEMENT] Add node-level current merge metrics

## 1.0.2 / 2018-01-09

- [ENHANCEMENT] Add index metrics [#85][#116] [#118]
- [ENHANCEMENT] Add cache metrics [#88]
- [ENHANCEMENT] Add documentation for the example dashboard [#84]
- [ENHANCEMNET] Expose load averages [#113]
- [BUGFIX] Fix role detection [#105][#110]
- [BUGFIX] Fix indexing calls and time metrics [#83]

## 1.0.1 / 2017-07-24

- [ENHANCEMENT] Add exporter instrumentation [#78]
- [BUGFIX] Exclude basic auth credentials from log [#71]
- [BUGFIX] Fix missing node store size metric

## 1.0.0 / 2017-07-03

- [ENHANCEMENT] Rewrite the codebase to reduce redundancy and improve
  extensibility [#65]
- [ENHANCEMENT] Add examples for Grafana and Prometheus [#66]
- [BREAKING] Removed several duplicate or redundant metrics [#65]
