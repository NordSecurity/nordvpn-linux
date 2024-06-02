#!/bin/bash
# --replay or --record
set -eux

args=$1

cd "$WORKDIR"/3rd-party/proxy || exit
./proxy.sh -r latte_config.yml "$args" &
pwd
cd "$WORKDIR" || exit
pwd
"$WORKDIR"/ci/install_deb.sh
sudo mv /usr/lib/nordvpn/openvpn /var/lib/nordvpn/openvpn.bak
sudo cp "$WORKDIR"/3rd-party/proxy/pretend_openvpn.sh /usr/lib/nordvpn/openvpn
