#!/bin/bash
set -e
docker load -i /tmp/image.tar
docker push localhost:5001/sealed-secrets-web:e2e
