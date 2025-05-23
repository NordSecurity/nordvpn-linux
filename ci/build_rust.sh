#!/usr/bin/env bash
set -euxo pipefail

source "${WORKDIR}/ci/export_lib_versions.sh"
source "${WORKDIR}/ci/populate_current_lib_ver.sh"

lib_root="${WORKDIR}/bin/deps/lib/"

declare -A targets=(
  [amd64]=x86_64-unknown-linux-gnu
  [i386]=i686-unknown-linux-gnu
  [arm64]=aarch64-unknown-linux-gnu
  [armhf]=armv7-unknown-linux-gnueabihf
  [armel]=arm-unknown-linux-gnueabi
)

function clone_if_absent() {
  if [[ $# -ne 3 ]]; then
    echo "Three parameters are required for clone_if_absent function:"
    echo "clone_if_absent <repo_url> <tag/version> <destination_dir>"
    echo -e "\t repo_url - URL to the repository being cloned"
    echo -e "\t tag/version - repository tag\version you want to clone"
    echo -e "\t destination_dir - where to clone the repo to"
    echo "You passed ${#} arg(s): ${*}"
    exit 1
  fi
  repo_url=$1
  version=$2
  dst_arch_dir=$3
  repo_name="${repo_url##*/}"
  repo_name="${repo_name%.git}"

  pushd "${dst_arch_dir}"
  if [[ ! -d "${dst_arch_dir}/${repo_name}" ]]; then
    git clone "${repo_url}"
  fi

  pushd "${repo_name}"
  # In case checkout does not work, try to fetch the desired version and check it out again
  set +e
  git checkout "${version}" || git fetch origin "${version}"
  set -e
  git checkout "${version}"
  popd
  popd
}

function build_rust() {
  if [[ $# -ne 1 ]]; then
    echo "One parameter is required for build_rust function:"
    echo "build_rust <repo_root>"
    echo -e "\t repo_root - directory of the rust repository"
    echo "You passed ${#} arg(s): ${*}"
    exit 1
  fi
  repo_root=$1
  pushd "${repo_root}"
  for arch in "${ARCHS_RUST[@]}"; do
    target_arch="${targets[$arch]}"
    cargo build --target "${target_arch}" --release
  done
  popd
}

function link_so_files() {
  if [[ $# -ne 1 ]]; then
    echo "One parameter is required for link_so_files function:"
    echo "link_so_files <library_name>"
    echo -e "\t library_name - name of the library which so directory to link"
    echo "You passed ${#} arg(s): ${*}"
    exit 1
  fi
  library_name=$1
  for arch in "${ARCHS_RUST[@]}"; do
    target_arch="${targets[${arch}]}"
    lib_dir="${lib_root}/${library_name}"
    target_dir="${WORKDIR}/build/foss/${library_name}/target"
    mkdir -p "${lib_dir}"
    ln -sfnr "${target_dir}" "${lib_dir}/current"
    ln -sfn "${target_arch}/release" "${target_dir}/${arch}"
  done
}

mkdir -p "${WORKDIR}/build/foss"

# ====================[  Build libtelio from source ]=========================
clone_if_absent "https://github.com/NordSecurity/libtelio.git" "${LIBTELIO_VERSION}" "${WORKDIR}/build/foss"

rm -rf "${lib_root}/current"

# BYPASS_LLT_SECRETS is needed for libtelio builds
BYPASS_LLT_SECRETS=1 build_rust "${WORKDIR}/build/foss/libtelio"
link_so_files "libtelio"

# ====================[  Build libdrop from source ]==========================
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

populate_current_ver "${lib_root}/current" "${lib_root}/libtelio/${LIBTELIO_VERSION}" "libtelio.so"
populate_current_ver "${lib_root}/current" "${lib_root}/libdrop/${LIBDROP_VERSION}" "libnorddrop.so"
