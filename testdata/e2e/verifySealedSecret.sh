#!/bin/bash
set -e

if [[ ${1} = *.json ]]; then
  cat ${1} | jq -r .apiVersion | grep "bitnami.com/v1alpha1"
  cat ${1} | jq -r .kind | grep "SealedSecret"
  cat ${1} | jq -r .metadata.name | grep "mysecretname"
  cat ${1} | jq -r .metadata.namespace | grep "mysecretnamespace"
else
  cat ${1} | yq .apiVersion | grep "bitnami.com/v1alpha1"
  cat ${1} | yq .kind | grep "SealedSecret"
  cat ${1} | yq .metadata.name | grep "mysecretname"
  cat ${1} | yq .metadata.namespace | grep "mysecretnamespace"
fi
