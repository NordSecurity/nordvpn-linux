FROM ghcr.io/nordsecurity/nordvpn-linux/golang:ubuntu22.04-20260709@sha256:b26ebd8e021c721dba2ce57b7d12dd269eab038dcd36df9a7e4ee239e340a3fb

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN go install github.com/google/go-licenses@v1.6.0 && \
    go clean -cache -modcache
