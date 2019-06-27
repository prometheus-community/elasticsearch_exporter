ARG target
FROM $target/alpine
LABEL maintainer="Jesse Stuart <hi@jessestuart.com>"

ARG arch
ENV ARCH=$arch

COPY elasticsearch_exporter /bin/elasticsearch_exporter
COPY qemu-* /usr/bin/

EXPOSE      9114
ENTRYPOINT  [ "/bin/elasticsearch_exporter" ]
