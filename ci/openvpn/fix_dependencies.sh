#!/bin/bash
set -euo pipefail

# Need to install package libcap-ng-dev for current arch,
# cannot have this package installed for all archs - each
# install of this package replaces previous one.
# This workaround became needed for openvpn 2.6.12.
# NOTE: this fix is needed only when building for different
# (other than amd64) architectures on CI.

ARCH=$([ "${ARCH}" == "aarch64" ] && echo arm64 || echo "${ARCH}")

echo "Check what version of libncap-ng is present:"
find /usr -name 'libcap-ng*'
apt list --installed | grep -i libcap-ng

echo "Install libcap-ng for ARCH: ${ARCH}"

apt-get update 
apt-get install -y libcap-ng-dev:"${ARCH}" libcap-ng0:"${ARCH}"

echo "Check what version of libncap-ng is present after all:"
find /usr -name 'libcap-ng*'
apt list --installed | grep -i libcap-ng
