#!/bin/bash
set -euxo pipefail

echo "Installing dependencies..."
sudo apt-get update
echo "wireshark-common wireshark-common/install-setuid boolean true" | sudo debconf-set-selections
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y \
    curl \
    dpkg-dev \
    git \
    iputils-ping \
    kmod \
    python3 \
    python3-pip \
    systemd \
    tshark \
    wireguard-tools

sudo usermod -aG wireshark vagrant

echo "Installing tester dependencies..."
pip3 install -r /vagrant/ci/docker/tester/requirements.txt || true