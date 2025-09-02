#!/usr/bin/env bash
set -euox pipefail

if [ "$#" -ne 1 ]; then
    echo "missing build type: debug or release"
    exit 1
fi

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

source "scripts/env.sh"
source "scripts/archs.sh"

echo Building on $(uname -m)

# convert build type to lower case
BUILD_TYPE="${1,,}"

# for release builds save the build symbols to diffent location to reduce app size
RELEASE_SYMBOLS=build/app/symbols/${NAME}_${VERSION}
rm -fr "${RELEASE_SYMBOLS}"

FLAGS=""
if [ "$BUILD_TYPE" == "release" ]; then
    echo Save debug symbols into ${RELEASE_SYMBOLS}
    FLAGS="--split-debug-info=${RELEASE_SYMBOLS}"
fi

echo Building application for ${BUILD_TYPE}
flutter build linux --${BUILD_TYPE} ${FLAGS}
