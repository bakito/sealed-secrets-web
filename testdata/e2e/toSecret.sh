#!/bin/bash

jq -R -s --argfile json ./secret.json -f ./secret.jq ./${1}
