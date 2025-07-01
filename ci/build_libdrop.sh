#!/bin/bash

set -euxo pipefail

source "${WORKDIR}/ci/build_rust.sh"
source "${WORKDIR}/ci/export_lib_versions.sh"
source "${WORKDIR}/ci/populate_current_lib_ver.sh"

clone_if_absent "https://github.com/NordSecurity/libdrop.git" "${LIBDROP_VERSION}" "${WORKDIR}/build/foss"

# libdrop does not define configuration for linkers for different architectures
linkers_config=$(
  cat <<EOF
[target.x86_64-unknown-linux-gnu]
linker = "x86_64-linux-gnu-gcc"

[target.i686-unknown-linux-gnu]
linker = "i686-linux-gnu-gcc"

[target.aarch64-unknown-linux-gnu]
linker = "aarch64-linux-gnu-gcc"

[target.armv7-unknown-linux-gnueabihf]
linker = "arm-linux-gnueabihf-gcc"

[target.arm-unknown-linux-gnueabi]
linker = "arm-linux-gnueabi-gcc"
EOF
)
mkdir -p "${WORKDIR}/build/foss/libdrop/.cargo"
echo "${linkers_config}" >"${WORKDIR}/build/foss/libdrop/config.toml"

build_rust "${WORKDIR}/build/foss/libdrop"
link_so_files "libdrop"

populate_current_ver "${lib_root}/current" "${lib_root}/libdrop/current" "libnorddrop.so"
