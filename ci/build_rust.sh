#!/usr/bin/env bash

source "${WORKDIR}/ci/archs.sh"

lib_root="${WORKDIR}/bin/deps/lib"

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
  for arch in "${ARCHS[@]}"; do
    rust_arch="${ARCHS_RUST[$arch]}"
    target_arch="${targets[$rust_arch]}"
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
  for arch in "${ARCHS[@]}"; do
    rust_arch="${ARCHS_RUST[$arch]}"
    target_arch="${targets[$rust_arch]}"
    lib_dir="${lib_root}/${library_name}"
    target_dir="${WORKDIR}/build/foss/${library_name}/target"
    mkdir -p "${lib_dir}"
    ln -sfnr "${target_dir}" "${lib_dir}/current"
    ln -sfn "${target_arch}/release" "${target_dir}/${arch}"
  done
}
