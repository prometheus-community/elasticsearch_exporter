{
  local custom = self,
  alert+:: {
    selector: error 'must provide selector for Elasticsearch alerts',
  },
  prometheusRules+:: {
    groups+: [
      {
        name: 'elasticsearch.application.rules',
        rules: [
          {
            record: ':elasticsearch_threadpool_utilisation:ratio',
            expr: |||
              elasticsearch_thread_pool_active_count{%(selector)s}
              /
              elasticsearch_thread_pool_threads_count{%(selector)s} >= 0
            ||| % custom.alert,
          },
          {
            // Default configuration, is write threads to match the number of CPU cores available
            // We record the ratio of threadpool usage against available CPU in the cluster.
            // hot-warm-cold architecture may need some additional labels for this to be useful.
            record: ':elasticsearch_cluster_threadpool_utilisation:ratio',
            expr: |||
              sum without (cluster, host, instance, type, es_master_node, es_data_node, es_ingest_node, es_client_node) (
                elasticsearch_thread_pool_active_count{%(selector)s}
              )
              /
              sum without (cluster, host, instance, type, es_master_node, es_data_node, es_ingest_node, es_client_node) (
                elasticsearch_thread_pool_threads_count{type="write", %(selector)s}
              )
            ||| % custom.alert,
          },
        ],
      },
    ],
  },
}
