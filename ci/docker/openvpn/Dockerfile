# builder-openvpn-armel
# note: debian has armel support which we need to build openvpn 2.6.12+
# note: debian:11 satisfies current libc version requirements
FROM debian:11

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN mkdir -p /.cache && chmod -R 777 /.cache

RUN dpkg --add-architecture armel&& \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends upgrade && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
        iptables \
        iproute2 \
        wget \
        curl \
        ca-certificates \
        git \
        unzip \
        pkg-config \
        dpkg-dev \
        make \
        gcc \
        gcc-arm-linux-gnueabi \
        libc6-dev-armel-cross \
        libcap-ng-dev:armel \
        libcap-ng0:armel \
        libnl-genl-3-200 \
        libnl-genl-3-200:armel \
        libnl-genl-3-dev \
        libnl-genl-3-dev:armel \
        libssl-dev:armel \
        procps && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
