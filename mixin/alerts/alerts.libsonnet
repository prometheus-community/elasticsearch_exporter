{
  // Using max() for alerts, because running multiple exporters in cluster would distort values
  local custom = self,
  alert+:: {
    selector: error 'must provide selector for Elasticsearch alerts',
  },
  prometheusAlerts+:: {
    groups+: [

      {
        name: 'elasticsearch.system.alerts',
        rules: [
          {
            alert: 'ElasticsearchNodeDiskWatermarkReached',
            expr: |||
              max by (cluster, instance, node) (
                1 - (elasticsearch_filesystem_data_free_bytes{%(selector)s} / elasticsearch_filesystem_data_size_bytes{%(selector)s})
              ) > %(esDiskLowWaterMark)s
            ||| % custom.alert,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'Disk Low Watermark Reached - disk saturation is {{ $value | humanizePercentage }}%',
              message: 'Disk Low Watermark Reached at {{ $labels.node }} node in {{ $labels.cluster }} cluster. Shards can not be allocated to this node anymore. You should consider adding more disk to the node.',
            },
          },
          {
            alert: 'ElasticsearchNodeDiskWatermarkReached',
            expr: |||
              max by (cluster, instance, node) (
                1 - (elasticsearch_filesystem_data_free_bytes{%(selector)s} / elasticsearch_filesystem_data_size_bytes{%(selector)s})
              ) > %(esDiskHighWaterMark)s
            ||| % custom.alert,
            'for': '5m',
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: 'Disk High Watermark Reached - disk saturation is {{ $value | humanizePercentage }}%',
              message: 'Disk High Watermark Reached at {{ $labels.node }} node in {{ $labels.cluster }} cluster. Some shards will be re-allocated to different nodes if possible. Make sure more disk space is added to the node or drop old indices allocated to this node.',
            },
          },
        ],
      },
      {
        name: 'elasticsearch.application.alerts',
        rules: [
          {
            alert: 'ElasticsearchClusterStatusRed',
            expr: |||
              max by (cluster) (elasticsearch_cluster_health_status{color="red", %(selector)s} == 1)
            ||| % custom.alert,
            'for': '%(esClusterHealthStatusRED)s' % custom.alert,
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: 'Cluster health status is RED',
              message: "Cluster {{ $labels.cluster }} health status has been RED for at least %(esClusterHealthStatusRED)s. Cluster does not accept writes, shards may be missing or master node hasn't been elected yet." % custom.alert,
            },
          },
          {
            alert: 'ElasticsearchClusterStatusYellow',
            expr: |||
              max by (cluster) (elasticsearch_cluster_health_status{color="yellow", %(selector)s} == 1)
            ||| % custom.alert,
            'for': '%(esClusterHealthStatusYELLOW)s' % custom.alert,
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'Cluster health status is YELLOW',
              message: 'Cluster {{ $labels.cluster }} health status has been YELLOW for at least %(esClusterHealthStatusYELLOW)s. Some shard replicas are not allocated.' % custom.alert,
            },
          },
          //{
          //  alert: 'ElasticsearchBulkRequestsRejectionJumps',
          //  expr: |||
          //    round( bulk:reject_ratio:rate2m * 100, 0.001 ) > %(esBulkPctIncrease)s
          //  ||| % $._config,
          //  'for': '10m',
          //  labels: {
          //    severity: 'warning',
          //  },
          //  annotations: {
          //    summary: 'High Bulk Rejection Ratio - {{ $value }}%',
          //    message: 'High Bulk Rejection Ratio at {{ $labels.node }} node in {{ $labels.cluster }} cluster. This node may not be keeping up with the indexing speed.',
          //  },
          //},
        ],
      },


    ],
  },
}
