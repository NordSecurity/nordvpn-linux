#!/usr/bin/env bash
set -euxo pipefail

exclude_self() {
  grep -P -v "github.com/NordSecurity/(?!systray)|moose"
}

mkdir -p "${WORKDIR}/dist"

# shellcheck disable=SC2046
go-licenses report $(go list -deps ./... | exclude_self) \
  --template "${WORKDIR}/ci/licenses.tpl" > "${WORKDIR}/dist/THIRD-PARTY-NOTICES.md"
