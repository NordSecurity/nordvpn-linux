#!/bin/bash
set -euo pipefail

source "${WORKDIR}/ci/env.sh"

TUNNELBLICK_VERSION="v3.8.8"
OPENVPN_TARBALL_DIR="${WORKDIR}/build/openvpn/tarballs"
OPENVPN_URL="https://github.com/Tunnelblick/Tunnelblick/raw/${TUNNELBLICK_VERSION}/third_party/sources"
mkdir -p "${OPENVPN_TARBALL_DIR}"
pushd "${OPENVPN_TARBALL_DIR}"
	wget --quiet -nc "${OPENVPN_URL}/openssl-${OPENSSL_VERSION}.tar.gz"
	wget --quiet -nc "${OPENVPN_URL}/lzo-${LZO_VERSION}.tar.gz"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/openvpn-${OPENVPN_VERSION}.tar.gz"
popd

OPENVPN_PATCHES_DIR="${WORKDIR}/build/openvpn/patches"
mkdir -p "${OPENVPN_PATCHES_DIR}"
pushd "${OPENVPN_PATCHES_DIR}"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/patches/02-tunnelblick-openvpn_xorpatch-a.diff"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/patches/03-tunnelblick-openvpn_xorpatch-b.diff"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/patches/04-tunnelblick-openvpn_xorpatch-c.diff"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/patches/05-tunnelblick-openvpn_xorpatch-d.diff"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/patches/06-tunnelblick-openvpn_xorpatch-e.diff"
popd

LIBNORD_VERSION="0.5.1"
LIBNORD_ID="6385"

if [[ "${FEATURES}" == *internal* ]]; then
	"${WORKDIR}"/ci/download_from_remote.sh \
		-O nord -p "${LIBNORD_ID}" -v "${LIBNORD_VERSION}" ${ARCH:+-a ${ARCH}} libnord.a
fi
