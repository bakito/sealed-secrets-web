load('ext://helm_resource', 'helm_resource', 'helm_repo')

config.define_bool('deploy', usage='Deploy sealed-secrets-web chart')
config.define_bool('deploy-app', usage='Deploy sealed-secrets-web chart')
config.define_bool('disable-deploy', usage='Disable sealed-secrets-web deployment')
config.define_bool('disable-deployment', usage='Disable sealed-secrets-web deployment')
config.define_bool('e2e', usage='Run in E2E mode (disables sealed-secrets-web deployment)')

cfg = config.parse()

deploy = True
if "deploy" in cfg and cfg["deploy"] != None:
    deploy = cfg["deploy"]
elif "deploy-app" in cfg and cfg["deploy-app"] != None:
    deploy = cfg["deploy-app"]

if cfg.get("disable-deploy") or cfg.get("disable-deployment") or cfg.get("e2e"):
    deploy = False

NAMESPACE = "sealed-secrets-web"
SEALED_SECRETS_NAMESPACE = "sealed-secrets"


# =============================================================================
# Bitnami Sealed Secrets
# =============================================================================
helm_repo(
    "sealed-secrets",
    "https://bitnami.github.io/sealed-secrets",
    resource_name="sealed-secrets-repo",
)

helm_resource(
    "sealed-secrets",
    "sealed-secrets/sealed-secrets",
    namespace=SEALED_SECRETS_NAMESPACE,
    flags=["--create-namespace", "--wait", "--timeout", "5m"],
    resource_deps=["sealed-secrets-repo"],
)

# =============================================================================
# sealed-secrets-web
# =============================================================================
#
# Deploy the local Helm chart or build the image for E2E testing.
#
# The E2E values are the same values used by the repository's E2E setup.
#

if deploy:
    docker_build(
        "sealed-secrets-web",
        ".",
        extra_tag=["sealed-secrets-web:latest", "sealed-secrets-web:e2e"],
    )

    helm_resource(
        "sealed-secrets-web",
        "./chart",
        namespace=NAMESPACE,
        image_deps=["sealed-secrets-web"],
        image_keys=[("image.repository", "image.tag")],
        flags=[
            "--create-namespace",
            "--wait",
            "--timeout", "5m",
            "--values", "testdata/e2e/e2e-values.yaml",
            "--set", "format=yaml",
        ],
        resource_deps=["sealed-secrets"],
        port_forwards=["8080:8080"],
    )

    print("""
============================================================
sealed-secrets-web E2E environment
============================================================

Web UI:
  http://localhost:8080

Namespaces:
  Sealed Secrets:      %s
  sealed-secrets-web:  %s

Run E2E tests:
  ./testdata/e2e/chainsaw/run.sh

============================================================
""" % (
        SEALED_SECRETS_NAMESPACE,
        NAMESPACE,
    ))
else:
    local_resource(
        "sealed-secrets-web",
        cmd="docker build -t sealed-secrets-web:latest -t sealed-secrets-web:e2e .",
        deps=[
            "main.go",
            "pkg",
            "Dockerfile",
            "assets",
            "static",
            "templates",
            "go.mod",
            "go.sum",
        ],
    )

    print("""
============================================================
sealed-secrets-web E2E environment (deployment disabled)
============================================================

Namespaces:
  Sealed Secrets:      %s

Run E2E tests:
  ./testdata/e2e/chainsaw/run.sh

============================================================
""" % (
        SEALED_SECRETS_NAMESPACE,
    ))
