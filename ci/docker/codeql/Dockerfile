FROM ubuntu:24.04

ARG CODEQL_VERSION=v2.18.3
ARG GO_VERSION=1.22.10
ARG PB_VERSION=3.20.1

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    pkg-config \
    libxml2-dev \
    curl \
    unzip \
    python3 \
    python3-yaml \
    libsqlite3-dev \
    python3-termcolor \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install CodeQL CLI
RUN curl -L -o codeql-linux64.zip https://github.com/github/codeql-cli-binaries/releases/download/${CODEQL_VERSION}/codeql-linux64.zip \
    && unzip codeql-linux64.zip \
    && mv codeql /usr/local/codeql-home \
    && ln -s /usr/local/codeql-home/codeql /usr/local/bin/codeql \
    && rm codeql-linux64.zip

# Install Go
RUN curl -L -o go${GO_VERSION}.linux-amd64.tar.gz https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && rm -rf /usr/local/go && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

# Install protoc
RUN curl -L -o protoc.zip "https://github.com/protocolbuffers/protobuf/releases/download/v${PB_VERSION}/protoc-${PB_VERSION}-linux-x86_64.zip" && \
    unzip -o protoc.zip -d /usr/local && \
    rm protoc.zip

ENV PATH=$PATH:/usr/local/bin:/usr/local/go/bin

# Download standard query packs and place them in a system-wide location
ENV CODEQL_SEARCH_PATH=/usr/local/share/codeql-packs
RUN mkdir -p /tmp/codeql-packs && \
    codeql pack download --dir /tmp/codeql-packs codeql/go-queries && \
    mkdir -p $CODEQL_SEARCH_PATH && \
    mv /tmp/codeql-packs/codeql/go-queries/* $CODEQL_SEARCH_PATH/ && \
    rm -rf /tmp/codeql-packs

# Mark the workspace as safe for git operations
RUN git config --global --add safe.directory /opt
