FROM ubuntu:22.04

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

EXPOSE 22

RUN apt-get update && \
    apt-get install -y openssh-server sudo iptables iputils-ping \
        iproute2 procps ca-certificates libxml2 libidn2-0 zlib1g sockperf curl && \
    echo "root:root" | chpasswd && \
    # Adding license for sockperf below
    mkdir -p /usr/share/licenses/sockperf/ && \
    curl -s https://raw.githubusercontent.com/Mellanox/sockperf/sockperf_v2/copying > /usr/share/licenses/sockperf/LICENSE && \
    echo PermitRootLogin yes >> /etc/ssh/sshd_config && \
    mkdir /run/sshd

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd -g ${GROUP_ID} qapeer && useradd -l -m -u ${USER_ID} -g qapeer qapeer && echo "qapeer ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

USER qapeer

COPY ci/docker/qa-peer/entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
