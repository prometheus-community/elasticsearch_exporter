FROM quay.io/prometheus/golang-builder AS builder

ARG PROMU_VERSION=0.13.0
ADD  https://github.com/prometheus/promu/releases/download/v${PROMU_VERSION}/promu-${PROMU_VERSION}.linux-amd64.tar.gz ./
RUN tar -xvzf promu-${PROMU_VERSION}.linux-amd64.tar.gz && mv promu-${PROMU_VERSION}.linux-amd64/promu /go/bin

ADD .   /go/src/github.com/prometheus-community/elasticsearch_exporter
WORKDIR /go/src/github.com/prometheus-community/elasticsearch_exporter

RUN go mod download
RUN make 

FROM scratch AS scratch

COPY --from=builder /go/src/github.com/prometheus-community/elasticsearch_exporter/elasticsearch_exporter  /bin/elasticsearch_exporter

EXPOSE      9114
ENTRYPOINT  [ "/bin/elasticsearch_exporter" ]

FROM quay.io/sysdig/sysdig-mini-ubi9:1.3.2 AS ubi

COPY --from=builder /go/src/github.com/prometheus-community/elasticsearch_exporter/elasticsearch_exporter  /bin/elasticsearch_exporter

EXPOSE      9114
ENTRYPOINT  [ "/bin/elasticsearch_exporter" ]
