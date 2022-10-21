#!/bin/bash

jq -R -s --argfile json testdata/e2e/secret.json -f testdata/e2e/secret.jq testdata/e2e/${1}
