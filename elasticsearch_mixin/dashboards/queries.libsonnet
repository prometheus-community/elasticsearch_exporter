local g = import './g.libsonnet';
local prometheusQuery = g.query.prometheus;

local variables = import './variables.libsonnet';

{
  /*
    General overview queries
  */

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

  /*
    Shard queries
  */

  activeShards:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          elasticsearch_cluster_health_active_shards{cluster=~"$cluster"}
        )
      |||
    ),

  activePrimaryShards:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          elasticsearch_cluster_health_active_primary_shards{cluster=~"$cluster"}
        )
      |||
    ),

  initializingShards:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          elasticsearch_cluster_health_initializing_shards{cluster=~"$cluster"}
        )
      |||
    ),

  reloactingShards:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          elasticsearch_cluster_health_reloacting_shards{cluster=~"$cluster"}
        )
      |||
    ),

  unassignedShards:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          elasticsearch_cluster_health_unassigned_shards{cluster=~"$cluster"}
        )
      |||
    ),

  delayedUnassignedShards:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(
          elasticsearch_cluster_health_delayed_unassigned_shards{cluster=~"$cluster"}
        )
      |||
    ),

  /*
    Documents queries
  */

  indexedDocuments:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        elasticsearch_indices_docs{cluster=~"$cluster"}
      |||
    ),
}
