#!/bin/bash
set -e

./runTestHealth.sh

if [ "${1}" != "skip-validate" ]; then
  ./runTestValidate.sh
else
  echo "Validation test is skipped"
fi

./runTestKubeseal.sh

./runTestCertificate.sh

./runTestDencode.sh

curl -s --show-error -H "Content-Type: application/json" -X POST --data @raw.json http://localhost/ssw/api/raw \
    | jq -r -e .secret
