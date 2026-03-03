#!/bin/bash
set -e
docker load -i /tmp/image.tar
kind load docker-image localhost:5001/sealed-secrets-web:e2e --name kind
