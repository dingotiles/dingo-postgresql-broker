#!/bin/bash

if [[ ! -f tmp/cf.yml ]]; then
  echo "Place your CF admin credentials into tmp/cf.yml"
  exit 1
fi
export PATRONI_BROKER_CONFIG=tmp/bosh-lite.yml
spruce merge config/bosh-lite.example.yml tmp/cf.yml > $PATRONI_BROKER_CONFIG

go run main.go broker
