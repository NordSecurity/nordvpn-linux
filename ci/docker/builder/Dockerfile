FROM golang:1.20.1-buster

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

ENV GOBIN=/usr/bin

RUN dpkg --add-architecture i386; \
    dpkg --add-architecture arm64; \
    dpkg --add-architecture armhf; \
    dpkg --add-architecture armel; \
    apt-get update; \
    apt-get -y install \
        gcc-arm-linux-gnueabi \
        gcc-arm-linux-gnueabihf \
        gcc-aarch64-linux-gnu \
        gcc-i686-linux-gnu \
        libxml2-dev \
        libxml2-dev:i386 \
        libxml2-dev:arm64 \
        libxml2-dev:armhf \
        # libxml2-dev:armel \
        iptables; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*

# This is not a dangerous hack. `libxml2-dev:armel` package is broken on Debian 10. Therefore,
# this symlink fakes the existence of armel build. The expected symbols in `libxml.so` are no
# different than `armhf` and since real `libxml2.so` is loaded at runtime and package has a
# dependency for `libxml2`, a proper library is loaded
RUN ln -s /usr/lib/arm-linux-gnueabihf/libxml2.so /usr/lib/arm-linux-gnueabi/libxml2.so

# Download and install musl to PATH (/usr/bin)
# Because i386 app compiled with Zig segfaults when linked against glibc during go runtime setup
RUN wget "http://musl.cc/i686-linux-musl-cross.tgz" -O /tmp/musl.tar.gz \
    && tar xf /tmp/musl.tar.gz -C /usr/lib \
    && ln -s /usr/lib/i686-linux-musl-cross/bin/i686-linux-musl-gcc /usr/bin/i686-linux-musl-gcc

CMD ["exec", "$@"]
