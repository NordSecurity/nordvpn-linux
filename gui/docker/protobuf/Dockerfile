# Build image to convert protobufs to dart code
# Update this when flutter builder is updated or find a way to keep in sync
FROM dart:3.10.0

LABEL org.opencontainers.image.source="https://github.com/NordSecurity/nordvpn-linux/blob/main/gui/docker/protobuf/Dockerfile"
LABEL org.opencontainers.image.description="Generates the protobuf to dart code, e.g. ./docker/generate_docker_image.sh ghcr.io/nordsecurity/nordvpn-linux/protobuf_dart-3.10.0 1.0.0 linux/amd64 docker/protobuf/"

WORKDIR /code/app

RUN apt-get update && \
    apt-get -y --no-install-recommends upgrade && \
    apt-get -y --no-install-recommends install \
        git \
        bash \
        sudo \
        wget && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    dart --version && \
    # Install protoc
    wget --progress=dot:giga https://github.com/protocolbuffers/protobuf/releases/download/v33.1/protoc-33.1-linux-x86_64.zip -O /tmp/protoc.zip && \
    echo "f3340e28a83d1c637d8bafdeed92b9f7db6a384c26bca880a6e5217b40a4328b /tmp/protoc.zip" | sha256sum -c - && \
    unzip /tmp/protoc.zip -d /usr/ && rm /tmp/protoc.zip

ENV DOCKER_FILE="https://github.com/NordSecurity/nordvpn-linux-gui/docker/protobuf/Dockerfile"

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd -g ${GROUP_ID} generator && useradd -l -m -u ${USER_ID} -g generator generator

USER generator

ENV PATH="${PATH}:/home/generator/.pub-cache/bin"

RUN dart pub global activate protoc_plugin
