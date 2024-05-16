local g = import 'g.libsonnet';

local dashboard = g.dashboard;

local variables = import './variables.libsonnet';

{
  grafanaDashboards+:: {
    'cluster.json':
      dashboard.new('%s Cluster' % $._config.dashboardNamePrefix)
      + dashboard.withTags($._config.dashboardTags)
      + dashboard.withRefresh('1m')
      + dashboard.time.withFrom(value='now-1h')
      + dashboard.graphTooltip.withSharedCrosshair()
      + dashboard.withVariables([
          variables.datasource,
          variables.cluster,
      ])
  }
}
