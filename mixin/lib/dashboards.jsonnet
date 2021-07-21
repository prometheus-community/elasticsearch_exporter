local elastic = (import '../mixin.libsonnet');

elastic.grafanaDashboards['legacy_elasticsearch.json'] {
  timezone: 'utc',
  tags+: ['mixin'],
}
