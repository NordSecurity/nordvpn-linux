#!/usr/bin/env bash
set -euox pipefail

source "${WORKDIR}/ci/env.sh"
source "${WORKDIR}/ci/archs.sh"
source "${WORKDIR}/ci/openvpn/env.sh"

PKG_TO_BUILD=$1
export PKG_HOMEPAGE="https://nordvpn.com/"
export PKG_DESCRIPTION="The NordVPN app for Linux protects your internet traffic with top-grade encryption and changes your IP address, so whatever you do online stays private and secure. Connect to over 7,100 high-speed servers covering 118 countries.\n\nYou can secure up to 10 devices with a single account. Enjoy a safer internet experience on all your devices."

# clean build dir
APP_DIR=${WORKDIR}/dist/app
rm -rf "${APP_DIR}"

SYMBOL_DIR=${WORKDIR}/dist/symbols
mkdir -p "${SYMBOL_DIR}"/{deb,rpm} || true

# rpm package repositories have architecture in their names and those names sometimes
# do not match with architecture names on other distros
STRIP="$(which eu-strip 2>/dev/null)" # architecture does not matter for strip

# shellcheck disable=SC2153
export BASEDIR=${APP_DIR}/packages/${NAME}_${VERSION}_${ARCH}

# make build dirs
mkdir -p "${BASEDIR}"/usr/{bin,sbin}
mkdir -p "${BASEDIR}"/usr/lib/${NAME}
mkdir -p "${BASEDIR}"/usr/share/man/man1

# shellcheck disable=SC2153
chmod +x "${WORKDIR}/bin/deps/openvpn/current/${ARCH}/openvpn"
"${STRIP}" "${WORKDIR}/bin/deps/openvpn/current/${ARCH}/openvpn"

export PKG_VERSION=${VERSION}

cp "${WORKDIR}/bin/${ARCH}/nordvpnd" "${BASEDIR}"/usr/sbin/nordvpnd
cp "${WORKDIR}/bin/${ARCH}/nordvpn" "${BASEDIR}"/usr/bin/nordvpn
cp "${WORKDIR}/bin/${ARCH}/nordfileshare" "${BASEDIR}"/usr/lib/${NAME}/nordfileshare
cp "${WORKDIR}/bin/${ARCH}/norduserd" "${BASEDIR}"/usr/lib/${NAME}/norduserd

# nfpm does not dereference symlinks on its own
# Avoid packaging errors in case of clean builds
mkdir -p "${WORKDIR}/bin/deps/lib/current/${ARCH}"
cp -rL "${WORKDIR}/bin/deps/lib/current" "${WORKDIR}/bin/deps/lib/current-dump"
trap 'rm -rf ${WORKDIR}/bin/deps/lib/current-dump' EXIT

cd "${WORKDIR}"

# extract symbols into files
# shellcheck disable=SC2153
# modify binaries in the target directory
"${STRIP}" -f "${SYMBOL_DIR}/${PKG_TO_BUILD}/nordvpnd-${ARCH}.debug" \
	"${BASEDIR}"/usr/sbin/nordvpnd
# shellcheck disable=SC2153
"${STRIP}" -f "${SYMBOL_DIR}/${PKG_TO_BUILD}/nordvpn-${ARCH}.debug" \
	"${BASEDIR}"/usr/bin/nordvpn
# shellcheck disable=SC2153
"${STRIP}" -f "${SYMBOL_DIR}/${PKG_TO_BUILD}/nordfileshare-${ARCH}.debug" \
	"${BASEDIR}"/usr/lib/${NAME}/nordfileshare
# shellcheck disable=SC2153
"${STRIP}" -f "${SYMBOL_DIR}/${PKG_TO_BUILD}/norduserd-${ARCH}.debug" \
	"${BASEDIR}"/usr/lib/${NAME}/norduserd

# pack
case "$PKG_TO_BUILD" in
"deb")
	# shellcheck disable=SC2153
	export PKG_ARCH=${ARCHS_DEB[$ARCH]}
	;;
"rpm")
	# shellcheck disable=SC2153
	export PKG_ARCH=${ARCHS_RPM[$ARCH]}
	;;
*)
	echo "unknown package type ${PKG_TO_BUILD}"
	exit 1
	;;
esac


envsubst <"${WORKDIR}"/ci/nfpm/template.yaml >"${BASEDIR}"/packages.yaml
mkdir -p "${APP_DIR}/${PKG_TO_BUILD}"
nfpm pkg --packager "${PKG_TO_BUILD}" -f "${BASEDIR}"/packages.yaml
mv "${WORKDIR}"/*."${PKG_TO_BUILD}" "${APP_DIR}/${PKG_TO_BUILD}"

# remove leftovers
rm -rf "${BASEDIR}"

# Only build GUI package if the architecture supports Flutter
if [[ -n "${ARCHS_FLUTTER[$ARCH]:-}" ]]; then
  echo "Building GUI package for Flutter-supported architecture: $ARCH"
  
  cleanup() {
    local file="${WORKDIR}/gui/pubspec.yaml"
    if [ -f "${file}.bak" ]; then
      mv -f "${file}.bak" "${file}"
      echo "Reverted changes to ${file}"
    fi
  }
  trap cleanup EXIT ERR INT TERM

  # Build GUI package using the existing GUI build script
  # Set auxiliary environment variables for the GUI script
  export SKIP_DIST_CLEAN="true"  # Don't clean dist dir since we're integrating
  export CUSTOM_SOURCE_DIR="${WORKDIR}/bin/${ARCH}/gui"  # Use our built GUI binaries

  # Set the parameters the sourced script expects
  set -- release "${PKG_TO_BUILD}" "${ARCH}"

  # Source the GUI build script (this will run in current directory context)
  (cd gui && source scripts/build_package.sh)

  # Move the generated GUI package to the expected location
  mv gui/dist/${PKG_TO_BUILD}/*.${PKG_TO_BUILD} "${APP_DIR}/${PKG_TO_BUILD}/"
  
else
  echo "Skipping GUI package build - architecture $ARCH not supported by Flutter"
fi
