FROM ubuntu:22.04

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

EXPOSE 22

RUN apt-get update && \
    apt-get install -y openssh-server sudo && \
    echo "root:root" | chpasswd && \
    echo PermitRootLogin yes >> /etc/ssh/sshd_config && \
    mkdir /run/sshd

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd -g ${GROUP_ID} qapeer && useradd -l -m -u ${USER_ID} -g qapeer qapeer && echo "qapeer ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

USER qapeer

COPY ci/docker/qa-peer/entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
