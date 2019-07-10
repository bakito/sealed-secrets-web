<div align="center">
  <img src="./assets/logo.png" />
  <br><br>

  A web interface for [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) by Bitnami.

  <img src="./assets/example1.png" width="100%" />
  <img src="./assets/example2.png" width="100%" />
</div>

**Sealed Secrets Web** is a web interface for [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) by Bitnami. The web interface let you encode, decode and encrypt you secrets. It uses the same functions as the [kubeseal](https://github.com/bitnami-labs/sealed-secrets/tree/master/cmd/kubeseal) command-line tool to encrypt your secrets. The web interface should be installed to your Kubernetes cluster, so your developers do not need access to your cluster via kubectl.

## Installation

**sealed-secrets-web** can be installed via our Helm chart:

```sh
helm repo add ricoberger https://ricoberger.github.io/helm-charts
helm up

helm upgrade --install sealed-secrets-web ricoberger/sealed-secrets-web
```

To modify the settings for Sealed Secrets you can modify the arguments for the Docker image with the `--set` flag. For example you can set a different `controller-name` during the installation with the following command:

```sh
helm upgrade --install sealed-secrets-web ricoberger/sealed-secrets-web --set image.args={"--controller-name=sealed-secrets"}
```
