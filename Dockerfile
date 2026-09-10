# Build stage with multi-arch support
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder
WORKDIR /go/src/app

RUN apk add --no-cache upx

ARG VERSION=main
ARG BUILD="N/A"
ARG TARGETPLATFORM
ARG TARGETOS=linux
ARG TARGETARCH
ARG TARGETVARIANT

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=${TARGETOS}

COPY . /go/src/app/

RUN GOARCH=${TARGETARCH} \
  GOARM=$([ "$TARGETARCH" = "arm" ] && { [ -n "$TARGETVARIANT" ] && echo "${TARGETVARIANT#v}" || echo "7"; } || echo "") \
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

# Verify the binary is built for the correct architecture and runs
RUN /opt/go/sealed-secrets-web --version

USER 1001
ENTRYPOINT ["/usr/bin/dumb-init", "--", "/opt/go/sealed-secrets-web"]
