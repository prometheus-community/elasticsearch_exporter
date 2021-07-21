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
            // Active threads / CPU cores
            // elasticsearch_thread_pool_threads_count{type="write} - to get the CPU cores
            record: ':elasticsearch_cluster_threadpool_utilisation:ratio',
            expr: |||
              sum without (host, name, instance, type) (
                elasticsearch_thread_pool_active_count{%(selector)s}
              )
              /
              sum without (host, name, instance, type) (
                elasticsearch_thread_pool_threads_count{type="write", %(selector)s}
              ) >= 0
            ||| % custom.alert,
          },
        ],
      },
    ],
  },
}
