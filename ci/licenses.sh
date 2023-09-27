#!/usr/bin/env bash
set -euxo pipefail

source "${WORKDIR}/ci/env.sh"

exclude_self() {
  grep -v "github.com/NordSecurity\|moose"
}

mkdir -p "${WORKDIR}/dist"

# shellcheck disable=SC2046
go-licenses report $(go list -deps ./... | exclude_self) \
  --template "${WORKDIR}/ci/licenses.tpl" > "${WORKDIR}/dist/THIRD-PARTY-NOTICES.md"
