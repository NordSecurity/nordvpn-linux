#!/usr/bin/env bash
set -euxo

# if host does not have ip6table modules loaded, we must loaded it the docker
if [[ ! $(sudo ip6tables -S) ]]; then
    if [[ ! $(command -v modprobe) ]]; then
        sudo apt -y install kmod
    fi
    sudo modprobe ip6table_filter
fi

find "${CI_PROJECT_DIR}"/dist/app/deb -type f -name "*amd64.deb" \
	-exec sudo apt install -y "{}" +
