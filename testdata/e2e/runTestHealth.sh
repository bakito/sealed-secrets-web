#!/bin/bash
set -e

echo "Test /_health Health Check"

curl --show-error --silent 'http://localhost/ssw/_health'
