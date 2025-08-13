FROM ubuntu:22.04

# Install Go (replace version as needed)
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
        python3 \
        python3-pip \
        libxml2-dev \
        libonig-dev \
        dpkg-dev \
        libxml2-dev \
        libonig-dev \
        libsqlite3-dev 

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

# Install Mage, golangci-lint, gosec
RUN go install github.com/magefile/mage@latest && \
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
go install github.com/securego/gosec/v2/cmd/gosec@latest
ENV PATH="/root/go/bin:${PATH}"
RUN mage -init

WORKDIR /workspaces/nordvpn-linux
