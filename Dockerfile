FROM golang:alpine

RUN apk update && apk add --update alpine-sdk

ADD .  $GOPATH/src/elasticsearch_exporter

RUN \
    cd $GOPATH/src/elasticsearch_exporter && \
    go build && \
    go install


EXPOSE      9108
ENTRYPOINT  [ "elasticsearch_exporter" ]
