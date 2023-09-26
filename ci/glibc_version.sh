#!/bin/bash
set -euxo pipefail

source "${WORKDIR}/ci/env.sh"

glibc_version=$1

# implemented it this way, because expansion did not work
binaries=(
  "${WORKDIR}/bin/${ARCH}/nordvpn"
  "${WORKDIR}/bin/${ARCH}/nordvpnd"
  "${WORKDIR}/bin/${ARCH}/nordfileshared"
  "${WORKDIR}/bin/deps/openvpn/${ARCH}/${OPENVPN_VERSION}/openvpn"
)

for binary in "${binaries[@]}"; do
  if [[ -f "${binary}" ]]; then
    go run "${WORKDIR}/cmd/checkelf/main.go" "${binary}" "${glibc_version}"
  fi
done
