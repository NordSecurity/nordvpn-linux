#!/usr/bin/env bash
set -euxo pipefail

source "${WORKDIR}/ci/env.sh"

if [ "${ENVIRONMENT}" = "prod" ]; then
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
fi

# build snap package
snapcraft --destructive-mode

# move snap package
mkdir -p "${WORKDIR}"/dist/app/snap
mv "${WORKDIR}"/*.snap "${WORKDIR}"/dist/app/snap/
