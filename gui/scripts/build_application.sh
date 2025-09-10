#!/usr/bin/env bash
set -euox pipefail

# build the application

# NOTE: Updating of the app version should happen before `scripts/env.sh`
# is sourced to export updated version

# update version info in pubspec.yaml
scripts/update_app_version.sh

# This cleans up the version updates made in `scripts/build_application.sh`
cleanup() {
  local file="pubspec.yaml"
  if [ -f "${file}.bak" ]; then
    mv -f "${file}.bak" "${file}"
    echo "Reverted changes to ${file}"
  fi
}
trap cleanup EXIT ERR INT TERM

source "${WORKDIR}/ci/env.sh"
source "${WORKDIR}/ci/archs.sh"

echo "Building on $(uname -m)"

BUILD_TYPE="${BUILD_TYPE:-debug}"
# convert build type to lower case
BUILD_TYPE="${BUILD_TYPE,,}"

# for release builds save the build symbols to diffent location to reduce app size
RELEASE_SYMBOLS=build/app/symbols/${NAME}_${VERSION}
rm -fr "${RELEASE_SYMBOLS}"

FLAGS=""
if [ "$BUILD_TYPE" == "release" ]; then
  echo "Save debug symbols into ${RELEASE_SYMBOLS}"
  FLAGS="--split-debug-info=${RELEASE_SYMBOLS}"
fi

echo "Building application for ${BUILD_TYPE}"
flutter clean
# shellcheck disable=SC2086
flutter build linux --"${BUILD_TYPE}" ${FLAGS}

OUTPUT_DIR="${WORKDIR}/bin/${ARCH}/gui"
mkdir -p "${OUTPUT_DIR}"
echo "Copying bundle to ${OUTPUT_DIR}"
cp -r "./build/linux/${ARCH}/${BUILD_TYPE}/bundle/"* "${OUTPUT_DIR}"

