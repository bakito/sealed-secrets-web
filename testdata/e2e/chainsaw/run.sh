#!/bin/bash

helm template cs-test-1 chart  -n '($namespace)' > ./testdata/e2e/chainsaw/template-default.yaml
helm template cs-test-2 chart --set includeLocalNamespaceOnly=true -n '($namespace)' > ./testdata/e2e/chainsaw/template-local-ns.yaml
helm template cs-test-3 chart --set sealedSecrets.serviceName=  -n '($namespace)' > ./testdata/e2e/chainsaw/template-service-name.yaml
helm template cs-test-4 chart --set disableLoadSecrets=true -n '($namespace)' > ./testdata/e2e/chainsaw/template-disable-load-secrets.yaml
helm template cs-test-5 chart --set disableLoadSecrets=false -n '($namespace)' > ./testdata/e2e/chainsaw/template-enable-load-secrets.yaml

chainsaw test --test-dir ./testdata/e2e/chainsaw/
