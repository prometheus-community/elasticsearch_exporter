local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;

local variables = import '../variables.libsonnet';

{
  runningNodes:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          elasticsearch_cluster_health_number_of_nodes{cluster=~"$cluster"}
        )
      |||
    ),
  dataNodes:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          elasticsearch_cluster_health_number_of_data_nodes{cluster=~"$cluster"}
        )
      |||
    ),

  pendingTasks:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          elasticsearch_cluster_health_number_of_pending_tasks{cluster=~"$cluster"}
        )
      |||
    ),
}
