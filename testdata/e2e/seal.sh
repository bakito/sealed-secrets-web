#!/bin/bash
set -e
./toSecret.sh ${1} \
  | curl -s --show-error -H "Content-Type: application/json" -X POST --data-binary @- http://localhost/ssw/api/seal \
  | jq -r .secret > ${2}
