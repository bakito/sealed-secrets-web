#!/bin/bash
set -e

echo "Test /api/kubeseal should seal secret having yaml input and yaml output"

SEALED_SECRET=$(curl --silent --show-error --request POST 'http://localhost/ssw/api/kubeseal' \
  --header 'Accept: application/x-yaml' \
  --data-binary '@stringData.yaml')

echo "$SEALED_SECRET" | yq -r .apiVersion | grep --quiet "bitnami.com/v1alpha1"
echo "$SEALED_SECRET" | yq -r .kind | grep --quiet "SealedSecret"
echo "$SEALED_SECRET" | yq -r .metadata.name | grep --quiet "mysecretname"
echo "$SEALED_SECRET" | yq -r .metadata.namespace | grep --quiet "mysecretnamespace"

echo "Test /api/kubeseal should seal secret having json input and yaml output"

SEALED_SECRET=$(curl --silent --show-error --request POST 'http://localhost/ssw/api/kubeseal' \
  --header 'Accept: application/x-yaml' \
  --data-binary '@stringData.json')

echo "$SEALED_SECRET" | yq -r .apiVersion | grep --quiet "bitnami.com/v1alpha1"
echo "$SEALED_SECRET" | yq -r .kind | grep --quiet "SealedSecret"
echo "$SEALED_SECRET" | yq -r .metadata.name | grep --quiet "mysecretname"
echo "$SEALED_SECRET" | yq -r .metadata.namespace | grep --quiet "mysecretnamespace"

echo "Test /api/kubeseal should seal secret having yaml input and json output"

SEALED_SECRET=$(curl --silent --show-error --request POST 'http://localhost/ssw/api/kubeseal' \
  --header 'Accept: application/json' \
  --data-binary '@stringData.yaml')

echo "$SEALED_SECRET" | jq -r .apiVersion | grep --quiet "bitnami.com/v1alpha1"
echo "$SEALED_SECRET" | jq -r .kind | grep --quiet "SealedSecret"
echo "$SEALED_SECRET" | jq -r .metadata.name | grep --quiet "mysecretname"
echo "$SEALED_SECRET" | jq -r .metadata.namespace | grep --quiet "mysecretnamespace"

echo "Test /api/kubeseal should seal secret having json input and json output"

SEALED_SECRET=$(curl --silent --show-error --request POST 'http://localhost/ssw/api/kubeseal' \
  --header 'Accept: application/json' \
  --data-binary '@stringData.json')

echo "$SEALED_SECRET" | jq -r .apiVersion | grep --quiet "bitnami.com/v1alpha1"
echo "$SEALED_SECRET" | jq -r .kind | grep --quiet "SealedSecret"
echo "$SEALED_SECRET" | jq -r .metadata.name | grep --quiet "mysecretname"
echo "$SEALED_SECRET" | jq -r .metadata.namespace | grep --quiet "mysecretnamespace"
