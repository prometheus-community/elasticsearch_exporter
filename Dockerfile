ARG target
FROM $target/alpine

ARG arch
ENV ARCH=$arch

COPY qemu-$ARCH-static* /usr/bin/

LABEL maintainer="Jesse Stuart <hi@jessestuart.com>"

COPY elasticsearch_exporter /bin/elasticsearch_exporter

EXPOSE      9114
ENTRYPOINT  [ "/bin/elasticsearch_exporter" ]
