#!/bin/bash

set -e -x

export GOPATH=$PWD/broker
export PATH=$GOPATH/bin:$PATH

cd ${GOPATH}/src/github.com/dingotiles/patroni-broker
go install github.com/onsi/ginkgo/ginkgo
ginkgo -r "$@"
