ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:glibc
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/elasticsearch_exporter /bin/elasticsearch_exporter


RUN go mod download github.com/alecthomas/units
RUN make

FROM scratch

COPY --from=builder /go/src/github.com/justwatchcom/elasticsearch_exporter/elasticsearch_exporter  /bin/elasticsearch_exporter

EXPOSE      7979
USER        nobody

ENTRYPOINT  [ "/bin/elasticsearch_exporter" ]
