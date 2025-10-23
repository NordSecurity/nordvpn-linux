#!/usr/bin/env bash
set -euxo pipefail

echo "=== OS ==="; cat /etc/os-release; dpkg --print-architecture
echo "=== Sources ==="; grep -Rhv '^#' /etc/apt/sources.list /etc/apt/sources.list.d/* || true
echo "=== Lists present? ==="; ls -lh /var/lib/apt/lists || true
echo "=== Before update: policy ==="; apt-cache policy xsltproc || true
echo "=== Update (IPv4) ==="; sudo apt-get update
echo "=== After update: policy ==="; apt-cache policy xsltproc || true

# if host does not have ip6table modules loaded, we must loaded it the docker
if [[ ! $(sudo ip6tables -S) ]]; then
    if [[ ! $(command -v modprobe) ]]; then
        sudo apt -y install kmod
    fi
    sudo modprobe ip6table_filter
fi

find "${WORKDIR}"/dist/app/deb -type f -name "*amd64.deb" \
	-exec sudo apt install -y "{}" +
