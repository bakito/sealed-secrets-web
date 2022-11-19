#!/bin/bash
set -e

echo "Test /api/dencode should b64 encode secret springData having yaml input and yaml output"
curl --silent --show-error --request POST 'http://localhost/ssw/api/dencode' \
  --header 'Content-Type: application/x-yaml' \
  --header 'Accept: application/x-yaml' \
  --data-binary '@stringData.yaml' \
  | diff --strip-trailing-cr --ignore-blank-lines data.yaml -

echo "Test /api/dencode should b64 decode secret data having yaml input and yaml output"
curl --silent --show-error --request POST 'http://localhost/ssw/api/dencode' \
  --header 'Content-Type: application/x-yaml' \
  --header 'Accept: application/x-yaml' \
  --data-binary '@data.yaml' \
  | yq --prettyPrint | diff --strip-trailing-cr --ignore-blank-lines stringData.yaml -

echo "Test /api/dencode should b64 encode secret springData having yaml input and json output"
curl --silent --show-error --request POST 'http://localhost/ssw/api/dencode' \
  --header 'Content-Type: application/x-yaml' \
  --header 'Accept: application/json' \
  --data-binary '@stringData.yaml' \
  | jq --sort-keys . \
  | diff <(jq --sort-keys . data.json)  -

echo "Test /api/dencode should b64 decode secret data having yaml input and json output"
curl --silent --show-error --request POST 'http://localhost/ssw/api/dencode' \
  --header 'Content-Type: application/x-yaml' \
  --header 'Accept: application/json' \
  --data-binary '@data.yaml' \
  | jq --sort-keys . \
  | diff <(jq --sort-keys . stringData.json) -

echo "Test /api/dencode should b64 encode secret springData having json input and json output"
curl --silent --show-error --request POST 'http://localhost/ssw/api/dencode' \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/json' \
  --data-binary '@stringData.json' \
  | jq --sort-keys . \
  | diff <(jq --sort-keys . data.json)  -

echo "Test /api/dencode should b64 decode secret data having json input and json output"
curl --silent --show-error --request POST 'http://localhost/ssw/api/dencode' \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/json' \
  --data-binary '@data.json' \
  | jq --sort-keys . \
  | diff <(jq --sort-keys . stringData.json) -


echo "Test /api/dencode should b64 encode secret springData having json input and yaml output"
curl --silent --show-error --request POST 'http://localhost/ssw/api/dencode' \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/x-yaml' \
  --data-binary '@stringData.json' \
  | diff --strip-trailing-cr --ignore-blank-lines data.yaml -

echo "Test /api/dencode should b64 decode secret data having json input and yaml output"
curl --silent --show-error --request POST 'http://localhost/ssw/api/dencode' \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/x-yaml' \
  --data-binary '@data.json' \
  | yq --prettyPrint | diff --strip-trailing-cr --ignore-blank-lines stringData.yaml -
