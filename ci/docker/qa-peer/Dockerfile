FROM ubuntu:22.04

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

EXPOSE 22

RUN apt-get update && \
    apt-get install -y openssh-server && \
    echo "root:root" | chpasswd && \
    echo PermitRootLogin yes >> /etc/ssh/sshd_config && \
    mkdir /run/sshd

CMD [ "/usr/sbin/sshd", "-D" ]
