#!/usr/bin/env bash
set -euox

APT_GET="$(which apt-get 2> /dev/null)"

case "$ENVIRONMENT" in
"qa")
    urlpath='nordvpn-test'
    export REPOSITORY='nordvpn-test-centos'
    ;;
"prod")
    urlpath='nordvpn'
    export REPOSITORY='nordvpn-centos'
    ;;
*)
    echo "$0 called with unknown argument '$1'" >&2
    exit 1
    ;;
esac

FILENAME=${NAME}-${VERSION}.noarch.rpm
# check if file exists in repo
STATUSCODE=$(curl -X HEAD -I -s -o /dev/null -w "%{http_code}" https://repo.nordvpn.com/yum/"${urlpath}"/centos/noarch/Packages/n/"${FILENAME}")

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

    if [[ ! -f ${WORKDIR}/dist/repo/rpm/${FILENAME} ]]; then
        echo "$WORKDIR/dist/repo/rpm/$FILENAME not found"
        exit 1
    fi
    pulp-admin rpm repo uploads rpm --repo-id="${REPOSITORY}"-noarch --file="${FILENAME}"
    pulp-admin rpm repo publish run --repo-id="${REPOSITORY}"-noarch
    ;;
*)
    echo "Got bad status code: $STATUSCODE"
    exit 1
    ;;
esac
