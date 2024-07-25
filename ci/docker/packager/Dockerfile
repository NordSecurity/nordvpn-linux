FROM ghcr.io/nordsecurity/nordvpn-linux/golang:ubuntu-20.04-20240725

# Need to build nfpm with patch; expeting sometime
# nfpm upstream will be fixed (issue submitted)
RUN cd && \
    apt-get update && \
    apt-get install -y --no-install-recommends file binutils && \
    wget --progress=dot:giga "https://github.com/go-task/task/releases/download/v3.34.1/task_linux_amd64.deb" -O /tmp/task.deb && \
    echo "c40af16a9bda90ffd55c50f72d8f5ad2e7a9a341dc1edf4154e42271463858ad /tmp/task.deb" | sha256sum -c - && \
    dpkg -i /tmp/task.deb && rm /tmp/task.deb && \
    git clone https://github.com/keliramu/nfpm.git && \
    cd nfpm && \
    git checkout ensure-reproducable-deb-new && \
    task setup && \
    task build && \
    strip ./nfpm && \
    file ./nfpm && \
    ls -lh ./nfpm && \
    ./nfpm --version

FROM ghcr.io/nordsecurity/nordvpn-linux/golang:ubuntu-20.04-20240725

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

# From first stage copy only nfpm binary
COPY --from=0 /root/nfpm/nfpm /usr/bin/nfpm

RUN apt-get update && \
    apt-get -y --no-install-recommends install elfutils gettext-base && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    # Install chglog
    go install github.com/goreleaser/chglog/cmd/chglog@v0.1.2 && \
    go clean -cache -modcache && \
    nfpm --version
