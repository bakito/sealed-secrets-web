#!/bin/bash
set -e

echo "Test /api/certificate should return public valid certificate"

curl --silent --show-error --request GET http://localhost/ssw/api/certificate \
    | openssl x509 -checkend 0 -dates
