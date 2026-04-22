#!/bin/bash

set -euxo pipefail

source "${WORKDIR}/ci/build_rust.sh"
source "${WORKDIR}/ci/export_lib_versions.sh"
source "${WORKDIR}/ci/populate_current_lib_ver.sh"

# Skip build if libtelio.so already exists for amd64
if [[ -f "${lib_root}/libtelio/current/amd64/libtelio.so" ]]; then
    echo "libtelio.so already exists, skipping build"
    exit 0
fi

mkdir -p "${WORKDIR}/build/foss"

clone_if_absent "https://github.com/NordSecurity/libtelio.git" "${LIBTELIO_VERSION}" "${WORKDIR}/build/foss"
rm -rf "${lib_root}/current"

pushd "${WORKDIR}/build/foss/libtelio"
rustup target add aarch64-unknown-linux-gnu \
    aarch64-unknown-linux-gnu \
    arm-unknown-linux-gnueabi \
    armv7-unknown-linux-gnueabihf \
    i686-unknown-linux-gnu
popd

# BYPASS_LLT_SECRETS is needed for libtelio builds
BYPASS_LLT_SECRETS=1 build_rust "${WORKDIR}/build/foss/libtelio"
link_so_files "libtelio"

populate_current_ver "${lib_root}/current" "${lib_root}/libtelio/current" "libtelio.so"
