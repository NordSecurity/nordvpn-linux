#!/usr/bin/env bash
set -euxo pipefail

source "${WORKDIR}/ci/env.sh"

# snap package will have stripped binaries - same as deb/rpm
STRIP="$(which eu-strip 2>/dev/null)"
BASEDIR="bin/${ARCH}"
# shellcheck disable=SC2153
"${STRIP}" "${BASEDIR}"/nordvpnd
# shellcheck disable=SC2153
"${STRIP}" "${BASEDIR}"/nordvpn
# shellcheck disable=SC2153
"${STRIP}" "${BASEDIR}"/nordfileshare
# shellcheck disable=SC2153
"${STRIP}" "${BASEDIR}"/norduserd

# shellcheck disable=SC2153
"${STRIP}" "${WORKDIR}/bin/deps/openvpn/${ARCH}/latest/openvpn"

# build snap package
snapcraft --destructive-mode

# move snap package
mkdir -p "${WORKDIR}"/dist/app/snap
mv "${WORKDIR}"/*.snap "${WORKDIR}"/dist/app/snap/
