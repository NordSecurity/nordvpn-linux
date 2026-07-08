FROM ghcr.io/nordsecurity/nordvpn-linux/golang:debian-bullseye-20260708@sha256:290eb49710a0a5a01bbe2ce063cae56f8d8c26f401cc2c32fd43e8a9d6fab46e

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN go install github.com/google/go-licenses@v1.6.0 && \
    go clean -cache -modcache
