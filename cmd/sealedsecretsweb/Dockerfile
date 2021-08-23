FROM alpine:latest

ARG REVISION
ARG VERSION

LABEL maintainer="Rico Berger" \
      git.ref=$REVISION \
      git.version=$VERSION \
      git.url="https://github.com/ricoberger/sealed-secrets-web"

ENV KUBESEAL_VERSION=v0.16.0

RUN apk add --no-cache --update curl ca-certificates && \
    curl -L https://github.com/bitnami-labs/sealed-secrets/releases/download/${KUBESEAL_VERSION}/kubeseal-linux-amd64 -o /usr/local/bin/kubeseal && \
    chmod +x /usr/local/bin/kubeseal \

HEALTHCHECK --interval=10s --timeout=3s --retries=3 CMD curl --fail http://localhost:8080/_health || exit 1

RUN addgroup -g 1000 sealedsecretsweb && \
    adduser -D -u 1000 -G sealedsecretsweb sealedsecretsweb
USER 1000

COPY ./bin/sealedsecretsweb-linux-amd64  /bin/sealedsecretsweb
COPY ./static/ /static/
EXPOSE 8080

ENTRYPOINT  [ "/bin/sealedsecretsweb" ]
