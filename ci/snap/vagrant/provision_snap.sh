#!/usr/bin/env bash

set -euxo pipefail

# used to provision the vagrant VM for snap

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

echo "Wait for snap to finish setting up:"
until sudo snap wait system seed.loaded; do
    echo -n "."
    sleep 1
done

echo "Installing tester dependencies..."
sudo pip3 install -r /vagrant/ci/docker/tester/requirements.txt

