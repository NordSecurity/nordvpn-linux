FROM ghcr.io/nordsecurity/nordvpn-linux/golang:debian-bullseye-20251014@sha256:a5156248a5497a3b261bd5288abb3ddc4b0f8177803e06ea53f50d95bbda5c7e

ENV GOBIN=/usr/bin

# nfpm version: v2.43.0 github.com/goreleaser/nfpm/v2/cmd/nfpm@2e47af64a3f33d32ec1395dbd56165c4feb2bc21
# chglog version: v0.7.3 github.com/goreleaser/chglog/cmd/chglog@f22c2a6f688d1d20958839794c5d5b84ccfc9ef3
RUN go install github.com/goreleaser/nfpm/v2/cmd/nfpm@2e47af64a3f33d32ec1395dbd56165c4feb2bc21 \
    && go install github.com/goreleaser/chglog/cmd/chglog@f22c2a6f688d1d20958839794c5d5b84ccfc9ef3 

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN apt-get update && \
    apt-get -y --no-install-recommends install elfutils gettext-base && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    go clean -cache -modcache && \
    nfpm --version
