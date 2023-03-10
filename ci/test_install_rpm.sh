#!/usr/bin/env bash
set -euxo

case "$1" in
    centos) yum -y install yum-utils createrepo ;;
    fedora) dnf -y install dnf-plugins-core createrepo ;;
    opensuse) zypper -n install curl createrepo ;;
    *) echo "Can't recognise the OS" && exit 1 ;;
esac

mkdir -p "${REPO_DIR}/$(arch)" && cp -t "${REPO_DIR}/$(arch)" "${CI_PROJECT_DIR}/dist/app/rpm/*.$(arch).rpm"
createrepo "${REPO_DIR}/$(arch)"
echo "[nordvpn]
name=nordvpn
baseurl=file:///$REPO_DIR/$(arch)
enabled=1
gpgcheck=0" | tee "${REPO_DIR}"/nordvpn.repo 
"${CI_PROJECT_DIR}"/contrib/scripts/install.sh -n -b "" -k "https://repo.nordvpn.com/gpg/nordvpn_public.asc" -r "${REPO_DIR}/nordvpn.repo"
