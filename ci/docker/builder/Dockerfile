FROM ghcr.io/nordsecurity/nordvpn-linux/golang:ubuntu-20.04-20241203

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN mkdir -p /.cache && chmod -R 777 /.cache

RUN dpkg --add-architecture i386 && \
    dpkg --add-architecture arm64 && \
    dpkg --add-architecture armhf && \
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
        # libxml2-dev:armel \
        # Install libc6-dev-armel-cross instead of libxml2-dev:armel to provide libc6-dev for armel
        libc6-dev-armel-cross \
        procps && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# This is not a dangerous hack. `libxml2-dev:armel` package is broken on Debian 10. Therefore,
# this symlink fakes the existence of armel build. The expected symbols in `libxml.so` are no
# different than `armhf` and since real `libxml2.so` is loaded at runtime and package has a
# dependency for `libxml2`, a proper library is loaded
RUN ln -s /usr/lib/arm-linux-gnueabihf/libxml2.so /usr/lib/arm-linux-gnueabi/libxml2.so && \
    # Download and install musl to PATH (/usr/bin)
    # Because i386 app compiled with Zig segfaults when linked against glibc during go runtime setup
    wget --progress=dot:giga "https://musl.cc/i686-linux-musl-cross.tgz" -O /tmp/musl.tar.gz && \
    echo "5047afc68170a2910895db2dfa448227e71a984bfa2130a1bc946fd1015d722b80b15e4abf90c64300815aa84fe781cc8b8a72f10174f9dce96169e035911880 /tmp/musl.tar.gz" | sha512sum -c - && \
    tar -xzf /tmp/musl.tar.gz -C /usr/lib && rm /tmp/musl.tar.gz && \
    ln -s /usr/lib/i686-linux-musl-cross/bin/i686-linux-musl-gcc /usr/bin/i686-linux-musl-gcc && \
    # Install gocover-cobertura which is used for test coverage visualization in the diff view
    go install github.com/boumenot/gocover-cobertura@latest && \
    go clean -cache -modcache
