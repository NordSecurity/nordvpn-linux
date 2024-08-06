#!/usr/bin/env bash
set -euxo pipefail

src_dir=$(dirname -- ${BASH_SOURCE[0]})
source "${src_dir}/env.sh"

cores=$(nproc)

current_dir="${WORKDIR}/build/openvpn"
sources="${current_dir}/src"
tarballs="${current_dir}/tarballs"
output_dir="${WORKDIR}/bin/deps/openvpn/${ARCH}/${OPENVPN_VERSION}"
latest_dir="${WORKDIR}/bin/deps/openvpn/${ARCH}/latest"

patch_sources() {
  mkdir -p "${sources}"
  tar -xzf "${tarballs}/openssl-${OPENSSL_VERSION}.tar.gz" -C "${sources}"
  tar -xzf "${tarballs}/lzo-${LZO_VERSION}.tar.gz" -C "${sources}"
  tar -xzf "${tarballs}/openvpn-${OPENVPN_VERSION}.tar.gz" -C "${sources}"

  pushd "${sources}/openvpn-${OPENVPN_VERSION}"
    git apply ../../patches/02-tunnelblick-openvpn_xorpatch-a.diff
    git apply ../../patches/03-tunnelblick-openvpn_xorpatch-b.diff
    git apply ../../patches/04-tunnelblick-openvpn_xorpatch-c.diff
    git apply ../../patches/05-tunnelblick-openvpn_xorpatch-d.diff
    git apply ../../patches/06-tunnelblick-openvpn_xorpatch-e.diff
  popd
}
patch_sources

configure_openssl() {
  local compiler="${1}"
  ./Configure CC="${compiler}" \
    gcc no-asm --prefix="${current_dir}/openssl" -static -no-shared
}

configure_lzo() {
  local compiler="${1}"
  local target="${2}"
  local cflags="${3}"
  local ldflags="${4}"
  ./configure CC="${compiler}" \
    CFLAGS="${cflags}" \
    LDFLAGS="${ldflags}" \
    --prefix="${current_dir}/lzo" --host="${target}" --enable-static
}

configure_openvpn() {
  local compiler="${1}"
  local target="${2}"
  local cflags="${3}"
  local ldflags="${4}"
  ./configure CC="${compiler}" \
    CFLAGS="${cflags}" \
    LDFLAGS="${ldflags}" \
    OPENSSL_CFLAGS="-I${current_dir}/openssl/include" \
    LZO_CFLAGS="-I${current_dir}/lzo/include" \
    LIBS="-L${current_dir}/openssl/lib -L${current_dir}/lzo/lib -lssl -lcrypto -llzo2" \
    --prefix="${current_dir}/openvpn" --host="${target}" \
    --enable-static=yes --enable-iproute2 --disable-shared --disable-debug --disable-plugins
}

declare -A cross_compiler_map=(
    [i386]=i686-linux-gnu-gcc
    [amd64]=x86_64-linux-gnu-gcc
    [armel]=arm-linux-gnueabi-gcc
    [armhf]=arm-linux-gnueabihf-gcc
    [aarch64]=aarch64-linux-gnu-gcc
)

declare -A cross_compiler_map_openwrt=(
    [amd64]="x86_64-openwrt-linux-musl-gcc"
    [aarch64]="aarch64-openwrt-linux-musl-gcc"
)

pushd "${current_dir}"
  target=""
  compiler=""
  openssl_cflags=""
  openssl_ldflags=""
  lzo_cflags="-g -O2"
  lzo_ldflags=""
  openvpn_cflags="-Wall -Wno-unused-parameter -Wno-unused-function -g -O2 -D_FORTIFY_SOURCE=2 -std=c99 -fstack-protector"
  openvpn_ldflags="-Wl,-z,relro,-z,now -Wl,--as-needed"

  if [[ "${OS}" == "openwrt" ]]; then
    compiler="${cross_compiler_map_openwrt[${ARCH}]}"
    case "${ARCH}" in
      "amd64")
        target="x86_64-openwrt-linux-musl"
      ;;
      "aarch64")
        target="aarch64-openwrt-linux-musl"
      ;;
    esac
    prefix="$target-"
  else
    compiler="${cross_compiler_map[${ARCH}]}"
    case "${ARCH}" in
      "i386")
        target="i686-linux-gnu"
        prefix="$target-"
        openssl_cflags+=" -m32"
        openssl_ldflags+=" -m32"
        lzo_cflags+=" -m32"
        lzo_ldflags+=" -m32"
        openvpn_cflags+=" -m32"
        openvpn_ldflags+=" -m32"
      ;;
      "amd64")
        target="x86_64-linux-gnu"
        prefix="$target-"
      ;;
      "armel")
        target="arm-linux-gnueabi"
        prefix="$target-"
      ;;
      "armhf")
        target="arm-linux-gnueabihf"
        prefix="$target-"
      ;;
      "aarch64")
        target="aarch64-linux-gnu"
        prefix="$target-"
      ;;
    esac
  fi

  pushd "${sources}/openssl-${OPENSSL_VERSION}"
    configure_openssl "${compiler}"
    make -j$cores CFLAGS+="$openssl_cflags" LDFLAGS+="$openssl_ldflags" > /dev/null
    make install -j$cores CFLAGS+="$openssl_cflags" LDFLAGS+="$openssl_ldflags" > /dev/null
  popd

  pushd "${sources}/lzo-${LZO_VERSION}"
    configure_lzo "${compiler}" "${target}" "${lzo_cflags}" "${lzo_ldflags}"
    make -j$cores > /dev/null
    make install -j$cores > /dev/null
  popd

  pushd "${sources}/openvpn-${OPENVPN_VERSION}"
    configure_openvpn "${compiler}" "${target}" "${openvpn_cflags}" "${openvpn_ldflags}"
    make -j$cores > /dev/null
    make install -j$cores /dev/null
  popd
popd

mkdir -p "${output_dir}"
mv "${current_dir}/openvpn/sbin/openvpn" "${output_dir}"
ln -fsnr "${output_dir}" "${latest_dir}" 

rm -rf "${sources}"
rm -rf "${current_dir}/openssl"
rm -rf "${current_dir}/lzo"
rm -rf "${current_dir}/openvpn"
