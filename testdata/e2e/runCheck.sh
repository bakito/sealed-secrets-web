#!/bin/bash
# arg 1 method
# arg 2 src file
# arg 3 expected file
set -e

./toSecret.sh ${2} \
  | curl -s --show-error -H "Content-Type: application/json" -X POST --data @- http://localhost/ssw/api/${1} \
  | jq -r .secret \
  | diff ./${3} -

echo OK
