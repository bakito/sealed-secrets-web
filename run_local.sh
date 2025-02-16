#!/bin/bash

set -eo pipefail

# install registry
docker start kind-registry || docker run -d --restart=always -p "127.0.0.1:5001:5000" --name kind-registry registry:2

# startup kind
if ! kind get clusters | grep -qFx "kind"; then
  curl -L https://raw.githubusercontent.com/bakito/kind-with-registry-action/main/kind-config.yaml -o testdata/e2e/kind-config.yaml
  kind create cluster --name kind --config=testdata/e2e/kind-config.yaml
fi

# setup registry
docker network connect kind kind-registry || true
kubectl apply -f https://raw.githubusercontent.com/bakito/kind-with-registry-action/main/configmap-registry.yaml

# setup ingress
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

# build image
./testdata/e2e/buildImage.sh

# install sealed secrets
helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
helm upgrade --install sealed-secrets sealed-secrets/sealed-secrets \
  --namespace sealed-secrets \
  --create-namespace --wait

# install sealed secrets web
./testdata/e2e/installSealedSecretsWebChart.sh yaml
