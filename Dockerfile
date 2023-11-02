ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:glibc

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/elasticsearch_exporter /bin/elasticsearch_exporter

EXPOSE      7979
USER        nobody
ENTRYPOINT  [ "/bin/elasticsearch_exporter" ]
