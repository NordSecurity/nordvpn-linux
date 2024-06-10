FROM ubuntu:22.04

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

COPY requirements.txt /tmp/requirements.txt

RUN apt-get update && \
    apt-get -y --no-install-recommends install \
        # linux app
        apt-utils curl git iputils-ping sudo kmod \
        # preinstall deps required by nordvpn
        libxml2 iproute2 iptables \
        # install wireguard tools for tests
        wireguard-tools \
        # install python for tests
        python3 python3-pip && \
    # install python packages for tests
    python3 -m pip install --no-cache-dir -r /tmp/requirements.txt && \
    # install thsark
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install tshark && \
    # make sure, that Docker does not hang during installation, when we get TUI screen
    yes yes | DEBIAN_FRONTEND=teletype dpkg-reconfigure wireshark-common && \
    # cleanup
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/*

# TODO: Move into separate selenium image or install just during tests
# Install firefox and jq for selenium tests
RUN curl -L https://packages.mozilla.org/apt/repo-signing-key.gpg -o /etc/apt/keyrings/packages.mozilla.org.asc && \
    echo "deb [signed-by=/etc/apt/keyrings/packages.mozilla.org.asc] https://packages.mozilla.org/apt mozilla main" > /etc/apt/sources.list.d/mozilla.list && \
    echo 'Package: *\nPin: origin packages.mozilla.org\nPin-Priority: 1000' > /etc/apt/preferences.d/mozilla && \
    apt-get update && \
    apt-get -y --no-install-recommends install firefox jq && \
    # cleanup
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install geckodriver for selenium tests
RUN url=$(curl -s https://api.github.com/repos/mozilla/geckodriver/releases/latest | jq -r \
    ' .assets[].browser_download_url | select(contains("linux64")) | select(endswith("tar.gz"))') && \
    curl -L -o geckodriver.tar.gz "$url" && \
    tar -xvzf geckodriver.tar.gz -C /usr/bin && \
    rm geckodriver.tar.gz

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd --system nordvpn && groupadd -g ${GROUP_ID} qa && useradd -l -m -u ${USER_ID} -g qa -G nordvpn qa && echo "qa ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
RUN usermod -a -G wireshark qa

USER qa
CMD ["exec", "$@"]
