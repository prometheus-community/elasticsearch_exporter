{
  alert+:: {
    selector: 'job=~"elasticsearch.*"',
    esDiskLowWaterMark: 0.85,
    esDiskHighWaterMark: 0.9,
    esClusterHealthStatusRED: '2m',
    esClusterHealthStatusYELLOW: '20m',
  },
  rule+:: {
    selector: 'job=~"elasticsearch.*"',
  },
  dashboard+:: {
    selector: 'job=~"elasticsearch.*"',
  },
}
