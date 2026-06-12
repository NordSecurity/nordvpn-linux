FROM ubuntu:24.04

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

EXPOSE 22

RUN apt-get update && \
    apt-get install -y \
        openssh-server sudo iptables iputils-ping \
        iproute2 procps ca-certificates libxml2 libidn2-0 zlib1g \
        curl \
        # systemd-resolved provides resolvectl
        systemd systemd-resolved && \
    echo "root:root" | chpasswd && \
    echo PermitRootLogin yes >> /etc/ssh/sshd_config && \
    mkdir /run/sshd

ARG USER_ID=1000
ARG GROUP_ID=1000

# Ubuntu 24.04 base image ships with a default "ubuntu" user/group pre-occupying
# UID/GID 1000, which causes "groupadd: GID '1000' already exists" and
# "useradd: UID 1000 is not unique" errors. Remove the conflicting user/group
# first before creating qapeer with the desired UID/GID.
RUN (getent passwd ${USER_ID} | cut -d: -f1 | xargs -r userdel -r || true) \
    && (getent group ${GROUP_ID} | cut -d: -f1 | xargs -r groupdel || true) \
    && groupadd -g ${GROUP_ID} qapeer \
    && useradd -l -m -u ${USER_ID} -g qapeer qapeer \
    && echo "qapeer ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/99-qapeer \
    && chmod 0440 /etc/sudoers.d/99-qapeer \
    && visudo -cf /etc/sudoers.d/99-qapeer

USER qapeer

COPY ci/docker/qa-peer/entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
