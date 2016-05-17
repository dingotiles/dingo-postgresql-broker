#!/bin/bash

set -e

export GOPATH=$PWD/broker:$GOPATH
cd broker/src/github.com/dingotiles/dingo-postgresql-broker
scripts/test.sh
