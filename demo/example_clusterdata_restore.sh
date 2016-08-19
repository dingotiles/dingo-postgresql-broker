#!/bin/bash

instance_id=$(cat -)

backup_dir=/tmp/dingo-postgresql-broker-demo-backups/
mkdir -p ${backup_dir}

cat ${backup_dir}/${instance_id}.json
