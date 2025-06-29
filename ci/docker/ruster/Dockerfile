FROM debian:bullseye-slim

ARG SQLITE_DOWNLOAD_URL_PREFIX
ARG DEBIAN_FRONTEND=noninteractive

# Rust installation borrowed from https://github.com/rust-lang/docker-rust/blob/47a11a42c5a321d688fae525eaa7f1cf9940198b/1.64.0/buster/slim/Dockerfile
ENV RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH=/usr/local/cargo/bin:$PATH \
    RUST_VERSION=1.80.1

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        gcc \
        libc6-dev \
        wget \
        ; \
    dpkgArch="$(dpkg --print-architecture)"; \
    # https://static.rust-lang.org/rustup/archive/{rustup-version}/{target-triple}/rustup-init.sha256 
    case "${dpkgArch##*-}" in \
        amd64) rustArch='x86_64-unknown-linux-gnu'; rustupSha256='0b2f6c8f85a3d02fde2efc0ced4657869d73fccfce59defb4e8d29233116e6db' ;; \
        armhf) rustArch='armv7-unknown-linux-gnueabihf'; rustupSha256='f21c44b01678c645d8fbba1e55e4180a01ac5af2d38bcbd14aa665e0d96ed69a' ;; \
        arm64) rustArch='aarch64-unknown-linux-gnu'; rustupSha256='673e336c81c65e6b16dcdede33f4cc9ed0f08bde1dbe7a935f113605292dc800' ;; \
        i386) rustArch='i686-unknown-linux-gnu'; rustupSha256='e7b0f47557c1afcd86939b118cbcf7fb95a5d1d917bdd355157b63ca00fc4333' ;; \
        *) echo >&2 "unsupported architecture: ${dpkgArch}"; exit 1 ;; \
    esac; \
    url="https://static.rust-lang.org/rustup/archive/1.26.0/${rustArch}/rustup-init"; \
    wget "$url"; \
    echo "${rustupSha256} *rustup-init" | sha256sum -c -; \
    chmod +x rustup-init; \
    ./rustup-init -y --no-modify-path --profile minimal --default-toolchain $RUST_VERSION --default-host ${rustArch}; \
    rm rustup-init; \
    chmod -R a+w $RUSTUP_HOME $CARGO_HOME; \
    rustup --version; \
    cargo --version; \
    rustc --version

RUN set -eux; \
  rustup target add \
    x86_64-unknown-linux-gnu \
    i686-unknown-linux-gnu \
    armv7-unknown-linux-gnueabihf \
    aarch64-unknown-linux-gnu \
    arm-unknown-linux-gnueabi

RUN set -eux; \
  mkdir -p /usr/lib/i386-linux-gnu \
    /usr/lib/x86_64-linux-gnu \
    /usr/lib/arm-linux-gnueabihf \
    /usr/lib/arm-linux-gnueabi \
    /usr/lib/aarch64-linux-gnu

RUN set -eux; \
    dpkg --add-architecture i386; \
    apt-get update; \
    apt-get install -y \
    	# needed for cargo to fetch dependencies
    	git \
	gcc-aarch64-linux-gnu \
	gcc-arm-linux-gnueabi \
	gcc-arm-linux-gnueabihf \
	gcc-i686-linux-gnu \
	linux-libc-dev:i386 \
	libc6-dev-i386

RUN set -eux; \
    dpkg --add-architecture i386;\
    dpkg --add-architecture armhf; \
    dpkg --add-architecture armel; \
    dpkg --add-architecture arm64; \
    apt-get update; \
    apt-get install -y \
        libsqlite3-dev \
        libsqlite3-dev:i386 \
        libsqlite3-dev:armhf \
        libsqlite3-dev:arm64; \
    ldconfig; \
    rm -rf /var/lib/apt/lists/*

# This is not a dangerous hack. `libsqlite3-dev:armel` package is broken on Debian 11. Therefore,
# this symlink fakes the existence of armel build. The expected symbols in `libsqlite3.so` are no
# different than `armhf` and since real `libsqlite3.so` is loaded at runtime and package has a
# dependency for `libsqlite3-0`, a proper library is loaded
RUN ln -s /usr/lib/arm-linux-gnueabihf/libsqlite3.so /usr/lib/arm-linux-gnueabi/libsqlite3.so

ENV CARGO_TARGET_I686_UNKNOWN_LINUX_GNU_LINKER=i686-linux-gnu-gcc
ENV CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu
ENV CARGO_TARGET_ARM_UNKNOWN_LINUX_GNUEABI_LINKER=arm-linux-gnueabi-gcc
ENV CARGO_TARGET_ARMV7_UNKNOWN_LINUX_GNUEABIHF_LINKER=arm-linux-gnueabihf-gcc
