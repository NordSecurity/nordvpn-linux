#!/bin/bash
set -euox

source "${CI_PROJECT_DIR}"/ci/env.sh

"${CI_PROJECT_DIR}"/ci/check_dependencies.sh

go mod download

# shellcheck disable=SC2046
golangci-lint run -c "${CI_PROJECT_DIR}"/.golangci-lint.yml \
	$(go list ./... | grep -v events/moose  | sed 's/github.com\/NordSecurity\/nordvpn-linux\///g') \
	-v
