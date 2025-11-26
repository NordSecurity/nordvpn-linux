# Build image to convert protobufs to dart code
# Update this when flutter builder is updated or find a way to keep in sync
FROM dart:3.8.1

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
    wget --progress=dot:giga https://github.com/protocolbuffers/protobuf/releases/download/v21.6/protoc-21.6-linux-x86_64.zip -O /tmp/protoc.zip && \
    echo "6a9fc36363a2d05d73fc363a46cd57d849068d33305db39f77daac8ba073e818 /tmp/protoc.zip" | sha256sum -c - && \
    unzip /tmp/protoc.zip -d /usr/ && rm /tmp/protoc.zip

ENV DOCKER_FILE="https://github.com/NordSecurity/nordvpn-linux-gui/docker/protobuf/Dockerfile"

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd -g ${GROUP_ID} generator && useradd -l -m -u ${USER_ID} -g generator generator

USER generator

ENV PATH="${PATH}:/home/generator/.pub-cache/bin"

RUN dart pub global activate protoc_plugin
