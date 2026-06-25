FROM ghcr.io/nordsecurity/nordvpn-linux/golang:debian-bullseye-20260706@sha256:4b086806ab44e4be65568463d2a3b7f2b8387e0be38974628809b26fd630e23a

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN go install github.com/google/go-licenses@v1.6.0 && \
    go clean -cache -modcache
