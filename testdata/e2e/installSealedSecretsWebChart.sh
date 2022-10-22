#!/bin/bash
set -e

helm upgrade --install sealed-secrets-web charts/sealed-secrets-web \
  --namespace sealed-secrets-web \
  --create-namespace \
  -f testdata/e2e/e2e-values.yaml \
  --set format=${1} \
  --atomic

echo "Wait for service to respond"
timeout 30s bash <<EOT
while true; do
  if [[ "\$(curl -s http://localhost/ssw/_health)" == "OK" ]]; then
    echo "Service Running"
    break
  fi
  sleep 1
  echo -n "."
done
EOT
