#!/bin/bash
set -euxo pipefail

destination_directory="${WORKDIR}/bin/deps/lib"

pushd $destination_directory

LIBXML_VERSION="2.15"
LIBXML_VERSION_MINOR="1"
LIBXML_SHA256SUM="c008bac08fd5c7b4a87f7b8a71f283fa581d80d80ff8d2efd3b26224c39bc54c"

libxml_dir="libxml2-${LIBXML_VERSION}.${LIBXML_VERSION_MINOR}"
libxml_tarbal="${libxml_dir}.tar.xz"
wget -nv -nc "https://download.gnome.org/sources/libxml2/${LIBXML_VERSION}/${libxml_tarbal}"
echo "${LIBXML_SHA256SUM} ${libxml_tarbal}" | sha256sum -c -

tar --extract --xz --file $libxml_tarbal

pushd $libxml_dir
./configure
make

cp .libs/libxml2.so ../current

popd