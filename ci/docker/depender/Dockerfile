FROM ghcr.io/nordsecurity/nordvpn-linux/golang:ubuntu-20.04-20240725

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN go install github.com/google/go-licenses@v1.6.0 && \
    go clean -cache -modcache
