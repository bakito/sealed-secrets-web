#!/bin/bash

helm template cs-test-1 chart  -n '($namespace)' > ./testdata/e2e/chainsaw/template-default.yaml
helm template cs-test-2 chart --set includeLocalNamespaceOnly=true -n '($namespace)' > ./testdata/e2e/chainsaw/template-local-ns.yaml
helm template cs-test-3 chart --set sealedSecrets.serviceName=  -n '($namespace)' > ./testdata/e2e/chainsaw/template-service-name.yaml

chainsaw test --test-dir ./testdata/e2e/chainsaw/
