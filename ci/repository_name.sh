#!/usr/bin/env bash
set -euxo

PACKAGE="${1}"
REPOSITORY=""

case "${ENVIRONMENT}" in
    "qa")
        REPOSITORY="nordvpn-test"
        ;;
    "prod")
        REPOSITORY="nordvpn"
        ;;
    *)
        echo "$0 called with unknown environment: ${ENVIRONMENT}" >&2
        exit 1
        ;;
esac

case "${PACKAGE}" in
    "deb")
        REPOSITORY="${REPOSITORY}-debian"
        ;;
    "rpm")
        REPOSITORY="${REPOSITORY}-centos"
        ;;
    *)
        echo "$0 called with unknown package: ${PACKAGE}" >&2
        exit 1
        ;;
esac

export REPOSITORY
