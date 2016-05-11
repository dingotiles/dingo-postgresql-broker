#!/bin/bash

set -e
set -x

instance_id=$1
if [[ "${instance_id}X" == "X" ]]; then
  echo "USAGE: $0 instance-id"
  exit 1
fi

curl -v -u starkandwayne:starkandwayne "localhost:3000/v2/service_instances/${instance_id}" \
    -d '{}' \
    -X PUT -H "X-Broker-API-Version: 2.8" -H "Content-Type: application/json"
