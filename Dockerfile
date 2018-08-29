FROM alpine

ARG APP_VERSION=0.1.0
ARG DOWNLOAD_URL=https://github.com/andrexus/hetzner-server-market-exporter/releases/download/v$APP_VERSION/linux_amd64_hetzner-server-market-exporter

LABEL maintainer="Andrew Tarasenko andrexus@gmail.com"

WORKDIR /usr/local/bin

RUN apk update && \
    apk add ca-certificates wget && \
    update-ca-certificates && \
    wget -q $DOWNLOAD_URL -O /usr/local/bin/hetzner-server-market-exporter && \
    chmod +x /usr/local/bin/hetzner-server-market-exporter && \
    rm -rf /var/cache/apk/*

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/hetzner-server-market-exporter"]