local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;

local variables = import '../variables.libsonnet';

{
  memoryUsage:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        elasticsearch_jvm_memory_used_bytes{cluster=~"$cluster"}
      |||
    )
    + prometheusQuery.withLegendFormat('{{name}} {{area}}'),

  memoryUsageAverage15:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        avg_over_time(
          elasticsearch_jvm_memory_used_bytes{cluster=~"$cluster"}[15m]
        ) /
        elasticsearch_jvm_memory_max_bytes{cluster=~"$cluster"}
      |||
    )
    + prometheusQuery.withLegendFormat('{{name}} {{area}}'),

  memoryMax:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        elasticsearch_jvm_memory_max_bytes{cluster=~"$cluster"}
      |||
    )
    + prometheusQuery.withLegendFormat('{{name}} {{area}}'),

  gcSeconds:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rate(
          elasticsearch_jvm_gc_collection_seconds_sum{cluster=~"$cluster"}[$__rate_interval]
        )
      |||
    )
    + prometheusQuery.withLegendFormat('{{name}} {{gc}}'),
}
