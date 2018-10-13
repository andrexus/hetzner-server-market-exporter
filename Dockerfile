FROM andrexus/baseimage

ARG APP_VERSION=0.1.1
ARG DOWNLOAD_URL=https://github.com/andrexus/hetzner-server-market-exporter/releases/download/v$APP_VERSION/linux_amd64_hetzner-server-market-exporter

WORKDIR /srv

RUN apk update && \
    update-ca-certificates && \
    wget -q $DOWNLOAD_URL -O /srv/hetzner-server-market-exporter && \
    chmod +x /srv/hetzner-server-market-exporter && \
    ln -s /srv/hetzner-server-market-exporter /usr/bin/hetzner-server-market-exporter && \
    rm -rf /var/cache/apk/*

EXPOSE 8080

USER app

CMD ["hetzner-server-market-exporter"]
