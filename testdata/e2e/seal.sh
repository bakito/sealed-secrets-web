#!/bin/bash
set -e
./toSecret.sh ${1} \
  | curl -s --show-error -H "Content-Type: application/json" -X POST --data-binary @- localhost:8080/api/seal \
  | jq -r .secret > ${2}
