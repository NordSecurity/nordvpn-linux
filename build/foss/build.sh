#!/usr/bin/env bash
set -euox

source "${CI_PROJECT_DIR}/ci/archs.sh"
source "${CI_PROJECT_DIR}/ci/env.sh"

declare -A targets=(
  [amd64]=x86_64-unknown-linux-gnu
  [aarch64]=aarch64-unknown-linux-gnu
  [i386]=i686-unknown-linux-gnu
  [armhf]=armv7-unknown-linux-gnueabihf
  [armel]=arm-unknown-linux-gnueabi
)

pushd "${CI_PROJECT_DIR}/build/foss"
for arch in "${ARCHS[@]}"; do
  target="${targets[$arch]}"
  cargo build --target "${target}" --release
  mkdir -p "${CI_PROJECT_DIR}/bin/deps/foss/${arch}"
  ln -frs "${CI_PROJECT_DIR}/build/foss/target/${target}/release" "${CI_PROJECT_DIR}/bin/deps/foss/${arch}/latest"
done
popd
