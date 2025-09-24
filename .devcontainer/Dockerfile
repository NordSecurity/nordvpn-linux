FROM ubuntu:22.04

RUN pwd
RUN echo "$PWD"

ENV GO_VERSION=1.22.4

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    wget \
    git \
    curl \
    bash \
    bash-completion \
    build-essential \
    coreutils \
    ca-certificates \
    pkg-config \
    python3 \
    python3-pip \
    libxml2-dev \
    libonig-dev \
    dpkg-dev \
    libxml2-dev \
    libonig-dev \
    libsqlite3-dev \
    iproute2 \
    elfutils \
    gettext \
    sudo \
    unzip \
    snapd \
    docker.io

# Install Go
RUN wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

# Install Rust (via rustup)
RUN curl https://sh.rustup.rs -sSf | bash -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

# Enable Docker CLI for non-root user
# RUN usermod -aG docker vscode

# Install Mage, NFPM, golangci-lint, gosec, git-secrets
RUN go install github.com/magefile/mage@latest && \
    go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install github.com/securego/gosec/v2/cmd/gosec@latest && \
    go install github.com/google/go-licenses/v2@latest && \
    git clone https://github.com/awslabs/git-secrets.git && \
    cd git-secrets && make install

ENV PATH="/root/go/bin:${PATH}"
RUN mage -init

ARG WORKSPACE_DIR=/workspaces
WORKDIR ${WORKSPACE_DIR}
