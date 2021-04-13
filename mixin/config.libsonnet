{
  alert+:: {
    selector: 'job=~"elasticsearch.*"',
    esDiskLowWaterMark: 0.85,
    esDiskHighWaterMark: 0.9,
    esClusterHealthStatusRED: '2m',
    esClusterHealthStatusYELLOW: '20m',
    esClusterThreadpoolErrorThreshold: 0,
    esClusterThreadpoolEvalTime: '5m',
    esClusterThreadpoolWarningTime: '2m',
    esClusterThreadpoolCriticalTime: '10m',
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
