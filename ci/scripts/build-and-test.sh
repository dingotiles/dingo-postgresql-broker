#!/bin/bash

set -e -x

export GOPATH=$PWD/broker:$PWD/broker/src/github.com/dingotiles/patroni-broker/Godeps/_workspace
export PATH=$GOPATH/bin:$PATH

cd ${GOPATH}/src/github.com/dingotiles/patroni-broker
go install github.com/onsi/ginkgo/ginkgo
ginkgo -r "$@"
