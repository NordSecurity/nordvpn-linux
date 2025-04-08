#!/bin/sh

# check for root access
SUDO=
if [ "$(id -u)" -ne 0 ]; then
    SUDO=$(command -v sudo 2> /dev/null)

    if [ ! -x "${SUDO}" ]; then
        echo "Error: Run this script as root"
        exit 1
    fi
fi

set -e
ARCH=$(uname -m)
BASE_URL=https://repo.nordvpn.com
KEY_BASE_URL=https://repo.nordvpn.com
KEY_PATH=/gpg/nordvpn_public.asc
REPO_PATH_DEB=/deb/nordvpn/debian
REPO_PATH_RPM=/yum/nordvpn/centos
RELEASE="stable main"
ASSUME_YES=false
APP_VERSION=
PACKAGE="nordvpn"
NOSECRT=false

# Parse command line arguments. Available arguments are:
# -n                Non-interactive mode. With this flag present, 'assume yes' or 
#                   'non-interactive' flags will be passed when installing packages.
# -b <url>          The base URL of the public key and repository locations.
# -k <path>         Path to the public key for the repository.
# -d <path|file>    Repository location for debian packages.
# -v <version>      Debian package version to use.
# -r <path|file>    Repository location for rpm packages.
# -p <package>      Package name to install: <nordvpn> or <nordvpn-release>
# -a <arch>         Architecture e.g. "noarch" for nordvpn-release rpm case
# -s                Do not do security checks: allow not signed repo and packages.
while getopts 'nb:k:d:r:v:p:a:s' opt
do
    case $opt in
        n) ASSUME_YES=true ;;
        b) BASE_URL=$OPTARG ;;
        k) KEY_PATH=$OPTARG ;;
        d) REPO_PATH_DEB=$OPTARG ;;
        r) REPO_PATH_RPM=$OPTARG ;;
        v) APP_VERSION=$OPTARG ;;
        p) PACKAGE=$OPTARG ;;
        a) ARCH=$OPTARG ;;
        s) NOSECRT=true ;;
        *) ;;
    esac
done

# Construct the paths to the package repository and its key
PUB_KEY=${KEY_BASE_URL}${KEY_PATH}
REPO_URL_DEB=${BASE_URL}${REPO_PATH_DEB}
REPO_URL_RPM=${BASE_URL}${REPO_PATH_RPM}

check_cmd() {
    command -v "$1" 2> /dev/null
}

get_install_opts_for_apt() {
    flags=$(get_install_opts_for "apt")
    RETVAL="$flags"
}

get_install_opts_for_yum() {
    flags=$(get_install_opts_for "yum")
    RETVAL="$flags"
}

get_install_opts_for_dnf() {
    flags=$(get_install_opts_for "dnf")
    RETVAL="$flags"
}

get_install_opts_for_zypper() {
    flags=$(get_install_opts_for "zypper")
    RETVAL="$flags"
}

get_install_opts_for() {
    if $ASSUME_YES; then
        case "$1" in
            zypper)
                echo "-n";;
            *)
                echo "-y";;
        esac
    fi
    echo ""
}

get_update_secrt_opts_for_apt() {
    RETVAL=""
    if $NOSECRT; then
        RETVAL="--allow-insecure-repositories"
    fi
}

get_install_secrt_opts_for_apt() {
    RETVAL=""
    if $NOSECRT; then
        RETVAL="--allow-unauthenticated"
    fi
}

# For any of the following distributions, these steps are performed:
# 1. Add the NordVPN repository key
# 2. Add the NordVPN repository
# 3. Install NordVPN

# Install NordVPN for Debian, Ubuntu, Elementary OS, and Linux Mint
# (with the apt-get package manager)
install_apt() {
    if check_cmd apt-get; then
        get_install_opts_for_apt
        install_opts="${RETVAL}"
        get_update_secrt_opts_for_apt
        update_secrt="${RETVAL}"
        get_install_secrt_opts_for_apt
        install_secrt="${RETVAL}"

        export DEBIAN_FRONTEND=noninteractive 

        # Ensure apt is set up to work with https sources
        ${SUDO} apt-get ${install_opts} ${update_secrt} update
        ${SUDO} apt-get ${install_opts} ${install_secrt} install apt-transport-https

        # Add the repository key with either wget or curl
        if check_cmd wget; then
            wget -qO - "${PUB_KEY}" | ${SUDO} tee /etc/apt/trusted.gpg.d/nordvpn_public.asc > /dev/null
        elif check_cmd curl; then
            curl -s "${PUB_KEY}" | ${SUDO} tee /etc/apt/trusted.gpg.d/nordvpn_public.asc > /dev/null
        else
            echo "Couldn't find wget or curl - one of them is needed to proceed with the installation"
            exit 1
        fi

        echo "deb ${REPO_URL_DEB} ${RELEASE}" | ${SUDO} tee /etc/apt/sources.list.d/nordvpn-app.list
        ${SUDO} apt-get ${install_opts} ${update_secrt} update
        if [ ! -z "$APP_VERSION" ]; then
            ${SUDO} apt-get ${install_opts} ${install_secrt} install "${PACKAGE}"="$APP_VERSION"
        else
            ${SUDO} apt-get ${install_opts} ${install_secrt} install "${PACKAGE}"
        fi
        exit
    fi
}

