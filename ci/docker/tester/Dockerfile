FROM ubuntu:22.04

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

COPY requirements.txt /tmp/requirements.txt

RUN apt-get update \
    # linux app
    && apt-get install -y apt-utils curl git iputils-ping sudo \
    # preinstall deps required by nordvpn
    && apt-get install -y libxml2 iproute2 iptables \
    # install wireguard tools for tests
    && apt-get install -y wireguard-tools \
    # install python for tests
    && apt-get install -y python3 python3-pip \
    # install python packages for tests
    && python3 -m pip install -r /tmp/requirements.txt \
    # install thsark
    DEBIAN_FRONTEND=noninteractive apt-get install -y tshark \
    # make sure, that Docker does not hang during installation, when we get TUI screen
    yes yes | DEBIAN_FRONTEND=teletype dpkg-reconfigure wireshark-common \
    # cleanup
    && apt-get clean

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd --system nordvpn && groupadd -g ${GROUP_ID} qa && useradd -l -m -u ${USER_ID} -g qa -G nordvpn qa && echo "qa ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
RUN usermod -a -G wireshark qa

USER qa

CMD ["exec", "$@"]
