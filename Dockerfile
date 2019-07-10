FROM alpine:latest

ARG REVISION
ARG VERSION

LABEL maintainer="Rico Berger"
LABEL git.ref=$REVISION
LABEL git.version=$VERSION
LABEL git.url="https://github.com/ricoberger/sealed-secrets-web"

RUN apk add --no-cache --update curl ca-certificates
HEALTHCHECK --interval=10s --timeout=3s --retries=3 CMD curl --fail http://localhost:8080/_health || exit 1

RUN addgroup -g 1000 sealedsecretsweb && \
    adduser -D -u 1000 -G sealedsecretsweb sealedsecretsweb
USER sealedsecretsweb

COPY ./bin/sealedsecretsweb-linux-amd64  /bin/sealedsecretsweb
COPY ./static/ /static/
EXPOSE 8080

ENTRYPOINT  [ "/bin/sealedsecretsweb" ]
