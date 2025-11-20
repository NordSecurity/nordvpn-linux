# Used to create the image for building a Flutter application
FROM ubuntu:20.04

LABEL org.opencontainers.image.source="https://github.com/NordSecurity/nordvpn-linux/blob/main/gui/docker/flutter/Dockerfile"
LABEL org.opencontainers.image.description="Builds the GUI. Must be build for arm64 and amd64. e.g. ./docker/generate_docker_image.sh ghcr.io/nordsecurity/nordvpn-linux/flutter-3.38.1 1.0.0 linux/amd64,linux/arm64 docker/flutter"

# this is needed for dpkg not to be interactive
ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get -y --no-install-recommends upgrade && \
    apt-get -y --no-install-recommends install \
        bash \
        ca-certificates \
        clang \
        cmake \
        git \
        libgtk-3-dev \
        ninja-build \
        pkg-config \
        sudo \
        unzip \
        curl \
        xvfb \
        xz-utils && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ARG USER_ID=1000
ARG GROUP_ID=1000
ARG USER_NAME=flutter_user
ARG HOME=/home/${USER_NAME}
ARG FLUTTER_VERSION=3.38.1

RUN groupadd -g ${GROUP_ID} ${USER_NAME} && useradd -l -m -u ${USER_ID} -g ${USER_NAME} ${USER_NAME}
USER ${USER_NAME}
WORKDIR ${HOME}

ENV PATH="${PATH}:${HOME}/.pub-cache/bin:${HOME}/flutter/bin"

# install flutter from source code. flutter doctor will fetch the needed files based on the platform ARM64 or x64
# create and build a project to ensure everything works
RUN git clone https://github.com/flutter/flutter.git && \
    cd flutter && \
    git checkout -f tags/${FLUTTER_VERSION} && \
    flutter doctor -v && \
    flutter config --no-enable-android --no-enable-windows-desktop --no-enable-macos-desktop --no-enable-ios --no-enable-fuchsia --no-enable-swift-package-manager && \
    rm -rf flutter/dev flutter/docs flutter/examples flutter/.github flutter/.vscode flutter/.idea && \
    rm -rf ${HOME}/flutter/bin/cache/artifacts/engine/android-* && \
    cd ~ && flutter create tst_build && cd tst_build && flutter build linux && cd .. && rm -fr tst_build && \
    flutter doctor
