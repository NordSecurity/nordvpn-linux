FROM golang:1.20.1-buster

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

ENV GOBIN=/usr/bin

RUN go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.5.1 && \
    go install github.com/goreleaser/chglog/cmd/chglog@v0.1.2

RUN apt-get update; \
    apt-get -y install elfutils gettext-base; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*

CMD ["exec", "$@"]
