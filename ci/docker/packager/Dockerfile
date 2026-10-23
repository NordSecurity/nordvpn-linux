# versions.yaml: FROM ghcr.io/nordsecurity/nordvpn-linux/golang:{image-golang}
FROM ghcr.io/nordsecurity/nordvpn-linux/golang:ubuntu22.04-20262110@sha256:df63d87fdd8700cf5b8d574152a625ccf86873ea34ebd86fa9f456269348b834

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
