#!/usr/bin/env bash
set -euxo pipefail

APT_GET="$(which apt-get 2> /dev/null)"

if [[ -x "$APT_GET" ]]; then
    "$APT_GET" update
    # git is required by before_script
    "$APT_GET" -y install git
fi
