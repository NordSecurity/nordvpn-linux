FROM golang:1.20.1-bullseye

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN apt-get update -y \
    && apt-get -y install devscripts \
    && apt-get -y install curl sudo \
    && curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s v2.13.0 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd -g ${GROUP_ID} scanner && useradd -l -m -u ${USER_ID} -g scanner scanner && echo "scanner ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

USER scanner

CMD ["exec", "$@"]
