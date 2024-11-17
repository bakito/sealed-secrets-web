#!/bin/bash

helm template chart  | yq eval 'del(.metadata.namespace)' > ./testdata/e2e/chainsaw/template-default.yaml
helm template chart  --set includeLocalNamespaceOnly=true | yq eval 'del(.metadata.namespace)' > ./testdata/e2e/chainsaw/template-includeLocalNamespaceOnly.yaml


chainsaw test --test-dir ./testdata/e2e/chainsaw/
