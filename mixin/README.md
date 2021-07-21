# Monitoring Mixins for Elasticsearch Exporter

## Recording Rules

| name | description |
|---|---|
| :elasticsearch_threadpool_utilisation:ratio | Threadpool utilisation for a service node |
| :elasticsearch_cluster_threadpool_utilisation:ratio | Threadpool utilisation for the cluster |


## Alerts

| alertname | severity |
|---|---|
| ElasticsearchNodeDiskWatermarkReached | warning |
| ElasticsearchNodeDiskWatermarkReached | critical |
| ElasticsearchClusterStatusYellow | warning |
| ElasticsearchClusterStatusRed | critical |
| ElasticsearchThreadPoolRejectionError | warning |
| ElasticsearchThreadPoolRejectionError | critical |
| ElasticsearchSnapshotFailure | warning |
| ElasticsearchSnapshotMetricsUnavailable | warning |

## Resources

**Quickstart:**

    jsonnet lib/dashboards.jsonnet
    jsonnet -S lib/rules.jsonnet
    jsonnet -S lib/alerts.jsonnet

[Read the docs](https://github.com/monitoring-mixins/docs)
