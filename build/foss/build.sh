#!/usr/bin/env bash
set -euox

source "${WORKDIR}/ci/archs.sh"

declare -A targets=(
  [amd64]=x86_64-unknown-linux-gnu
  [aarch64]=aarch64-unknown-linux-gnu
  [i386]=i686-unknown-linux-gnu
  [armhf]=armv7-unknown-linux-gnueabihf
  [armel]=arm-unknown-linux-gnueabi
)

pushd "${WORKDIR}/build/foss"
for arch in "${ARCHS[@]}"; do
  target="${targets[$arch]}"
  cargo build --target "${target}" --release
  mkdir -p "${WORKDIR}/bin/deps/foss/${arch}"
  ln -frs "${WORKDIR}/build/foss/target/${target}/release" "${WORKDIR}/bin/deps/foss/${arch}/latest"
done
popd
