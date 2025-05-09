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

**Sealed Secrets Web** is a web interface for [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) by
Bitnami. The web interface let you encode, decode the keys in the `data` field of a secret, load existing Sealed Secrets
and create Sealed Secrets. Under the hood it uses Sealed Secrets service API to encrypt your secrets. The web interface
should be installed to your Kubernetes cluster, so your developers do not need access to
your cluster via kubectl.

- **Encode:** Base64 encodes each key in the `stringData` field in a secret.
- **Decode:** Base64 decodes each key in the `data` field in a secret.
- **Secrets:** Returns a list of all Sealed Secrets in all namespaces. With a click on the Sealed Secret the decrypted
  Kubernetes secret is loaded.
- **Seal:** Encrypt a Kubernetes secret and creates the Sealed Secret.
- **Validate:** Validate a Sealed Secret.

## Installation

**sealed-secrets-web** can be installed via our Helm chart:

```sh
helm repo add bakito https://charts.bakito.net
helm repo update

helm upgrade --install sealed-secrets-web bakito/sealed-secrets-web
```

To modify the settings for Sealed Secrets you can modify the arguments for the Docker image with the `--set` flag. For
example you can set a different `controller-name` during the installation with the following command:

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
Also, check available application options
at https://github.com/bakito/sealed-secrets-web/blob/main/pkg/config/types.go#L14-L22

## Api Usage

### Get current certificate

```bash
curl --request GET 'https://<SEALED_SECRETS_WEB_BASE_URL>/api/certificate'
```

### Seal a secret using servers certificate

#### having sealed secret as yaml output

```bash
curl --request POST 'https://<SEALED_SECRETS_WEB_BASE_URL>/api/kubeseal' \
  --header 'Accept: application/yaml' \
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

### Validate sealed secret

> **_NOTE:_**  Validate is only available when using cluster internal api (e.g. certURL not set)
> see [bitnami-labs/sealed-secrets](https://github.com/bitnami-labs/sealed-secrets/issues/1208)

```bash
curl --request POST 'https://<SEALED_SECRETS_WEB_BASE_URL>/api/validate' \
  --header 'Accept: application/yaml' \
  --data-binary '@stringData.yaml'
```

## Development

For development, we are using a local Kubernetes cluster using kind. When the cluster is created we install **Sealed
Secrets** using Helm:

```sh
./run_local.sh
```

Access the interface via http://localhost/ssw

## Traefik

This section is about using sealed-secrets-web with Traefik ingress controller.

Traefik does not by default strip the path when forwarding to application. If your path is `localhost/seal`, then your route will be parsed by Traefik and your application will be accessed at `/seal`, not `/`.

To configure Traefik correctly, apply the following resource:

```bash
$ cat << EOF > traefik-strip-prefix-middleware.yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: strip-prefix
  namespace: kube-system # can be anything, but needs ingress.metadata.annotations spec: traefik.ingress.kubernetes.io/router.middlewares: namespace-strip-prefix@kubernetescrd
spec:
  stripPrefixRegex:
    regex:
    - ^/[^/]+
EOF

$ kubectl apply -f traefik-strip-prefix-middleware.yaml
```

Next, in your `values.yaml`, adapt the following to your host:

```yaml
ingress:
  enabled: true
  className: traefik
  hosts:
  - host: your.host
    paths:
    - path: /seal
      pathType: ImplementationSpecific
  annotations:
    traefik.ingress.kubernetes.io/router.middlewares: kube-system-strip-prefix@kubernetescrd

sealedSecrets:
  certURL: https://your.host/v1/cert.pem

webLogs: true
webContext: /seal
```
