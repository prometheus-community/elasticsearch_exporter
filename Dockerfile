FROM        quay.io/prometheus/busybox:latest
MAINTAINER  The Prometheus Authors <prometheus-developers@googlegroups.com>

COPY elasticsearch_exporter  /bin/elasticsearch_exporter

EXPOSE      9108
ENTRYPOINT  [ "/bin/elasticsearch_exporter" ]
