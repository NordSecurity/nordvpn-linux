#!/usr/bin/env bash
set -euxo pipefail

# configure git safe-directories when running in containerized environment
if [[ -n ${DOCKER_ENV+x} ]]; then
  git config --global --add safe.directory "${WORKDIR}"
  git config --global --add safe.directory "${WORKDIR}/parts/nordvpn/build"
fi

source "${WORKDIR}/ci/env.sh"

# snap package will have stripped binaries - same as deb/rpm
STRIP="$(which eu-strip 2>/dev/null)"
BASEDIR="bin/${ARCH}"

echo "Current dir: "
pwd

echo "Listing contents of ${BASEDIR}:"
ls -la "${BASEDIR}"

echo "Listing contents of ${WORKDIR}:"
ls -la "${WORKDIR}"

echo "Listing contents of ${WORKDIR}/gui/nordvpn-gui"
ls -la ${BASEDIR}/gui/nordvpn-gui

# shellcheck disable=SC2153
"${STRIP}" "${BASEDIR}"/nordvpnd
# shellcheck disable=SC2153
"${STRIP}" "${BASEDIR}"/nordvpn
# shellcheck disable=SC2153
"${STRIP}" "${BASEDIR}"/gui/bundle/nordvpn-gui
# shellcheck disable=SC2153
"${STRIP}" "${BASEDIR}"/nordfileshare
# shellcheck disable=SC2153
"${STRIP}" "${BASEDIR}"/norduserd

# shellcheck disable=SC2153
"${STRIP}" "${WORKDIR}/bin/deps/openvpn/current/${ARCH}/openvpn"

# Snap does not dereference symlinks on its own
# Avoid packaging errors in case of clean builds
dump_dir="${WORKDIR}/bin/deps/lib/current-dump"
mkdir -p "${WORKDIR}/bin/deps/lib/current/${ARCH}"
cp -rL "${WORKDIR}/bin/deps/lib/current" "${dump_dir}"
# Avoid missing dir errors in case of no libraries used
[ "$(ls -A "${dump_dir}/${ARCH}")" ] || touch "${dump_dir}/${ARCH}/empty"
trap 'rm -rf ${WORKDIR}/bin/deps/lib/current-dump' EXIT

# build snap package
snapcraft pack --destructive-mode

# move snap package
mkdir -p "${WORKDIR}"/dist/app/snap
mv "${WORKDIR}"/*.snap "${WORKDIR}"/dist/app/snap/
