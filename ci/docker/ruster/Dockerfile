FROM ubuntu:18.04

ARG GL_ACCESS_TOKEN
ARG SQLITE_DOWNLOAD_URL_PREFIX

# Rust installation borrowed from https://github.com/rust-lang/docker-rust/blob/47a11a42c5a321d688fae525eaa7f1cf9940198b/1.64.0/buster/slim/Dockerfile
ENV RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH=/usr/local/cargo/bin:$PATH \
    RUST_VERSION=1.77.2

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
  curl --header "PRIVATE-TOKEN: ${GL_ACCESS_TOKEN}" -o /usr/lib/i386-linux-gnu/libsqlite3.so -L ${SQLITE_DOWNLOAD_URL_PREFIX}/linux_i686_sqlite3.so; \
  curl --header "PRIVATE-TOKEN: ${GL_ACCESS_TOKEN}" -o /usr/lib/x86_64-linux-gnu/libsqlite3.so -L ${SQLITE_DOWNLOAD_URL_PREFIX}/linux_x86_64_sqlite3.so; \
  curl --header "PRIVATE-TOKEN: ${GL_ACCESS_TOKEN}" -o /usr/lib/arm-linux-gnueabihf/libsqlite3.so -L ${SQLITE_DOWNLOAD_URL_PREFIX}/linux_armv7_eabihf_sqlite3.so; \
  curl --header "PRIVATE-TOKEN: ${GL_ACCESS_TOKEN}" -o /usr/lib/arm-linux-gnueabi/libsqlite3.so -L ${SQLITE_DOWNLOAD_URL_PREFIX}/linux_armv5_eabi_sqlite3.so; \
  curl --header "PRIVATE-TOKEN: ${GL_ACCESS_TOKEN}" -o /usr/lib/aarch64-linux-gnu/libsqlite3.so -L ${SQLITE_DOWNLOAD_URL_PREFIX}/linux_aarch64_sqlite3.so; \
  apt-get remove -y --auto-remove curl

RUN echo "/usr/lib/i386-linux-gnu\n/usr/lib/x86_64-linux-gnu\n/usr/lib/arm-linux-gnueabihf\n/usr/lib/arm-linux-gnueabi\n/usr/lib/aarch64-linux-gnu" > /etc/ld.so.conf.d/sqlite3.conf

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
	libc6-dev-i386 \
	lib32gcc-7-dev; \
    rm -rf /var/lib/apt/lists/*;

# Installing protoc manually because package available in Ubuntu 18.04 does not contain
# files Google's .proto files which are supposed to be in `/usr/include` directory.
RUN set -eux; \
	apt-get update; \
	apt-get install -y curl unzip quilt python3-pip; \
	url="https://github.com/protocolbuffers/protobuf/releases/download/v21.12/protoc-21.12-linux-x86_64.zip"; \
	protocZipFileName="protoc-21.12-linux-x86_64.zip"; \
	protocSha256="3a4c1e5f2516c639d3079b1586e703fc7bcfa2136d58bda24d1d54f949c315e8"; \
	curl -L --output $protocZipFileName "$url"; \
	echo "$protocSha256 $protocZipFileName" | sha256sum --check --status; \
	unzip $protocZipFileName -d /usr; \
	rm $protocZipFileName; \
 	rm -rf /var/lib/apt/lists/*

RUN set -eux; \
  dpkg --add-architecture i386;\
  dpkg --add-architecture armhf;  \
  dpkg --add-architecture armel; \
  dpkg --add-architecture arm64; \
  ldconfig

RUN pip3 install typing; \
	pip3 install requests==2.21.0
