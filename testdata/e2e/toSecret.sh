#!/bin/bash
set -e

jq -R -s --argfile json ./secret.json -f ./secret.jq ./${1}
