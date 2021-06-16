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
              summary: 'Some shards will be re-allocated to different nodes if possible.',
              description: '[{{ $labels.cluster }}][{{ $labels.node }}] has disk usage of {{ $value | humanizePercentage }}. Add more disk space, or free up space.',
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
              summary: 'Some shards will be re-allocated to different nodes if possible.',
              description: '[{{ $labels.cluster }}][{{ $labels.node }}] has disk usage of {{ $value | humanizePercentage }}. Add more disk space, or free up space.',
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
              description: "Cluster {{ $labels.cluster }} health status has been RED for at least %(esClusterHealthStatusRED)s. Cluster does not accept writes, shards may be missing or master node hasn't been elected yet." % custom.alert,
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
              description: 'Cluster {{ $labels.cluster }} health status has been YELLOW for at least %(esClusterHealthStatusYELLOW)s. Some shard replicas are not allocated.' % custom.alert,
            },
          },
          {
            alert: 'ElasticsearchThreadPoolRejectionError',
            expr: |||
              max by (cluster, name, type) (irate(elasticsearch_thread_pool_rejected_count{%(selector)s}[%(esClusterThreadpoolEvalTime)s])) > %(esClusterThreadpoolErrorThreshold)s
            ||| % custom.alert,
            'for': '%(esClusterThreadpoolWarningTime)s' % custom.alert,
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: '[Elasticsearch] High rejection rate for {{ $labels.type }} threadpool',
              description: '[{{ $labels.cluster }}][{{ $labels.name }}] threadpool rejection over %(esClusterThreadpoolWarningTime)s > %(esClusterThreadpoolErrorThreshold)s' % custom.alert,
            },
          },
          {
            alert: 'ElasticsearchThreadPoolRejectionError',
            expr: |||
              max by (cluster, name, type) (irate(elasticsearch_thread_pool_rejected_count{%(selector)s}[%(esClusterThreadpoolEvalTime)s])) > %(esClusterThreadpoolErrorThreshold)s
            ||| % custom.alert,
            'for': '%(esClusterThreadpoolCriticalTime)s' % custom.alert,
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: '[Elasticsearch] High rejection rate for {{ $labels.type }} threadpool',
              description: '[{{ $labels.cluster }}][{{ $labels.name }}] threadpool rejection over %(esClusterThreadpoolCriticalTime)s > %(esClusterThreadpoolErrorThreshold)s' % custom.alert,
            },
          },
          {
            alert: 'ElasticsearchSnapshotFailure',
            expr: |||
              elasticsearch_snapshot_stats_snapshot_number_of_failures{%(selector)s} >  0
            ||| % custom.alert,
            'for': '%(esClusterSnapshotFailureTime)s' % custom.alert,
            labels: {
              severity: 'critical',
            },
            annotations: {
              summary: '[Elasticsearch] Snapshot failure',
              description: '[{{ $labels.cluster }}] Last snapshot failed.',
            },
          },
          {
            alert: 'ElasticsearchSnapshotMetricsUnavailable',
            expr: |||
              elasticsearch_snapshot_stats_up{%(selector)s} ==  0
            ||| % custom.alert,
            'for': '%(esClusterSnapshotStatsUpTime)s' % custom.alert,
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: '[Elasticsearch] Snapshot metrics API is broken',
              description: '[{{ $labels.cluster }}] No data for cluster backup status.',
            },
          },
        ],
      },
    ],
  },
}
