{
  alert+:: {
    selector: 'job=~"elasticsearch.*"',
    esDiskLowWaterMark: 0.85,
    esDiskHighWaterMark: 0.9,
    esClusterHealthStatusRED: '2m',
    esClusterHealthStatusYELLOW: '20m',
    esClusterThreadpoolErrorThreshold: 20,
    esClusterThreadpoolEvalTime: '5m',
    esClusterThreadpoolWarningTime: '10m',
    esClusterThreadpoolCriticalTime: '30m',
    esClusterSnapshotFailureTime: '8h',
    esClusterSnapshotStatsUpTime: '1h',
  },
  rule+:: {
    selector: 'job=~"elasticsearch.*"',
  },
  dashboard+:: {
    selector: 'job=~"elasticsearch.*"',
  },
}
