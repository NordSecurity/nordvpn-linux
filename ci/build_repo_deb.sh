#!/usr/bin/env bash
set -euox

# clean build dir
rm -rf "${WORKDIR}"/dist/repo/deb

BASEDIR=${WORKDIR}/dist/repo/deb/${NAME}_${VERSION}_all

mkdir -p "${BASEDIR}"/{etc,DEBIAN}
mkdir -p "${BASEDIR}"/etc/apt/{sources.list.d,trusted.gpg.d}

cat <<EOF > "${BASEDIR}"/DEBIAN/conffiles
/etc/apt/sources.list.d/nordvpn.list
/etc/apt/trusted.gpg.d/nordvpn-keyring.gpg
EOF

# copy repo file
cp "${WORKDIR}"/contrib/repo/sources/deb/sources."${ENVIRONMENT}" "${BASEDIR}"/etc/apt/sources.list.d/nordvpn.list
# fetch key
wget -qO - https://repo.nordvpn.com/gpg/nordvpn_public.asc | gpg --dearmor > "${BASEDIR}"/etc/apt/trusted.gpg.d/nordvpn-keyring.gpg

# calculate weight and build the control template
WEIGHT=$(du -s "${BASEDIR}" | awk '{print $1}')

cat <<EOF > "${BASEDIR}"/DEBIAN/control
Package: nordvpn-release
Version: ${VERSION}
Depends: gnupg | gnupg2, apt-transport-https
Recommends: gpgv
Architecture: all
Maintainer: https://nordvpn.com/
Installed-Size: ${WEIGHT}
Section: misc
Priority: optional
Homepage: https://nordvpn.com/
Description: Package to install NordVPN GPG key and APT repo
EOF

# pack
dpkg-deb --build "${BASEDIR}"

# remove leftovers
rm -rf "${BASEDIR}"
