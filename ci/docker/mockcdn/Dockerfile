FROM python:3.12.10-slim-bookworm

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

COPY server.py /server.py
COPY configs /configs

RUN apt-get update && \
    apt-get upgrade -y && \
    DEBIAN_FRONTEND="noninteractive" apt-get -y install --no-install-recommends openssh-server iproute2 tree && \
    rm -rf /var/lib/apt/lists/* && \
    echo PermitRootLogin yes >> /etc/ssh/sshd_config && \
    mkdir /run/sshd && \
    rm -f /etc/ssh/ssh_host_*key*

RUN echo 'root:root' | chpasswd
RUN ssh-keygen -t ed25519 -f /etc/ssh/ssh_host_ed25519_key -N ''

EXPOSE 22
EXPOSE 8000

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]