FROM ghcr.io/nordsecurity/nordvpn-linux/tester:1.2.0

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

USER root

COPY requirements.txt /tmp/requirements.txt

# Install firefox and jq for selenium tests
RUN curl -L https://packages.mozilla.org/apt/repo-signing-key.gpg -o /etc/apt/keyrings/packages.mozilla.org.asc && \
    echo "deb [signed-by=/etc/apt/keyrings/packages.mozilla.org.asc] https://packages.mozilla.org/apt mozilla main" > /etc/apt/sources.list.d/mozilla.list && \
    echo 'Package: *\nPin: origin packages.mozilla.org\nPin-Priority: 1000' > /etc/apt/preferences.d/mozilla && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get -y --no-install-recommends install firefox jq && \
    # cleanup
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install selenium
RUN python3 -m pip install --no-cache-dir -r /tmp/requirements.txt && rm -rf /tmp/*

# Install geckodriver for selenium tests
RUN url=$(curl -s https://api.github.com/repos/mozilla/geckodriver/releases/latest | jq -r \
    ' .assets[].browser_download_url | select(contains("linux64")) | select(endswith("tar.gz"))') && \
    curl -L -o geckodriver.tar.gz "$url" && \
    tar -xvzf geckodriver.tar.gz -C /usr/bin && \
    rm geckodriver.tar.gz

USER qa
CMD ["exec", "$@"]
