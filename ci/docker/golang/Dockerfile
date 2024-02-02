FROM debian:buster-slim
# TODO: Change base image to ubuntu:20.04 on 2024-06-30
# Then to debian:bullseye-slim on 2025-05-29
# Then to ubuntu:22.04 on 2026-06-30
# Then to debian:bookworm-slim on 2027-06-01
# Then to ubuntu:24:04 on 2028-06-30

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

# Based on: https://github.com/docker-library/golang/blob/a6fd6eceb0cb26da2fceefb4353768c472f84420/1.21/bookworm/Dockerfile
ENV GOLANG_VERSION 1.21.6
ENV GOLANG_SHA256 3f934f40ac360b9c01f616a9aa1796d227d8b0328bf64cb045c7b8c4ee9caea4

RUN apt-get update && \
    apt-get -y --no-install-recommends upgrade && \
    apt-get -y --no-install-recommends install \
        iptables \
        iproute2 \
        wget \
        curl \
        ca-certificates \
        git && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    # Install Go from the official binaries
    wget --progress=dot:giga "https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz && \
    echo "$GOLANG_SHA256 /tmp/go.tar.gz" | sha256sum -c - && \
    tar -xzf /tmp/go.tar.gz -C /usr/local && rm /tmp/go.tar.gz

# Don't auto-upgrade the gotoolchain
# https://github.com/docker-library/golang/issues/472
ENV GOTOOLCHAIN=local
ENV GOBIN=/usr/bin
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" "$GOPATH/pkg/mod" && chmod -R 1777 "$GOPATH"
