local g = import 'g.libsonnet';

local dashboard = g.dashboard;
local row = g.panel.row;

local panels = import './panels.libsonnet';
local queries = import './queries.libsonnet';
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
      + dashboard.withPanels(
        g.util.grid.makeGrid([
          row.new('Overview')
          + row.withPanels([
            panels.stat.nodes('Nodes', queries.runningNodes),
            panels.stat.nodes('Data Nodes', queries.dataNodes),
            panels.stat.nodes('Pending Tasks', queries.pendingTasks),
          ]),
          row.new('Shards')
          + row.withPanels([
            panels.stat.nodes('Active', queries.activeShards),
            panels.stat.nodes('Active Primary', queries.activePrimaryShards),
            panels.stat.nodes('Initializing', queries.initializingShards),
            panels.stat.nodes('Relocating', queries.reloactingShards),
            panels.stat.nodes('Unassigned', queries.unassignedShards),
            panels.stat.nodes('DelayedUnassigned', queries.delayedUnassignedShards),
          ]),
          row.new('Documents')
          + row.withPanels([
            panels.timeSeries.base('Indexed Documents', queries.indexedDocuments),

          ]),
        ], panelWidth=4, panelHeight=3),
      ),
  },
}
