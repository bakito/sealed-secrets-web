#!/bin/bash
set -e

echo "Test /_health Health Check"

curl --show-error --silent -w "\n%{http_code}" 'http://localhost/ssw/_health'

echo
