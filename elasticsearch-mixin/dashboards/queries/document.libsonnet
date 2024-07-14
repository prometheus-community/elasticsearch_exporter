local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;

local variables = import '../variables.libsonnet';

{
  indexedDocuments:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        elasticsearch_indices_docs{cluster=~"$cluster"}
      |||
    ),

  indexSize:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        elasticsearch_indices_store_size_bytes{cluster=~"$cluster"}
      |||
    ),

  indexRate:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rate(elasticsearch_indices_indexing_index_total{cluster=~"$cluster"}[$__rate_interval])
      |||
    )
    + prometheusQuery.withLegendFormat('{{name}}'),

  queryRate:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rate(elasticsearch_indices_search_query_total{cluster=~"$cluster"}[$__rate_interval])
      |||
    )
    + prometheusQuery.withLegendFormat('{{name}}'),

  queueCount:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(elasticsearch_thread_pool_queue_count{cluster=~"$cluster",type!="management"}) by (type)
      |||
    )
    + prometheusQuery.withLegendFormat('{{type}}'),

}
