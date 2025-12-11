#!/bin/bash
set -e
docker build -f Dockerfile --build-arg VERSION=e2e-tests -t localhost:5001/sealed-secrets-web:e2e .
docker save localhost:5001/sealed-secrets-web:e2e -o /tmp/image.tar
