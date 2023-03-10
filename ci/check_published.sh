#!/usr/bin/env bash
set -euo

source "${CI_PROJECT_DIR}/ci/archs.sh"

EXIT_CODE=0

check_status() {
    package=$1
    arch=$2
    url=
    case "$package" in
        "deb")
            url=https://repo.nordvpn.com/deb/nordvpn/debian/pool/main/${NAME}_${VERSION}-${REVISION}_${arch}.deb
            ;;
        "rpm")
            url=https://repo.nordvpn.com/yum/nordvpn/centos/${arch}/Packages/n/${NAME}-${VERSION}-${REVISION}.${arch}.rpm
            ;;
    esac

    statuscode=$(curl -X HEAD -I -s -o /dev/null -w "%{http_code}" "${url}")
    case $statuscode in
        200)
            ;;
        40*)
            echo "${package} ${arch} not published"
            EXIT_CODE=1
            ;;
        *)
            echo "Got bad status code for ${package}: ${statuscode}"
            EXIT_CODE=1
            ;;
    esac
}

for arch in "${ARCHS[@]}"; do
    check_status "deb" "${ARCHS_DEB[$arch]}"
    check_status "rpm" "${ARCHS_RPM[$arch]}"
done

exit $EXIT_CODE
