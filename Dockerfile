FROM alpine:latest

ENV GOPATH /go
ENV APPPATH $GOPATH/src/github.com/ewr/elasticsearch_exporter

COPY . $APPPATH

RUN apk add --update -t build-deps go git mercurial libc-dev gcc libgcc \
    && cd $APPPATH && go get -d && go build -o /elasticsearch_exporter \
    && apk del --purge build-deps && rm -rf $GOPATH

EXPOSE 9108

ENTRYPOINT ["/elasticsearch_exporter"]
CMD ["-es.uri=http://localhost:9200", "--es.all=true", "--web.listen-address=:9108", "--web.telemetry-path=/metrics"]
