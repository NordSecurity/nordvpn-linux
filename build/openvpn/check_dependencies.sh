#!/bin/bash
set -euo pipefail

source "${WORKDIR}/build/openvpn/env.sh"

openvpn_tarbal_dir="${WORKDIR}/build/openvpn/tarballs"

tunnelblick_version="v3.8.8"
tunnelblick_sha256sum="747aeed732f5303b408006d40736ef2ce276e0d99671b2110bb6b4f2ff7a52ca"

tunnelblick_url="https://github.com/Tunnelblick/Tunnelblick/raw/${tunnelblick_version}/third_party/sources"

openssl_ver_tr=$(echo "$OPENSSL_VERSION" | tr . _)

mkdir -p "${openvpn_tarbal_dir}"
pushd "${openvpn_tarbal_dir}"
	openvpn_tarbal="openvpn-${OPENVPN_VERSION}.tar.gz"
	openssl_tarbal="openssl-${OPENSSL_VERSION}.tar.gz"
	lzo_tarbal="lzo-${LZO_VERSION}.tar.gz"

	wget -nv -nc "https://swupdate.openvpn.org/community/releases/${openvpn_tarbal}"
	wget -nv -nc "https://github.com/openssl/openssl/releases/download/OpenSSL_${openssl_ver_tr}/${openssl_tarbal}"
	wget -nv -nc "https://www.oberhumer.com/opensource/lzo/download/${lzo_tarbal}"

	echo "${OPENVPN_SHA256SUM} ${openvpn_tarbal}" | sha256sum -c -
	echo "${OPENSSL_SHA256SUM} ${openssl_tarbal}" | sha256sum -c -
	echo "${LZO_SHA256SUM} ${lzo_tarbal}" | sha256sum -c -
popd

openvpn_patches_dir="${WORKDIR}/build/openvpn/patches"
mkdir -p "${openvpn_patches_dir}"
pushd "${openvpn_patches_dir}"
	wget -nv -nc "${tunnelblick_url}/openvpn/openvpn-${OPENVPN_VERSION}/patches/02-tunnelblick-openvpn_xorpatch-a.diff"
	wget -nv -nc "${tunnelblick_url}/openvpn/openvpn-${OPENVPN_VERSION}/patches/03-tunnelblick-openvpn_xorpatch-b.diff"
	wget -nv -nc "${tunnelblick_url}/openvpn/openvpn-${OPENVPN_VERSION}/patches/04-tunnelblick-openvpn_xorpatch-c.diff"
	wget -nv -nc "${tunnelblick_url}/openvpn/openvpn-${OPENVPN_VERSION}/patches/05-tunnelblick-openvpn_xorpatch-d.diff"
	wget -nv -nc "${tunnelblick_url}/openvpn/openvpn-${OPENVPN_VERSION}/patches/06-tunnelblick-openvpn_xorpatch-e.diff"
	[[ "$(sha256sum <<< "$(cat ./*diff)" | awk $'{print $1}')" == "${tunnelblick_sha256sum}" ]] || exit 1
popd
