#!/bin/bash

json=$(cat -)
instance_id=$(echo $json | jq -r ".instance_id")

backup_dir=/tmp/dingo-postgresql-broker-demo-backups/
mkdir -p ${backup_dir}

echo $json > ${backup_dir}/${instance_id}.json
