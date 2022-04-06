FROM quay.io/prometheus/golang-builder as builder

ARG PROMU_VERSION=0.13.0
ADD  https://github.com/prometheus/promu/releases/download/v${PROMU_VERSION}/promu-${PROMU_VERSION}.linux-amd64.tar.gz ./
RUN tar -xvzf promu-${PROMU_VERSION}.linux-amd64.tar.gz && mv promu-${PROMU_VERSION}.linux-amd64/promu /go/bin

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
