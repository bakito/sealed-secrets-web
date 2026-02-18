# Build stage with multi-arch support
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
WORKDIR /go/src/app

RUN apk update && apk add upx

ARG VERSION=main
ARG BUILD="N/A"
ARG TARGETPLATFORM
ARG BUILDPLATFORM

# Parse target platform
RUN case "$TARGETPLATFORM" in \
  "linux/amd64") export GOARCH=amd64 ;; \
  "linux/arm64") export GOARCH=arm64 ;; \
  "linux/arm/v7") export GOARCH=arm GOARM=7 ;; \
  "linux/386") export GOARCH=386 ;; \
  *) export GOARCH=amd64 ;; \
  esac && echo "GOARCH=$GOARCH" > /tmp/buildenv && echo "GOARM=${GOARM}" >> /tmp/buildenv

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux

COPY . /go/src/app/

RUN . /tmp/buildenv && \
  go build -a -installsuffix cgo \
    -ldflags="-w -s -X github.com/bakito/sealed-secrets-web/pkg/version.Version=${VERSION} -X github.com/bakito/sealed-secrets-web/pkg/version.Build=${BUILD}" \
    -o sealed-secrets-web . && \
  upx -q sealed-secrets-web

# Final application image
FROM alpine:latest
WORKDIR /opt/go

LABEL maintainer="bakito <github@bakito.ch>" \
      org.opencontainers.image.description="A web interface for Sealed Secrets by Bitnami."

EXPOSE 8080
RUN apk add --no-cache dumb-init

COPY --from=builder /go/src/app/sealed-secrets-web /opt/go/sealed-secrets-web

USER 1001
ENTRYPOINT ["/usr/bin/dumb-init", "--", "/opt/go/sealed-secrets-web"]
