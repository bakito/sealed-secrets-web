FROM bitnami/golang:1.17 as builder

WORKDIR /go/src/app

RUN apt-get update && apt-get install -y upx

ARG VERSION=main
ARG BUILD="N/A"

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux

ADD ./go.mod /tmp/go.mod
RUN export KUBESEAL_VERSION=$(cat /tmp/go.mod  | grep github.com/bitnami-labs/sealed-secrets | awk '{print $2}') && \
    export KUBESEAL_ARCH=$(dpkg --print-architecture) && \
    export KUBESEAL_FILE="kubeseal-${KUBESEAL_VERSION#?}-linux-${KUBESEAL_ARCH}"; \
    export KUBESEAL_URL="https://github.com/bitnami-labs/sealed-secrets/releases/download/${KUBESEAL_VERSION}/${KUBESEAL_FILE}.tar.gz"; \
    echo "Download kubeseal ${KUBESEAL_VERSION}/${KUBESEAL_ARCH} from ${KUBESEAL_URL}" && \
    curl -L ${KUBESEAL_URL} | tar -xz -C /tmp && \
    chmod +x /tmp/kubeseal && \
    upx -q /tmp/kubeseal

ADD . /go/src/app/

RUN go build -a -installsuffix cgo -ldflags="-w -s -X github.com/bakito/sealed-secrets-web/pkg/version.Version=${VERSION} -X github.com/bakito/sealed-secrets-web/pkg/version.Build=${BUILD}" -o sealed-secrets-web . && \
    upx -q sealed-secrets-web


# application image
FROM alpine:latest
WORKDIR /opt/go

LABEL maintainer="bakito <github@bakito.ch>"
EXPOSE 8080
RUN apk add --no-cache dumb-init
ENTRYPOINT ["/usr/bin/dumb-init", "--","/opt/go/sealed-secrets-web"]
COPY --from=builder /go/src/app/sealed-secrets-web /opt/go/sealed-secrets-web
COPY --from=builder /tmp/kubeseal /usr/local/bin/kubeseal
USER 1001
