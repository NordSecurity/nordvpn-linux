FROM ghcr.io/nordsecurity/nordvpn-linux/golang:debian-bullseye-20260706@sha256:4b086806ab44e4be65568463d2a3b7f2b8387e0be38974628809b26fd630e23a

ENV GOBIN=/usr/bin

# nfpm version: v2.43.4 github.com/goreleaser/nfpm/v2/cmd/nfpm@d0d484200d974e7ac6d06ed6d5f02d82d4952b3f
# chglog version: v0.7.3 github.com/goreleaser/chglog/cmd/chglog@f22c2a6f688d1d20958839794c5d5b84ccfc9ef3
RUN go install github.com/goreleaser/nfpm/v2/cmd/nfpm@d0d484200d974e7ac6d06ed6d5f02d82d4952b3f \
    && go install github.com/goreleaser/chglog/cmd/chglog@f22c2a6f688d1d20958839794c5d5b84ccfc9ef3 

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN apt-get update && \
    apt-get -y --no-install-recommends install elfutils gettext-base python3 file && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    go clean -cache -modcache && \
    nfpm --version
