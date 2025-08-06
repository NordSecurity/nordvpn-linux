#!/usr/bin/env bash

set -euxo pipefail

# This needs to be in sync with TUNNELBLICK_TAG
OPENVPN_VERSION="2.6.12"
export OPENVPN_VERSION
OPENVPN_SHA256SUM="1c610fddeb686e34f1367c347e027e418e07523a10f4d8ce4a2c2af2f61a1929"
export OPENVPN_SHA256SUM

# Used to download the patches for OpenVPN obfuscation. This needs to be in sync with OPENVPN_VERSION.
TUNNELBLICK_TAG="v6.0beta09" # it is a beta tag, because no other non-beta still has version 2.6.12
export TUNNELBLICK_TAG
TUNNELBLICK_SHA256SUM="ea4e810e15c963a53fe3625cf37e078ed118b9a6879d92ce9a01c3395c9aad42"
export TUNNELBLICK_SHA256SUM

OPENSSL_VERSION="3.0.17"
export OPENSSL_VERSION
OPENSSL_SHA256SUM="dfdd77e4ea1b57ff3a6dbde6b0bdc3f31db5ac99e7fdd4eaf9e1fbb6ec2db8ce"
export OPENSSL_SHA256SUM

LZO_VERSION="2.10"
export LZO_VERSION
LZO_SHA256SUM="c0f892943208266f9b6543b3ae308fab6284c5c90e627931446fb49b4221a072"
export LZO_SHA256SUM
