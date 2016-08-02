#!/bin/bash

json=$(cat -)
service_instance_name=$(echo $json | jq -r ".service_instance_name")

backup_dir=/tmp/dingo-postgresql-broker-demo-backups/
mkdir -p ${backup_dir}

grep $service_instance_name ${backup_dir}/*.json
