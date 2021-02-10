local legacyDashboard = (import 'legacy.libsonnet');

{
  grafanaDashboards+:: {
    'legacy_elasticsearch.json': legacyDashboard,
  },
}
