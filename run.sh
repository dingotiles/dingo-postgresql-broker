#!/bin/bash

export PATRONI_BROKER_CONFIG=${PATRONI_BROKER_CONFIG:-"demo/bosh-lite.example.yml"}

go run main.go broker
