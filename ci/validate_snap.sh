#!/usr/bin/env bash
set -euxo pipefail

source "${WORKDIR}"/ci/env.sh

STRIPPED_STATUS="with debug_info, not stripped"
if [ "${ENVIRONMENT}" = "prod" ]; then
    STRIPPED_STATUS=", stripped"
fi

echo "~~~check1: install app snap should not fail"
sudo snap install --dangerous "${WORKDIR}/dist/app/snap/*${ARCH}.snap"

echo "~~~running on host info: "
uname -a

echo "~~~check2: installed file exists in expected location"
nordvpn_file="/snap/nordvpn/current/bin/nordvpn"
file_info=$(file ${nordvpn_file})

echo "TARGET ARCH: ${ARCH}"
echo "${file_info}"

echo "~~~check2.1: binary is of expected architecture"
case "${ARCH}" in
"armhf")
echo "${file_info}" | grep "ELF 32-bit LSB pie executable, ARM, EABI5"
;;
"aarch64")
echo "${file_info}" | grep "ELF 64-bit LSB pie executable, ARM aarch64"
;;
"amd64")
echo "${file_info}" | grep "ELF 64-bit LSB pie executable, x86-64"
;;
esac

echo "~~~check2.2: binary is stripped/not stripped"
echo "${file_info}" | grep "${STRIPPED_STATUS}"

# give some time for service to start
sleep 5

echo "~~~info: nordvpnd service status"
systemctl status snap.nordvpn.nordvpnd.service

sleep 5

echo "~~~check3: socket file: if file present -> service is started/running"
ls -la /var/snap/nordvpn/common/run/nordvpn/nordvpnd.sock

echo "~~~fix permissions"
sudo groupadd nordvpn
sudo usermod -aG nordvpn "${USER}"
sudo snap connect nordvpn:network-control
sudo snap connect nordvpn:network-observe
sudo snap connect nordvpn:firewall-control
sudo snap connect nordvpn:login-session-observe
sudo snap connect nordvpn:system-observe

echo "~~~check4: minimal test"
nordvpn version
nordvpn status
nordvpn settings

echo "~~~DONE: SUCCESS!"
