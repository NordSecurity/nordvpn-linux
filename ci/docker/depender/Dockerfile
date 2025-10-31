FROM ghcr.io/nordsecurity/nordvpn-linux/golang:debian-bullseye-20251014@sha256:a5156248a5497a3b261bd5288abb3ddc4b0f8177803e06ea53f50d95bbda5c7e

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN go install github.com/google/go-licenses@v1.6.0 && \
    go clean -cache -modcache
