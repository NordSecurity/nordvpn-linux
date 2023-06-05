FROM ubuntu:18.04

# Rust installation borrowed from https://github.com/rust-lang/docker-rust/blob/47a11a42c5a321d688fae525eaa7f1cf9940198b/1.64.0/buster/slim/Dockerfile
ENV RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH=/usr/local/cargo/bin:$PATH \
    RUST_VERSION=1.64.0

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        gcc \
        libc6-dev \
        wget \
        ; \
    dpkgArch="$(dpkg --print-architecture)"; \
    case "${dpkgArch##*-}" in \
        amd64) rustArch='x86_64-unknown-linux-gnu'; rustupSha256='5cc9ffd1026e82e7fb2eec2121ad71f4b0f044e88bca39207b3f6b769aaa799c' ;; \
        armhf) rustArch='armv7-unknown-linux-gnueabihf'; rustupSha256='48c5ecfd1409da93164af20cf4ac2c6f00688b15eb6ba65047f654060c844d85' ;; \
        arm64) rustArch='aarch64-unknown-linux-gnu'; rustupSha256='e189948e396d47254103a49c987e7fb0e5dd8e34b200aa4481ecc4b8e41fb929' ;; \
        i386) rustArch='i686-unknown-linux-gnu'; rustupSha256='0e0be29c560ad958ba52fcf06b3ea04435cb3cd674fbe11ce7d954093b9504fd' ;; \
        *) echo >&2 "unsupported architecture: ${dpkgArch}"; exit 1 ;; \
    esac; \
    url="https://static.rust-lang.org/rustup/archive/1.25.1/${rustArch}/rustup-init"; \
    wget "$url"; \
    echo "${rustupSha256} *rustup-init" | sha256sum -c -; \
    chmod +x rustup-init; \
    ./rustup-init -y --no-modify-path --profile minimal --default-toolchain $RUST_VERSION --default-host ${rustArch}; \
    rm rustup-init; \
    chmod -R a+w $RUSTUP_HOME $CARGO_HOME; \
    rustup --version; \
    cargo --version; \
    rustc --version; \
    apt-get remove -y --auto-remove \
        wget \
        ; \
    rm -rf /var/lib/apt/lists/*;

RUN rustup target add \
        x86_64-unknown-linux-gnu \
        i686-unknown-linux-gnu \
        armv7-unknown-linux-gnueabihf \
        aarch64-unknown-linux-gnu \
        arm-unknown-linux-gnueabi

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

RUN pip3 install typing; \
	pip3 install requests==2.21.0

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd -g ${GROUP_ID} ruster && useradd -l -m -u ${USER_ID} -g ruster ruster && echo "ruster ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

USER ruster