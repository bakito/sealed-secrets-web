#!/bin/bash
openssl req -x509 \
  -newkey rsa:4096 \
  -keyout tls.key \
  -out tls.crt \
  -sha256 \
  -days 36500 \
  -nodes \
  -subj '/CN=my-domain.com'
