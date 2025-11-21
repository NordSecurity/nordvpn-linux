#!/usr/bin/env bash
set -euxo pipefail

# store the path where the local repository for the application is
REPO_PATH="${REPO_DIR}"

case "$1" in
    centos) yum -y install yum-utils createrepo ;;
    fedora) dnf -y install dnf-plugins-core createrepo ;;
    opensuse) 
        zypper refresh && zypper -n install curl createrepo_c
        # with zypper the path needs to be updated to point to the nordvpn.repo file, inside the local repo folder
        REPO_PATH="${REPO_DIR}/nordvpn.repo"
    ;;
    *) echo "Can't recognize the OS" && exit 1 ;;
esac

mkdir -p "${REPO_DIR}/$(arch)" && cp -t "${REPO_DIR}/$(arch)" "${WORKDIR}/dist/app/rpm/"*".$(arch).rpm"
createrepo "${REPO_DIR}/$(arch)"
echo "[nordvpn]
name=nordvpn
baseurl=file:///$REPO_DIR/$(arch)
enabled=1
gpgcheck=0" | tee "${REPO_DIR}"/nordvpn.repo 
"${WORKDIR}"/test/qa/install.sh -n -b "" -r "${REPO_PATH}"
