FROM golang:1.20.1-buster

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

ENV GOBIN=/usr/bin
ENV GOPATH=/go

RUN go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.5.1 && \
    go install github.com/goreleaser/chglog/cmd/chglog@v0.1.2

RUN apt-get update; \
    apt-get -y install elfutils gettext-base; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd --system nordvpn && groupadd -g ${GROUP_ID} packager && useradd -l -m -u ${USER_ID} -g packager -G nordvpn packager && echo "packager ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

RUN chown -R packager:packager /go

USER packager

CMD ["exec", "$@"]
