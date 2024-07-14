local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;

local variables = import '../variables.libsonnet';

{
  transportTXRate:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rate(
          elasticsearch_transport_rx_size_bytes_total{cluster=~"$cluster"}[$__rate_interval]
        )
      |||
    )
    + prometheusQuery.withLegendFormat('{{name}} TX'),

  transportRXRate:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rate(
          elasticsearch_transport_tx_size_bytes_total{cluster=~"$cluster"}[$__rate_interval]
        )
      |||
    )
    + prometheusQuery.withLegendFormat('{{name}} RX'),
}
