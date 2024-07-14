local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;

local variables = import '../variables.libsonnet';

{
  threadPoolActive:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        elasticsearch_thread_pool_active_count{cluster=~"$cluster"}
      |||
    )
    + prometheusQuery.withLegendFormat('{{type}}'),

  threadPoolRejections:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        elasticsearch_thread_pool_rejected_count{cluster=~"$cluster"}
      |||
    )
    + prometheusQuery.withLegendFormat('{{name}} {{type}}'),
}
