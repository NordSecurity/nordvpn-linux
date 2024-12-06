#!/usr/bin/env bash
set -euxo pipefail

source "${WORKDIR}/ci/snap_functions.sh"

# if host does not have ip6table modules loaded, we must loaded it the docker
if [[ ! $(sudo ip6tables -S) ]]; then
    if [[ ! $(command -v modprobe) ]]; then
        sudo apt -y install kmod
    fi
    sudo modprobe ip6table_filter
fi

echo "~~~REMOVE previous SNAP package"
sudo snap remove --purge nordvpn

echo "~~~INSTALL new SNAP package"
find "${WORKDIR}"/ -type f -name "*amd64.snap" \
	-exec sudo snap install --dangerous "{}" +

echo "~~~GRANT permissions - connect snap interfaces"
snap_connect_interfaces

echo "~~~INSTALL Snap DONE."
