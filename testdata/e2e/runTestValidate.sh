#!/bin/bash
set -e

curl --version

echo "Test /api/validate should respond 200 if sealed secret is valid"

SEALED_SECRET=$(curl --silent --show-error --request POST 'http://localhost/ssw/api/kubeseal' \
  --header 'Accept: application/x-yaml' \
  --data-binary '@stringData.yaml')

echo "$SEALED_SECRET" | yq -r .apiVersion | grep --quiet "bitnami.com/v1alpha1"
echo "$SEALED_SECRET" | yq -r .kind | grep --quiet "SealedSecret"
echo "$SEALED_SECRET" | yq -r .metadata.name | grep --quiet "mysecretname"
echo "$SEALED_SECRET" | yq -r .metadata.namespace | grep --quiet "mysecretnamespace"

RESPONSE=$(curl --silent --show-error --request POST 'http://localhost/ssw/api/validate' \
  --header 'Accept: text/plain' \
  --data-binary "$SEALED_SECRET" \
  --output /dev/null -w "%{http_code}" )

echo "$RESPONSE" | grep --quiet 200

echo "Test /api/validate should respond 400 if sealed secret is invalid"

INVALID_SECRET=$(curl --silent --show-error --request POST 'http://localhost/ssw/api/kubeseal' \
  --header 'Accept: application/x-yaml' \
  --data-binary '@stringData.yaml' | yq '.metadata.name = "wrongname"')

echo "$INVALID_SECRET" | yq -r .apiVersion | grep --quiet "bitnami.com/v1alpha1"
echo "$INVALID_SECRET" | yq -r .kind | grep --quiet "SealedSecret"
echo "$INVALID_SECRET" | yq -r .metadata.name | grep --quiet "wrongname"
echo "$INVALID_SECRET" | yq -r .metadata.namespace | grep --quiet "mysecretnamespace"

RESPONSE=$(curl --silent --show-error --request POST 'http://localhost/ssw/api/validate' \
  --header 'Accept: text/plain' \
  --data-binary "$INVALID_SECRET" \
  --output /dev/null -w "%{http_code}" )

echo "$RESPONSE" | grep --quiet 400
