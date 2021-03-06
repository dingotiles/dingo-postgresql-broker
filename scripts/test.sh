#!/usr/bin/env bash

if [[ ! -x "$(command -v etcd)" ]]; then
    echo >&2 "etcd is not installed. Please install etcd to run tests"
    exit 2
fi

# Create a temp dir and clean it up on exit
TEMPDIR=`mktemp -d -t broker-test.XXX`
etcd --data-dir ${TEMPDIR} > /dev/null 2>&1 &
etcd_pid=$!
trap "rm -rf $TEMPDIR && kill ${etcd_pid}" EXIT HUP INT QUIT TERM

# Run the tests
if [[ "${1}X" == "X" ]]; then
  echo "--> Running all tests"
  go list ./... | grep -v '/vendor/' | xargs -n1 go test -cover -timeout=360s
else
  go test -cover -timeout=360s $@
fi
