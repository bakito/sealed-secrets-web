#!/bin/bash
set -e

helm upgrade --install sealed-secrets-web charts/sealed-secrets-web \
  --namespace sealed-secrets-web \
  --create-namespace \
  -f testdata/e2e/e2e-values.yaml \
  --set format=${1} \
  --atomic
