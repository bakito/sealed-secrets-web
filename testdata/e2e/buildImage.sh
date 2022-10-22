#!/bin/bash
set -e
docker build -f Dockerfile -t localhost:5001/sealed-secrets-web:build .
docker push  localhost:5001/sealed-secrets-web:build
