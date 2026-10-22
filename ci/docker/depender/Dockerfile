FROM ghcr.io/nordsecurity/nordvpn-linux/golang:ubuntu22.04-20262110@sha256:df63d87fdd8700cf5b8d574152a625ccf86873ea34ebd86fa9f456269348b834

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN go install github.com/google/go-licenses@v1.6.0 && \
    go clean -cache -modcache
