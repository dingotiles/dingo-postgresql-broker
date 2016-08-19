#!/bin/bash

json=$(cat -)
service_instance_name=$(echo $json | jq -r ".name")
space_guid=$(echo $json | jq -r ".space_guid")

backup_dir=/tmp/dingo-postgresql-broker-demo-backups/
mkdir -p ${backup_dir}

jq -r "select(.space_guid == \"${space_guid}\" and .service_instance_name == \"${service_instance_name}\")" ${backup_dir}/*.json