# Install NordVPN for RHEL and CentOS
# (with the yum package manager)
install_yum() {
    if check_cmd yum && check_cmd yum-config-manager; then
        get_install_opts_for_yum
        install_opts="${RETVAL}"

        repo="${REPO_URL_RPM}"
        if [ ! -f "${REPO_URL_RPM}" ]; then
            repo="${repo}/${ARCH}"
        fi

        ${SUDO} rpm -v --import "${PUB_KEY}"
        ${SUDO} yum-config-manager --add-repo "${repo}"
        if [ ! -z "${APP_VERSION}" ]; then
            ${SUDO} yum ${install_opts} install --nogpgcheck "${PACKAGE}"-"${APP_VERSION}"."${ARCH}"
        else
            ${SUDO} yum ${install_opts} install --nogpgcheck "${PACKAGE}"
        fi
        exit
    fi
}

# Install NordVPN for Fedora and QubesOS
# (with the dnf package manager)
install_dnf() {
    if check_cmd dnf5; then
        get_install_opts_for_dnf
        install_opts="${RETVAL}"
        
        repo="${REPO_URL_RPM}"
        if [ ! -f "${REPO_URL_RPM}" ]; then
            repo="${repo}/${ARCH}"
        fi

        ${SUDO} rpm -v --import "${PUB_KEY}"
        ${SUDO} dnf5 config-manager addrepo --id="nordvpn" --set=baseurl="${repo}" --set=enabled=1 --overwrite
        if [ ! -z "${APP_VERSION}" ]; then
            ${SUDO} dnf5 ${install_opts} install --nogpgcheck "${PACKAGE}"-"${APP_VERSION}"."${ARCH}"
        else
            ${SUDO} dnf5 ${install_opts} install --nogpgcheck "${PACKAGE}"
        fi
        exit
    fi
    if check_cmd dnf; then
        get_install_opts_for_dnf
        install_opts="${RETVAL}"
        
        repo="${REPO_URL_RPM}"
        if [ ! -f "${REPO_URL_RPM}" ]; then
            repo="${repo}/${ARCH}"
        fi

        ${SUDO} rpm -v --import "${PUB_KEY}"
        ${SUDO} dnf ${install_opts} install 'dnf-command(config-manager)'
        ${SUDO} dnf config-manager --add-repo "${repo}"
        if [ ! -z "${APP_VERSION}" ]; then
            ${SUDO} dnf ${install_opts} install --nogpgcheck "${PACKAGE}"-"${APP_VERSION}"."${ARCH}"
        else
            ${SUDO} dnf ${install_opts} install --nogpgcheck "${PACKAGE}"
        fi
        exit
    fi
}

# Install NordVPN for openSUSE
# (with the zypper package manager)
install_zypper() {
    if check_cmd zypper; then
        if ! check_cmd curl; then
            echo "Curl is needed to proceed with the installation"
            exit 1
        fi
        get_install_opts_for_zypper
        install_opts="${RETVAL}"
        
        ${SUDO} rpm -v --import "${PUB_KEY}"
        if [ -f "${REPO_URL_RPM}" ]; then
            ${SUDO} zypper addrepo -f "${REPO_URL_RPM}"
        else 
            ${SUDO} zypper addrepo -g -f "${REPO_URL_RPM}/${ARCH}" nordvpn
        fi
        ${SUDO} zypper ${install_opts} install -y "${PACKAGE}"
        exit
    fi
}

install_apt
install_yum
install_dnf
install_zypper

# None of the known package managers (apt, yum, dnf, zypper) are available
echo "Error: Couldn't identify the package manager"
exit 1
