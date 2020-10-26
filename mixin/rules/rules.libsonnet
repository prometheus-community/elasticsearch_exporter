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
            // Default configuration, is write threads to match the number of CPU cores available
            // We record the ratio of threadpool usage against available CPU in the cluster.
            // hot-warm-cold architecture may need some additional labels for this to be useful.
            record: ':elasticsearch_threadpool_utilisation:sum_rate',
            expr: |||
              sum by (cluster, type) (
                sum without (host, name) (
                  elasticsearch_thread_pool_active_count{%(selector)s}
                )
              )
              /
              sum by (cluster, type) (
                sum without (host, name) (
                  elasticsearch_thread_pool_threads_count{%(selector)s}
                )
              )
            ||| % custom.alert,
          },
        ],
      },
    ],
  },
}
