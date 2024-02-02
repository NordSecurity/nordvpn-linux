FROM ghcr.io/nordsecurity/nordvpn-linux/golang:debian-buster-20240201

# need to build nfpm with patch; expeting sometime
# nfpm upstream will be fixed (issue submitted)
RUN cd && \
    apt update && \
    apt install -y file binutils && \
    sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d && \
    export PATH=$PATH:~/bin && \
    git clone https://github.com/keliramu/nfpm.git && \
    cd nfpm && \
    git checkout ensure-reproducable-deb-new && \
    /root/bin/task setup && \
    /root/bin/task build && \
    strip ./nfpm && \
    file ./nfpm && \
    ls -lh ./nfpm && \
    ./nfpm --version


FROM ghcr.io/nordsecurity/nordvpn-linux/golang:debian-buster-20240201

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

ENV GOBIN=/usr/bin
ENV GOPATH=/go

# from first stage copy only nfpm binary
COPY --from=0 /root/nfpm/nfpm /usr/bin/nfpm

RUN go install github.com/goreleaser/chglog/cmd/chglog@v0.1.2 && \
    apt update && \
    apt -y install elfutils gettext-base sudo && \
    apt clean && \
    rm -rf /var/lib/apt/lists/* && \
    nfpm --version

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd --system nordvpn && \
    groupadd -g ${GROUP_ID} packager && \
    useradd -l -m -u ${USER_ID} -g packager -G nordvpn packager && \
    echo "packager ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers && \
    chown -R packager:packager /go 

USER packager

CMD ["exec", "$@"]
