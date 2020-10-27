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
          {
            alert: 'ElasticsearchThreadPoolRejectionError',
            expr: |||
              sum without (host, name, instance, es_master_node, es_data_node, es_ingest_node, es_client_node)(
                increase(elasticsearch_thread_pool_rejected_count{%(selector)s}[%(esClusterThreadpoolEvalTime)s])
              ) >  %(esClusterThreadpoolErrorThreshold)s
            ||| % custom.alert,
            'for': '%(esClusterThreadpoolWarningTime)s' % custom.alert,
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: '[Elasticsearch] High rejection rate for {{ $labels.type }} threadpool',
              message: '[{{ $labels.cluster }}] threadpool rejection over %(esClusterThreadpoolWarningTime)s > %(esClusterThreadpoolErrorThreshold)s' % custom.alert,
            },
          },
          {
            alert: 'ElasticsearchThreadPoolRejectionError',
            expr: |||
              sum without (host, name, instance, es_master_node, es_data_node, es_ingest_node, es_client_node)(
                increase(elasticsearch_thread_pool_rejected_count{%(selector)s}[%(esClusterThreadpoolEvalTime)s])
              ) >  %(esClusterThreadpoolErrorThreshold)s
            ||| % custom.alert,
            'for': '%(esClusterThreadpoolCriticalTime)s' % custom.alert,
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: '[Elasticsearch] High rejection rate for {{ $labels.type }} threadpool',
              message: '[{{ $labels.cluster }}] threadpool rejection over %(esClusterThreadpoolCriticalTime)s > %(esClusterThreadpoolErrorThreshold)s' % custom.alert,
            },
          },
        ],
      },
    ],
  },
}
