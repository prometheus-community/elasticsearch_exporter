FROM alpine
RUN set -x \
    && apk --update upgrade \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*
COPY elasticsearch_exporter  /bin/elasticsearch_exporter
USER nobody:nobody

EXPOSE      9108
ENTRYPOINT  ["/bin/elasticsearch_exporter"]
