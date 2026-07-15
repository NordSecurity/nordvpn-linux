FROM ghcr.io/nordsecurity/nordvpn-linux/golang:ubuntu22.04-20260709@sha256:b26ebd8e021c721dba2ce57b7d12dd269eab038dcd36df9a7e4ee239e340a3fb

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN mkdir -p /.cache && chmod -R 777 /.cache

RUN dpkg --add-architecture i386 && \
    dpkg --add-architecture arm64 && \
    dpkg --add-architecture armhf && \
        # Point foreign architectures to ports.ubuntu.com
    sed -i 's/^deb /deb [arch=amd64,i386] /' /etc/apt/sources.list && \
    echo "deb [arch=arm64,armhf] http://ports.ubuntu.com/ubuntu-ports jammy main restricted universe multiverse" >> /etc/apt/sources.list && \
    echo "deb [arch=arm64,armhf] http://ports.ubuntu.com/ubuntu-ports jammy-updates main restricted universe multiverse" >> /etc/apt/sources.list && \
    echo "deb [arch=arm64,armhf] http://ports.ubuntu.com/ubuntu-ports jammy-security main restricted universe multiverse" >> /etc/apt/sources.list && \
    echo "deb [arch=arm64,armhf] http://ports.ubuntu.com/ubuntu-ports jammy-backports main restricted universe multiverse" >> /etc/apt/sources.list && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
        pkg-config \
        dpkg-dev \
        make \
        gcc \
        gcc-arm-linux-gnueabi \
        gcc-arm-linux-gnueabihf \
        gcc-aarch64-linux-gnu \
        gcc-i686-linux-gnu \
        libxml2-dev \
        libxml2-dev:i386 \
        libxml2-dev:arm64 \
        libxml2-dev:armhf \
        libcap-ng-dev \
        libcap-ng0:i386 \
        libcap-ng0:arm64 \
        libcap-ng0:armhf \
        libnl-genl-3-200 \
        libnl-genl-3-dev \
        libnl-genl-3-200:i386 \
        libnl-genl-3-dev:i386 \
        libnl-genl-3-200:arm64 \
        libnl-genl-3-dev:arm64 \
        libnl-genl-3-200:armhf \
        libnl-genl-3-dev:armhf \
        libsqlite3-dev \
        libsqlite3-dev:armhf \
        libsqlite3-dev:arm64 \
        libsqlite3-dev:i386 \
        # libxml2-dev:armel \
        # Install libc6-dev-armel-cross instead of libxml2-dev:armel to provide libc6-dev for armel
        libc6-dev-armel-cross \
        procps \
        # Python and file for arch sanity check post build
        python3 \
        nftables \
        conntrack \
        file && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# This is not a dangerous hack. `libxml2-dev:armel` package is broken on Debian 10. Therefore,
# this symlink fakes the existence of armel build. The expected symbols in `libxml.so` are no
# different than `armhf` and since real `libxml2.so` is loaded at runtime and package has a
# dependency for `libxml2`, a proper library is loaded
RUN ln -s /usr/lib/arm-linux-gnueabihf/libxml2.so /usr/lib/arm-linux-gnueabi/libxml2.so; \
    ln -s /usr/lib/arm-linux-gnueabihf/libsqlite3.so /usr/lib/arm-linux-gnueabi/libsqlite3.so; \
    ldconfig
    # Install gocover-cobertura which is used for test coverage visualization in the diff view
RUN go install github.com/boumenot/gocover-cobertura@latest && \
    go clean -cache -modcache
