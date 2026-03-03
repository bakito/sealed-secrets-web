#!/bin/bash
set -e
docker load -i /tmp/image.tar
kind load docker-image sealed-secrets-web:e2e --name e2e
