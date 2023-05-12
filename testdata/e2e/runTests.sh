#!/bin/bash
set -e

./runTestKubeseal.sh

./runTestCertificate.sh

./runTestDencode.sh

curl -s --show-error -H "Content-Type: application/json" -X POST --data @raw.json http://localhost/ssw/api/raw \
    | jq -r -e .secret

exit 1
