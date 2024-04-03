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
    && DEBIAN_FRONTEND=noninteractive apt-get install -y tshark \
    # make sure, that Docker does not hang during installation, when we get TUI screen
    && yes yes | DEBIAN_FRONTEND=teletype dpkg-reconfigure wireshark-common \
    # install add-apt-repository, to easily add repos to apt
    && apt-get -y install software-properties-common \
    # ensure PPA/deb/apt version of Firefox is preferred & install Firefox
    && sudo add-apt-repository ppa:mozillateam/ppa \
    && echo 'Package: *\nPin: release o=LP-PPA-mozillateam\nPin-Priority: 1001' > /etc/apt/preferences.d/mozilla-firefox \
    && apt-get -y install firefox \
    # install jq
    && apt-get -y install jq \
    # cleanup
    && apt-get clean

# Install geckodriver for selenium tests
RUN url=$(curl -s https://api.github.com/repos/mozilla/geckodriver/releases/latest | jq -r \
    ' .assets[].browser_download_url | select(contains("linux64")) | select(endswith("tar.gz"))') \
    && curl -L -o geckodriver.tar.gz "$url" \
    && tar -xvzf geckodriver.tar.gz -C /usr/bin \
    && rm geckodriver.tar.gz

ARG USER_ID=1000
ARG GROUP_ID=1000

# yq is used by LATTE proxy
RUN curl -L -o /usr/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 \
    && chmod +x /usr/bin/yq 

RUN groupadd --system nordvpn && groupadd -g ${GROUP_ID} qa && useradd -l -m -u ${USER_ID} -g qa -G nordvpn qa && echo "qa ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
RUN usermod -a -G wireshark qa

USER qa
CMD ["exec", "$@"]
