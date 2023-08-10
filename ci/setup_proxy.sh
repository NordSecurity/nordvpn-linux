#!/bin/bash
# --replay or --record
set -eux

args=$1

cd "$CI_PROJECT_DIR"/3rd-party/proxy || exit
./proxy.sh -r latte_config.yml "$args" &
pwd
cd "$CI_PROJECT_DIR" || exit
pwd
"$CI_PROJECT_DIR"/ci/install_deb.sh
sudo mv /var/lib/nordvpn/openvpn /var/lib/nordvpn/openvpn.bak
sudo cp "$CI_PROJECT_DIR"/3rd-party/proxy/pretend_openvpn.sh /var/lib/nordvpn/openvpn