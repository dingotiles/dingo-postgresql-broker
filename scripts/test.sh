#!/usr/bin/env bash

if [[ ! -x "$(command -v etcd)" ]]; then
    echo >&2 "etcd is not installed. Please install etcd to run tests"
    exit 2
fi

# Create a temp dir and clean it up on exit
TEMPDIR=`mktemp -d -t broker-test.XXX`
trap "rm -rf $TEMPDIR" EXIT HUP INT QUIT TERM

etcd --data-dir ${TEMPDIR} > /dev/null 2>&1 &
etcd_pid=$!
trap "kill ${etcd_pid}" EXIT HUP INT QUIT TERM

# Run the tests
echo "--> Running tests"
go list ./... | grep -v '/vendor/' | xargs -n1 go test -cover -timeout=360s
