#!/usr/bin/env bash
set -euxo pipefail

source "${CI_PROJECT_DIR}/ci/env.sh"

exclude_self() {
  grep -v "github.com/NordSecurity\|moose"
}

mkdir -p "${CI_PROJECT_DIR}/dist"

# shellcheck disable=SC2046
go-licenses report $(go list -deps ./... | exclude_self) \
  --template "${CI_PROJECT_DIR}/ci/licenses.tpl" > "${CI_PROJECT_DIR}/dist/THIRD-PARTY-NOTICES.md"
