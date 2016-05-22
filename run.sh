#!/bin/bash

export PATRONI_BROKER_CONFIG=demo/bosh-lite.yml

go run main.go broker
