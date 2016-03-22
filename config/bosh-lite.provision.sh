#!/bin/bash

set -e

# service_id=$(curl -s -u starkandwayne:starkandwayne localhost:3000/v2/catalog | jq -r ".services[0].id")
# plan_id=$(curl -s -u starkandwayne:starkandwayne localhost:3000/v2/catalog | jq -r ".services[0].plans[0].id")
instance_id=$(uuid)

set -x
curl -v -u starkandwayne:starkandwayne localhost:3000/v2/service_instances/${instance_id} \
   -X PUT -H "X-Broker-API-Version: 2.8" -H "Content-Type: application/json" \
   -d '{"service_id": "beb5973c-e1b2-11e5-a736-c7c0b526363d", "plan_id": "b96d0936-e423-11e5-accb-93d374e93368", "organization_guid": "unknown", "space_guid": "unknown", "parameters": {"node-count": 2}}'
set +x

echo To deprovision:
echo   ./config/bosh-lite.deprovision.sh $instance_id
