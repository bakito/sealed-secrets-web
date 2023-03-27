[![end-2-end Helm Chart Tests](https://github.com/bakito/sealed-secrets-web/actions/workflows/e2e.yaml/badge.svg)](https://github.com/bakito/sealed-secrets-web/actions/workflows/e2e.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bakito/sealed-secrets-web)](https://goreportcard.com/report/github.com/bakito/sealed-secrets-web)
[![Coverage Status](https://coveralls.io/repos/github/bakito/sealed-secrets-web/badge.svg?branch=main&service=github)](https://coveralls.io/github/bakito/sealed-secrets-web?branch=main)

<div align="center">
  <img src="./assets/logo.png" />
  <br><br>

  A web interface for [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) by Bitnami.

  <img src="./assets/example1.png" width="100%" />
  <img src="./assets/example2.png" width="100%" />
</div>

**Sealed Secrets Web** is a web interface for [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) by Bitnami. The web interface let you encode, decode the keys in the `data` field of a secret, load existing Sealed Secrets and create Sealed Secrets. Under the hood it uses the [kubeseal](https://github.com/bitnami-labs/sealed-secrets/tree/master/cmd/kubeseal) command-line tool to encrypt your secrets. The web interface should be installed to your Kubernetes cluster, so your developers do not need access to your cluster via kubectl.

- **Encode:** Base64 encodes each key in the `stringData` field in a secret.
- **Decode:** Base64 decodes each key in the `data` field in a secret.
- **Secrets:** Returns a list of all Sealed Secrets in all namespaces. With a click on the Sealed Secret the decrypted Kubernetes secret is loaded.
- **Seal:** Encrypt a Kubernetes secret and creates the Sealed Secret.

## Installation

**sealed-secrets-web** can be installed via our Helm chart:

```sh
helm repo add bakito https://charts.bakito.net
helm repo update

helm upgrade --install sealed-secrets-web bakito/sealed-secrets-web
```

To modify the settings for Sealed Secrets you can modify the arguments for the Docker image with the `--set` flag. For example you can set a different `controller-name` during the installation with the following command:

```sh
helm upgrade --install sealed-secrets-web bakito/sealed-secrets-web \
  --set sealedSecrets.namespace=sealed-secrets \
  --set sealedSecrets.serviceName=sealed-secrets

```

or if you want to disable ability to load existing secrets, and use the tool purelly to seal new ones you can use:
```sh
helm upgrade --install sealed-secrets-web bakito/sealed-secrets-web \
  --set disableLoadSecrets=true
```

To render templates locally:

```sh
cd chart
helm template . -f values.yaml
```

You can check helm values available at https://github.com/bakito/sealed-secrets-web/blob/main/chart/values.yaml
Also, check available application options at https://github.com/bakito/sealed-secrets-web/blob/main/pkg/config/types.go#L14-L22

## Api Usage

### Get current certificate

```bash
curl --request GET 'https://<SEALED_SECRETS_WEB_BASE_URL>/api/certificate'
```

### Seal a secret using servers certificate

#### having sealed secret as yaml output

```bash
curl --request POST 'https://<SEALED_SECRETS_WEB_BASE_URL>/api/kubeseal' \
  --header 'Accept: application/x-yaml' \
  --data-binary '@stringData.yaml'
```

#### having sealed secret as json output

```bash
curl --request POST 'https://<SEALED_SECRETS_WEB_BASE_URL>/api/kubeseal' \
  --header 'Accept: application/json' \
  --data-binary '@stringData.yaml'
```

#### sealing one value with default scope

```bash
curl -request POST 'https://<SEALED_SECRETS_WEB_BASE_URL>/api/raw' \
     --header 'Content-Type: application/json' \
     --data '{ "name": "mysecretname", "namespace": "mysecretnamespace", "value": "value to seal" }'
```

## Development

For development, we are using a local Kubernetes cluster using kind. When the cluster is created we install **Sealed Secrets** using Helm:

```sh
# install registry
docker run -d --restart=always -p "127.0.0.1:5001:5000" --name kind-registry registry:2

# startup kind
kind create cluster --config=testdata/e2e/kind/config.yaml

# setup registry
docker network connect kind kind-registry
kubectl apply -f testdata/e2e/kind/configmap-registry.yaml

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
helm install sealed-secrets sealed-secrets/sealed-secrets \
  --namespace sealed-secrets \
  --create-namespace \
  --atomic

install sealed secrets web
./testdata/e2e/installSealedSecretsWebChart.sh yaml

```

Access the interface via http://localhost/ssw
