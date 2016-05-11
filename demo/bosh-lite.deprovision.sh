#!/bin/bash

set -e
set -x

instance_id=$1
if [[ "${instance_id}X" == "X" ]]; then
  echo "USAGE: $0 instance-id"
  exit 1
fi

service_id=$(curl -s -u starkandwayne:starkandwayne localhost:3000/v2/catalog | jq -r ".services[0].id")
plan_id=$(curl -s -u starkandwayne:starkandwayne localhost:3000/v2/catalog | jq -r ".services[0].plans[0].id")

curl -v -u starkandwayne:starkandwayne "localhost:3000/v2/service_instances/${instance_id}?service_id=${service_id}&plan_id=${plan_id}" \
   -X DELETE -H "X-Broker-API-Version: 2.8" -H "Content-Type: application/json"
