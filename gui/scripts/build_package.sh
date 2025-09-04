#!/usr/bin/env bash
set -euox pipefail

if [ "$#" -ne 3 ]; then
  echo -e "missing parameters:\n - build type: debug | release. \n - package type: deb | rpm \n - binaries architecture: amd64 | arm64"
  exit 1
fi

# TODO: Improve the way the version is updated. Currently, it needs to be
# done both here and in `build_application.sh`.

# NOTE: Updating of the app version should happen before `scripts/env.sh`
# is sourced to export updated version

# update version info in pubspec.yaml
scripts/update_app_version.sh

# This cleans up the version updates made in `pubspec.yaml`
cleanup() {
  local file="pubspec.yaml"
  if [ -f "${file}.bak" ]; then
    mv -f "${file}.bak" "${file}"
    echo "Reverted changes to ${file}"
  fi
}
trap cleanup EXIT ERR INT TERM

source "scripts/env.sh"
source "scripts/archs.sh"

# build type
BUILD_TYPE="${1,,}"
# package type
export PKG_TO_BUILD="${2,,}"
# binaries architecture
ARCH="${3,,}"

declare -A FLUTTER_FOLDER_NAME=(
  [arm64]=arm64
  [x86_64]=x64
  [amd64]=x64
)
ARCH_FOLDER_NAME=${FLUTTER_FOLDER_NAME[$ARCH]}

export PKG_HOMEPAGE="https://nordvpn.com/"
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

rm -fr "$DIST_DIR"
# build package structure
# the application expects to have the same "bundle" folder copied as it is somewhere
# splitting in different location into the system would break the application.
export INSTALL_DIR="/opt/${NAME}"
mkdir -p "${APP_BUNDLE_DIR}/${INSTALL_DIR}"
cp -r "build/linux/${ARCH_FOLDER_NAME}/${BUILD_TYPE}/bundle/"* "${APP_BUNDLE_DIR}${INSTALL_DIR}"

# generate changelog
readarray -d '' files < <(printf '%s\0' "contrib/changelog/prod/"*.md | sort -rzV)
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

# detect architecture type for packages
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

# create nfpm package description
envsubst <templates/nfpm_template.yaml >"${APP_BUNDLE_DIR}"/packages.yaml

# create desktop file
envsubst <templates/nordvpn-gui_template.desktop >"${APP_BUNDLE_DIR}/${NAME}.desktop"

# create install scripts
mkdir -p "${APP_BUNDLE_DIR}"/scriptlets/{deb,rpm}
envsubst <templates/scriptlets/deb/postinst_template >"${APP_BUNDLE_DIR}"/scriptlets/deb/postinst
envsubst <templates/scriptlets/deb/postrm_template >"${APP_BUNDLE_DIR}"/scriptlets/deb/postrm
envsubst <templates/scriptlets/rpm/post_template >"${APP_BUNDLE_DIR}"/scriptlets/rpm/post
envsubst <templates/scriptlets/rpm/postun_template >"${APP_BUNDLE_DIR}"/scriptlets/rpm/postun

# build package
OUT_PKG_DIR=${DIST_DIR}/${PKG_TO_BUILD}
echo "Build ${PKG_TO_BUILD} for ${ARCHS_DEB[$ARCH]} in ${OUT_PKG_DIR}"
mkdir -p "${OUT_PKG_DIR}"
nfpm pkg --packager "${PKG_TO_BUILD}" -f "${APP_BUNDLE_DIR}/packages.yaml" -t "${OUT_PKG_DIR}"
