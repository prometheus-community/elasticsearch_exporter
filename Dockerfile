ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:glibc

ADD .   /go/src/github.com/prometheus-community/elasticsearch_exporter
WORKDIR /go/src/github.com/prometheus-community/elasticsearch_exporter

RUN go mod download
RUN make 

FROM scratch as scratch

COPY --from=builder /go/src/github.com/prometheus-community/elasticsearch_exporter/elasticsearch_exporter  /bin/elasticsearch_exporter

EXPOSE      9114
ENTRYPOINT  [ "/bin/elasticsearch_exporter" ]

FROM quay.io/sysdig/sysdig-mini-ubi:1.2.12 as ubi

COPY --from=builder /go/src/github.com/prometheus-community/elasticsearch_exporter/elasticsearch_exporter  /bin/elasticsearch_exporter

EXPOSE      9114
ENTRYPOINT  [ "/bin/elasticsearch_exporter" ]
