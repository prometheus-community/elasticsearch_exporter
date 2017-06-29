# calculate filesytem used and free percent
elasticsearch_filesystem_data_used_percent = 100 * (elasticsearch_filesystem_data_size_bytes - elasticsearch_filesystem_data_free_bytes) / elasticsearch_filesystem_data_size_bytes
elasticsearch_filesystem_data_free_percent = 100 - elasticsearch_filesystem_data_used_percent

# alert if too few nodes are running
ALERT ElasticsearchTooFewNodesRunning
  IF elasticsearch_cluster_health_number_of_node < 3
  FOR 5m
  LABELS {severity="critical"}
  ANNOTATIONS {description="There are only {{$value}} < 3 ElasticSearch nodes running", summary="ElasticSearch running on less than 3 nodes"}

# alert if heap usage is over 90%
ALERT ElasticsearchHeapTooHigh
  IF elasticsearch_jvm_memory_used_bytes{area="heap"} / elasticsearch_jvm_memory_max_bytes{area="heap"} > 0.9
  FOR 15m
  LABELS {severity="critical"}
  ANNOTATIONS {description="The heap usage is over 90% for 15m", summary="ElasticSearch node {{$labels.node}} heap usage is high"}
