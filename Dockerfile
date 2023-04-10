FROM golang:1.20 as builder
WORKDIR /go/src/app

RUN apt-get update && apt-get install -y upx curl

ARG VERSION=main
ARG BUILD="N/A"
ARG TARGETPLATFORM=linux/amd64

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux

COPY . /go/src/app/

RUN export GOARCH=$(basename ${TARGETPLATFORM}) && \
    echo "GOARCH ${GOARCH}" && \
    go build -a -installsuffix cgo -ldflags="-w -s -X github.com/bakito/sealed-secrets-web/pkg/version.Version=${VERSION} -X github.com/bakito/sealed-secrets-web/pkg/version.Build=${BUILD}" -o sealed-secrets-web . && \
    upx -q sealed-secrets-web


# application image
FROM alpine:latest
WORKDIR /opt/go

LABEL maintainer="bakito <github@bakito.ch>" \
      org.opencontainers.image.description="A web interface for Sealed Secrets by Bitnami."
EXPOSE 8080
RUN apk add --no-cache dumb-init
ENTRYPOINT ["/usr/bin/dumb-init", "--","/opt/go/sealed-secrets-web"]
COPY --from=builder /go/src/app/sealed-secrets-web /opt/go/sealed-secrets-web
USER 1001
