#!/usr/bin/env bash
set -euxo pipefail

APT_GET="$(which apt-get 2> /dev/null)"

case "$ENVIRONMENT" in
"qa")
    urlpath='nordvpn-test'
    export REPOSITORY='nordvpn-test-debian'
    ;;
"prod")
    urlpath='nordvpn'
    export REPOSITORY='nordvpn-debian'
    ;;
*)
    echo "$0 called with unknown argument '$1'" >&2
    exit 1
    ;;
esac

FILENAME=${NAME}_${VERSION}_all.deb
# check if file exists in repo
STATUSCODE=$(curl -X HEAD -I -s -o /dev/null -w "%{http_code}" https://repo.nordvpn.com/deb/"${urlpath}"/debian/pool/main/"${FILENAME}")

case "$STATUSCODE" in
"200")
    exit 0
    ;;
"403"|"404")
    # we then need to upload
    if [[ -x "$APT_GET" ]]; then
        "$APT_GET" update
        "$APT_GET" -y install
    fi

    if [[ ! -f ${WORKDIR}/dist/repo/deb/${FILENAME} ]]; then
        echo "$WORKDIR/dist/repo/deb/$FILENAME not found"
        exit 1
    fi
    pulp-admin deb repo uploads deb --repo-id="${REPOSITORY}" --file="${FILENAME}"
    pulp-admin deb repo publish run --repo-id="${REPOSITORY}"
    ;;
*)
    echo "Got bad status code: $STATUSCODE"
    exit 1
    ;;
esac
