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

#TODO: go to gui directory, and simply run nfpm from there?
cleanup() {
  local file="${WORKDIR}/gui/pubspec.yaml"
  if [ -f "${file}.bak" ]; then
    mv -f "${file}.bak" "${file}"
    echo "Reverted changes to ${file}"
  fi
}
trap cleanup EXIT ERR INT TERM


gui/scripts/update_app_version.sh
source "gui/scripts/env.sh"
# NAME=nordvpn-gui
# export NAME

export PKG_DESCRIPTION="The NordVPN app for Linux now offers a visual interface for effortless online security. NordVPN protects your internet traffic with top-grade encryption and changes your IP address, so whatever you do online stays private and secure. Connect to over 7,100 high-speed servers covering 118 countries.\n\nYou can secure up to 10 devices with a single account. Enjoy a safer internet experience on all your devices."
export PKG_VERSION=${VERSION}

# variables used into the package installation scripts
export SUCCESS_INSTALL_MESSAGE="NordVPN GUI for Linux successfully installed!"
export INSTALL_SCRIPT="
# create symbolic link for the GUI executable
ln -s /opt/${NAME}/${NAME} /usr/bin/${NAME}
chmod +x /usr/bin/${NAME}
"

# used in package uninstall scripts
export UNINSTALL_SCRIPT="
# remove symbolic link for the GUI executable
rm -fr /usr/bin/${NAME}
"

# prepare folders structure for packaging
DIST_DIR="dist"
export APP_BUNDLE_DIR="$DIST_DIR/source/${NAME}_${VERSION}_${ARCH}"
# export APP_BUNDLE_DIR="bin/${ARCH}/gui"

# rm -fr "$DIST_DIR"

export INSTALL_DIR="/opt/${NAME}"
mkdir -p "${APP_BUNDLE_DIR}/${INSTALL_DIR}"
cp -r "${WORKDIR}/bin/${ARCH}/gui/"* "${APP_BUNDLE_DIR}${INSTALL_DIR}"


# generate changelog
readarray -d '' files < <(printf '%s\0' "gui/contrib/changelog/prod/"*.md | sort -rzV)
for filename in "${files[@]}"; do
  entry_name=$(basename "${filename}" .md)
  entry_tag=${entry_name%_*}
  entry_date=$(stat -c "%Y" "${filename}")

  printf "\055 semver: %s
  date: %s
  packager: NordVPN Linux Team <linux@nordvpn.com>
  deb:
    urgency: medium
    distributions:
      - stable
  changes:" \
    "${entry_tag}" "$(date -d@"${entry_date}" +%Y-%m-%dT%H:%M:%SZ)" >>"${DIST_DIR}"/changelog.yml

  while read -r line || [ -n "$line" ]; do
    printf "\n   - note: |-\n      %s" "${line:1}" >>"${DIST_DIR}"/changelog.yml
  done <"${filename}"

  printf "\n\n" >>"${DIST_DIR}"/changelog.yml
done

# create nfpm package description
# envsubst <gui/templates/nfpm_template.yaml >"${WORKDIR}"/gui/packages.yaml
envsubst <gui/templates/nfpm_template.yaml >"${APP_BUNDLE_DIR}"/packages.yaml

# create desktop file
envsubst <gui/templates/nordvpn-gui_template.desktop >"${APP_BUNDLE_DIR}/${NAME}.desktop"

# create install scripts
mkdir -p "${APP_BUNDLE_DIR}"/scriptlets/{deb,rpm}
envsubst <gui/templates/scriptlets/deb/postinst_template >"${APP_BUNDLE_DIR}"/scriptlets/deb/postinst
envsubst <gui/templates/scriptlets/deb/postrm_template >"${APP_BUNDLE_DIR}"/scriptlets/deb/postrm
envsubst <gui/templates/scriptlets/rpm/post_template >"${APP_BUNDLE_DIR}"/scriptlets/rpm/post
envsubst <gui/templates/scriptlets/rpm/postun_template >"${APP_BUNDLE_DIR}"/scriptlets/rpm/postun

OUT_PKG_DIR="${DIST_DIR}/${PKG_TO_BUILD}/gui"
echo "Build ${PKG_TO_BUILD} for ${ARCHS_DEB[$ARCH]} in ${OUT_PKG_DIR}"
mkdir -p "${OUT_PKG_DIR}"
# nfpm pkg --packager "${PKG_TO_BUILD}" -f "${WORKDIR}/gui/packages.yaml" -t "${APP_DIR}/${PKG_TO_BUILD}"
nfpm pkg --packager "${PKG_TO_BUILD}" -f "${APP_BUNDLE_DIR}/packages.yaml" -t "${APP_DIR}/${PKG_TO_BUILD}"

# mv "${OUT_PKG_DIR}"/*."${PKG_TO_BUILD}" "

# remove leftovers
rm -rf "${BASEDIR}"
rm -rf "${APP_BUNDLE_DIR}"
