#!/usr/bin/env bash
set -euxo pipefail

# if host does not have ip6table modules loaded, we must loaded it the docker
if [[ ! $(sudo ip6tables -S) ]]; then
    if [[ ! $(command -v modprobe) ]]; then
        sudo apt -y install kmod
    fi
    sudo modprobe ip6table_filter
fi

sudo apt install -y "${WORKDIR}"/dist/app/deb/nordvpn_*_amd64.deb

# NOTE: we ignore dependencies here in the case of releasing breaking change
# (so GUI can depend on not-yet-released major version of daemon)
sudo dpkg -i --force-depends "${WORKDIR}"/dist/app/deb/nordvpn-gui_*_amd64.deb
