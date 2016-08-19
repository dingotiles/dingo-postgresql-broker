#!/bin/bash

json=$(cat -)
from_uri=$(echo $json | jq -r ".from_uri")
to_uri=$(echo $json | jq -r ".to_uri")

echo "{\"msg\":\"Failed to copy from $from_uri to $to_uri\"}"
exit 1
