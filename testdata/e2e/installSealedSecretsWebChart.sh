#!/bin/bash
set -e

helm upgrade --install sealed-secrets-web  charts/sealed-secrets-web \
  --namespace sealed-secrets-web \
  --create-namespace \
  --set image.tag=main \
  --set format=${1} \
  --set image.args[0]="--kubeseal-arguments=--controller-name=sealed-secrets --controller-namespace=sealed-secrets" \
  --atomic
