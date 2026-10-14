#!/bin/bash
set -e

echo "Test /api/keys should return the list of configured keys"

KEYS=$(curl --silent --show-error --request GET 'http://localhost/api/keys')
echo "$KEYS" | jq -e '.keys | length == 2'
echo "$KEYS" | jq -e '.keys | contains(["primary"])'
echo "$KEYS" | jq -e '.keys | contains(["secondary"])'

echo "Test /api/kubeseal with primary key should seal secret"

SEALED_SECRET=$(curl --silent --show-error --request POST 'http://localhost/api/kubeseal' \
  --header 'Accept: application/yaml' \
  --header 'X-Sealing-Key: primary' \
  --data-binary '@stringData.yaml')

echo "$SEALED_SECRET" | yq -r .apiVersion | grep --quiet "bitnami.com/v1alpha1"
echo "$SEALED_SECRET" | yq -r .kind | grep --quiet "SealedSecret"
echo "$SEALED_SECRET" | yq -r .metadata.name | grep --quiet "mysecretname"

echo "Test /api/kubeseal with secondary key should seal secret"

SEALED_SECRET=$(curl --silent --show-error --request POST 'http://localhost/api/kubeseal' \
  --header 'Accept: application/yaml' \
  --header 'X-Sealing-Key: secondary' \
  --data-binary '@stringData.yaml')

echo "$SEALED_SECRET" | yq -r .apiVersion | grep --quiet "bitnami.com/v1alpha1"
echo "$SEALED_SECRET" | yq -r .kind | grep --quiet "SealedSecret"
echo "$SEALED_SECRET" | yq -r .metadata.name | grep --quiet "mysecretname"

echo "Test /api/certificate with primary key should return valid certificate"

curl --silent --show-error --request GET 'http://localhost/api/certificate' \
  --header 'X-Sealing-Key: primary' \
  | openssl x509 -checkend 0 -dates

echo "Test /api/certificate with secondary key should return valid certificate"

curl --silent --show-error --request GET 'http://localhost/api/certificate' \
  --header 'X-Sealing-Key: secondary' \
  | openssl x509 -checkend 0 -dates

echo "Test /api/raw with primary key should seal raw data"

curl --silent --show-error -H "Content-Type: application/json" \
  -H "X-Sealing-Key: primary" \
  -X POST --data @raw.json http://localhost/api/raw \
  | jq -r -e .secret

echo "Test /api/raw with secondary key should seal raw data"

curl --silent --show-error -H "Content-Type: application/json" \
  -H "X-Sealing-Key: secondary" \
  -X POST --data @raw.json http://localhost/api/raw \
  | jq -r -e .secret

echo "All multi-key tests passed"
