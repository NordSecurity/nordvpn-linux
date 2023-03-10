#!/usr/bin/env bash
set -euo

export GOOS=linux
export CGO_ENABLED="1"
export SOCKET_DIR="/run/nordvpn"
# @TODO  where is this used ?
export SPK_NAME="NordVPN"

source "${CI_PROJECT_DIR}/ci/env.sh"
source "${CI_PROJECT_DIR}/ci/archs.sh"
