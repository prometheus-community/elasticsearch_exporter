FROM        alpine:3.4
MAINTAINER  The Prometheus Authors <prometheus-developers@googlegroups.com>

COPY . /go/src/github.com/justwatchcom/elasticsearch_exporter

WORKDIR /go/src/github.com/justwatchcom/elasticsearch_exporter

RUN apk --update add ca-certificates \
 && apk --update add --virtual build-deps go git \
 && GOPATH=/go go get \
 && GOPATH=/go go build -o /bin/elasticsearch_exporter \
 && apk del --purge build-deps \
 && rm -rf /go/bin /go/pkg /var/cache/apk/*

EXPOSE      9108
ENTRYPOINT  [ "/bin/elasticsearch_exporter" ]
