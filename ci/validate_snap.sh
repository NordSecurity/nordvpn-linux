#!/usr/bin/env bash
set -euxo pipefail

# NOTE: this script should be run in systemd/snapd environment, non-root

source "${WORKDIR}/ci/snap_functions.sh"

source "${WORKDIR}"/ci/env.sh

# to get over: snap "snapd" has "auto-refresh" change in progress
echo "~~~set snap refresh on hold - for snapd"
sudo snap refresh --hold='720h' snapd


# snap contains binaries which are always stripped, same as deb/rpm
STRIPPED_STATUS=", stripped"

ARCH=$([ "${ARCH}" == "aarch64" ] && echo arm64 || echo "${ARCH}")

FILES=$(find "${WORKDIR}/dist/app/snap/" -type f -name "*_${ARCH}.snap")
echo "${FILES}"

for FILE in $FILES; do
    echo "~~~check1: install app snap should not fail; FILE: ${FILE}"
    sudo snap install --dangerous "${FILE}"
    break #only one file is expected
done

echo "~~~set snap refresh on hold - for nordvpn"
sudo snap refresh --hold='720h' nordvpn


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
echo "${file_info}" | grep "ELF 32-bit"
echo "${file_info}" | grep "ARM, EABI5"
;;
"aarch64")
echo "${file_info}" | grep "ELF 64-bit"
echo "${file_info}" | grep "ARM aarch64"
;;
"amd64")
echo "${file_info}" | grep "ELF 64-bit"
echo "${file_info}" | grep "x86-64"
;;
esac

echo "~~~check2.2: binary is stripped/not stripped"
echo "${file_info}" | grep "${STRIPPED_STATUS}"

echo "~~~fix permissions"
sudo groupadd nordvpn
sudo usermod -aG nordvpn "${USER}"

echo "~~~connect snap interfaces"
snap_connect_interfaces

echo "~~~restart snap nordvpn service"
sudo snap stop nordvpn
sudo snap start nordvpn

sleep 5


SERVICE_UNIT=snap.nordvpn.nordvpnd.service

if systemctl is-failed --quiet "${SERVICE_UNIT}"; then
    echo "~~~snap logs nordvpn"
    sudo snap logs -n=100 nordvpn

    echo "~~~journalctl ${SERVICE_UNIT}"
    sudo journalctl -n 100 -u "${SERVICE_UNIT}"
fi

echo "~~~info: nordvpnd service status"

systemctl status "${SERVICE_UNIT}"

echo "~~~check3: socket file: if file present -> service is started/running"
ls -la /var/snap/nordvpn/common/run/nordvpn/nordvpnd.sock

echo "~~~check4: minimal test"
nordvpn version
nordvpn status
nordvpn settings

echo "~~~DONE: SUCCESS!"
