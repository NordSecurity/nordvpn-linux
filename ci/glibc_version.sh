#!/bin/bash
set -euxo pipefail

source "${CI_PROJECT_DIR}/ci/env.sh"

glibc_version=$1

# implemented it this way, because expansion did not work
binaries=(
  "${CI_PROJECT_DIR}/bin/${ARCH}/nordvpn"
  "${CI_PROJECT_DIR}/bin/${ARCH}/nordvpnd"
  "${CI_PROJECT_DIR}/bin/${ARCH}/nordfileshared"
  "${CI_PROJECT_DIR}/bin/deps/openvpn/${ARCH}/${OPENVPN_VERSION}/openvpn"
)

for binary in "${binaries[@]}"; do
  if [[ -f "${binary}" ]]; then
    go run "${CI_PROJECT_DIR}/cmd/checkelf/main.go" "${binary}" "${glibc_version}"
  fi
done
