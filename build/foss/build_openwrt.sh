#!/usr/bin/env bash
set -euox

source "${WORKDIR}/ci/archs.sh"
source "${WORKDIR}/ci/env.sh"

declare -A targets=(
  [amd64]=x86_64-unknown-linux-musl
  [aarch64]=aarch64-unknown-linux-musl
)

declare -A cc=(
  [amd64]="x86_64-openwrt-linux-musl-gcc"
  [aarch64]="aarch64-openwrt-linux-musl-gcc"
)

pushd "${WORKDIR}/build/foss"
for arch in "${ARCHS[@]}"; do
  target="${targets[$arch]}"
  compiler="${cc[$arch]}"
  TARGET_CC="${compiler}" cargo build --target "${target}" --release
  mkdir -p "${WORKDIR}/bin/deps/foss/${arch}"
  ln -frs "${WORKDIR}/build/foss/target/${target}/release" "${WORKDIR}/bin/deps/foss/${arch}/latest"
done
popd
