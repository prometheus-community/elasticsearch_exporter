#!/bin/bash

MIXIN_PATH=./elasticsearch-mixin
MIXIN_OUT_PATH=./elasticsearch-mixin/compiled

rm -rf ${MIXIN_OUT_PATH} && mkdir ${MIXIN_OUT_PATH}
pushd ${MIXIN_PATH} && jb install && popd
mixtool generate all --output-alerts ${MIXIN_OUT_PATH}/alerts.yaml --output-rules ${MIXIN_OUT_PATH}/rules.yaml --directory ${MIXIN_OUT_PATH}/dashboards ${MIXIN_PATH}/mixin.libsonnet
