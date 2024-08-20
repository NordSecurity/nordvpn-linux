#!/usr/bin/env bash
set -euxo pipefail

source "${WORKDIR}/ci/env.sh"

cd ${WORKDIR}

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

# translate arch id
TARGET_ARCH_4SNAP=$([ "$ARCH" == "aarch64" ] && echo arm64 || echo $ARCH)

# prepare snapcraft.yaml
cp ${WORKDIR}/snap/local/snapcraft.yaml.template ${WORKDIR}/snap/snapcraft.yaml

sed -i 's\TARGET_ARCH_4SNAP\'"${TARGET_ARCH_4SNAP}"'\g' ${WORKDIR}/snap/snapcraft.yaml
sed -i 's\TARGET_ARCH_4APP\'"${ARCH}"'\g' ${WORKDIR}/snap/snapcraft.yaml

# build snap package
snapcraft --destructive-mode

# move snap package
mkdir -p ${WORKDIR}/dist/snaps
mv ${WORKDIR}/*.snap ${WORKDIR}/dist/snaps/
