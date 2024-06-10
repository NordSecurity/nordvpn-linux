FROM ghcr.io/nordsecurity/nordvpn-linux/tester:1.2.0

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

ARG PROXY_URL
ARG PROXY_BRANCH="v1.0.2"

USER root

# Install openssh-client for git clone
RUN apt-get update && \
    apt-get -y --no-install-recommends install \
        openssh-client && \
    # cleanup
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# yq is used by LATTE proxy
RUN curl -L -o /usr/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && \
    chmod +x /usr/bin/yq

# Install LATTE proxy
RUN --mount=type=secret,id=gitlab_ssh_key GIT_SSH_COMMAND='ssh -i /run/secrets/gitlab_ssh_key -o IdentitiesOnly=yes -o StrictHostKeyChecking=no' git clone --depth 1 --branch $PROXY_BRANCH $PROXY_URL /tmp/proxy && \
    USER=qa /tmp/proxy/proxy.sh -i && \
    echo $PROXY_BRANCH > /opt/proxy/branch && \
    rm -rf /tmp/*

USER qa
CMD ["exec", "$@"]
