#!/usr/bin/env bash
set -euo

export GOOS=linux
export CGO_ENABLED="1"
export SOCKET_DIR="/run/nordvpn"
# @TODO  where is this used ?
export SPK_NAME="NordVPN"

source "${WORKDIR}/ci/env.sh"
source "${WORKDIR}/ci/archs.sh"
