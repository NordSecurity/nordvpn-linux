#!/bin/bash

set -euxo pipefail

source "${WORKDIR}/ci/build_rust.sh"

mkdir -p "${WORKDIR}/build/foss"

clone_if_absent "https://github.com/NordSecurity/libtelio.git" "${LIBTELIO_VERSION}" "${WORKDIR}/build/foss"

rm -rf "${lib_root}/current"

# BYPASS_LLT_SECRETS is needed for libtelio builds
BYPASS_LLT_SECRETS=1 build_rust "${WORKDIR}/build/foss/libtelio"
link_so_files "libtelio"

populate_current_ver "${lib_root}/current" "${lib_root}/libtelio/${LIBTELIO_VERSION}" "libtelio.so"
