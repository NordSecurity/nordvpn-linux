#!/bin/bash
set -euo pipefail

source "${CI_PROJECT_DIR}/ci/env.sh"

OPENVPN_TARBALL_DIR="${CI_PROJECT_DIR}/build/openvpn/tarballs"
OPENVPN_URL="https://github.com/Tunnelblick/Tunnelblick/raw/${TUNNELBLICK_VERSION}/third_party/sources"
mkdir -p "${OPENVPN_TARBALL_DIR}"
pushd "${OPENVPN_TARBALL_DIR}"
	wget --quiet -nc "${OPENVPN_URL}/openssl-${OPENSSL_VERSION}.tar.gz"
	wget --quiet -nc "${OPENVPN_URL}/lzo-${LZO_VERSION}.tar.gz"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/openvpn-${OPENVPN_VERSION}.tar.gz"
popd

OPENVPN_PATCHES_DIR="${CI_PROJECT_DIR}/build/openvpn/patches"
mkdir -p "${OPENVPN_PATCHES_DIR}"
pushd "${OPENVPN_PATCHES_DIR}"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/patches/02-tunnelblick-openvpn_xorpatch-a.diff"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/patches/03-tunnelblick-openvpn_xorpatch-b.diff"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/patches/04-tunnelblick-openvpn_xorpatch-c.diff"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/patches/05-tunnelblick-openvpn_xorpatch-d.diff"
	wget --quiet -nc "${OPENVPN_URL}/openvpn/openvpn-${OPENVPN_VERSION}/patches/06-tunnelblick-openvpn_xorpatch-e.diff"
popd

LIBNORD_ID="6385"
NEXUS_ID="4226"

if [[ "${FEATURES}" == *internal* ]]; then
	"${CI_PROJECT_DIR}"/contrib/scripts/download_from_remote.sh -c "${NEXUS_ID}" -r qa \
		-n libnord.a -d nord -i "${LIBNORD_ID}" -v "${LIBNORD_VERSION}" ${ARCH:+-a ${ARCH}}
fi
