FROM ubuntu:22.04

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

COPY requirements.txt /tmp/requirements.txt

# Install Debian packages
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install \
        # linux app
        apt-utils curl git iputils-ping sudo kmod systemd \
        # preinstall deps required by nordvpn
        libxml2 iproute2 iptables \
        # install wireguard tools for tests
        wireguard-tools \
        # install python for tests
        python3 python3-pip \
        # install tshark  \
        tshark \
        # changelog parsing
        dpkg-dev && \
    # make sure, that Docker does not hang during installation, when we get TUI screen
    yes yes | DEBIAN_FRONTEND=teletype dpkg-reconfigure wireshark-common && \
    # cleanup
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install Python packages for tests
RUN python3 -m pip install --no-cache-dir -r /tmp/requirements.txt && rm -rf /tmp/*

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd --system nordvpn && groupadd -g ${GROUP_ID} qa && useradd -l -m -u ${USER_ID} -g qa -G nordvpn qa && echo "qa ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
RUN usermod -a -G wireshark qa

USER qa
CMD ["exec", "$@"]
