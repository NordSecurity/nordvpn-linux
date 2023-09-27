#!/usr/bin/env bash
set -euox pipefail

source "${WORKDIR}/ci/build_environment.sh"

PKG_TO_BUILD=$1
export PKG_HOMEPAGE="https://nordvpn.com/"
export PKG_MAINTAINER=${PKG_HOMEPAGE}
export PKG_VENDOR=${PKG_HOMEPAGE}
export PKG_DESCRIPTION="The best online VPN service for speed and security\nNordVPN protects your privacy online and\nlets access media content without regional restrictions.\nStrong encryption and no-log policy\nwith 5000+ servers in 60+ countries."

# clean build dir
APP_DIR=${WORKDIR}/dist/app
rm -rf "${APP_DIR}"

SYMBOL_DIR=${WORKDIR}/dist/symbols
mkdir -p "${SYMBOL_DIR}"/{deb,rpm} || true

# rpm package repositories have architecture in their names and those names sometimes
# do not match with architecture names on other distros
STRIP="$(which eu-strip 2>/dev/null)" # architecture does not matter for strip

"${WORKDIR}"/ci/check_dependencies.sh

# shellcheck disable=SC2153
export BASEDIR=${APP_DIR}/packages/${NAME}_${VERSION}_${ARCH}

# make build dirs
mkdir -p "${BASEDIR}"/usr/{bin,sbin}
mkdir -p "${BASEDIR}"/usr/share/man/man1

# shellcheck disable=SC2153
chmod +x "${WORKDIR}/bin/deps/openvpn/${ARCH}/${OPENVPN_VERSION}/openvpn"
"${STRIP}" "${WORKDIR}/bin/deps/openvpn/${ARCH}/${OPENVPN_VERSION}/openvpn"


export PKG_VERSION=${VERSION}

# extract symbols into files
# shellcheck disable=SC2153
"${STRIP}" -f "${SYMBOL_DIR}/${PKG_TO_BUILD}/nordvpnd-${ARCH}.debug" \
	"${WORKDIR}/bin/${ARCH}/nordvpnd"
# shellcheck disable=SC2153
"${STRIP}" -f "${SYMBOL_DIR}/${PKG_TO_BUILD}/nordvpn-${ARCH}.debug" \
	"${WORKDIR}/bin/${ARCH}/nordvpn"
# shellcheck disable=SC2153
	"${STRIP}" -f "${SYMBOL_DIR}/${PKG_TO_BUILD}/nordfileshared-${ARCH}.debug" \
		"${WORKDIR}/bin/${ARCH}/nordfileshared"

cp "${WORKDIR}/bin/${ARCH}/nordvpnd" "${BASEDIR}"/usr/sbin/nordvpnd
cp "${WORKDIR}/bin/${ARCH}/nordvpn" "${BASEDIR}"/usr/bin/nordvpn
cp "${WORKDIR}/bin/${ARCH}/nordfileshared" "${BASEDIR}"/usr/bin/nordfileshared
cd "${WORKDIR}"

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
