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
"${STRIP}" "${WORKDIR}/bin/deps/openvpn/current/${ARCH}/openvpn"

mkdir -p "${WORKDIR}/bin/deps/current-dump"

# Snap does not dereference symlinks on its own
cp -rL "${WORKDIR}/bin/deps/lib/current" "${WORKDIR}/bin/deps/lib/current-dump"
trap 'rm -rf ${WORKDIR}/bin/deps/lib/current-dump' EXIT

# build snap package
snapcraft --destructive-mode

# move snap package
mkdir -p "${WORKDIR}"/dist/app/snap
mv "${WORKDIR}"/*.snap "${WORKDIR}"/dist/app/snap/
