#!/bin/bash
set -euxo pipefail

source "${CI_PROJECT_DIR}/ci/env.sh"

# implemented it this way, because expansion did not work
binaries=(
  "${CI_PROJECT_DIR}/bin/deps/openvpn/${ARCH}/${OPENVPN_VERSION}/openvpn"
)

for binary in "${binaries[@]}"; do
  if [[ -f "${binary}" ]]; then
    hardening-check -q --nocfprotection "${binary}"
  fi
done

# Stack protection is forcibly disabled when building Go apps
# https://github.com/golang/go/blob/202a1a57064127c3f19d96df57b9f9586145e21c/src/runtime/cgo/cgo.go#L28
binaries=(
  "${CI_PROJECT_DIR}/bin/${ARCH}/nordvpnd"
  "${CI_PROJECT_DIR}/bin/${ARCH}/nordfileshared"
  "${CI_PROJECT_DIR}/bin/${ARCH}/nordvpn"
)

for binary in "${binaries[@]}"; do
  if [[ -f "${binary}" ]]; then
    hardening-check -q --nostackprotector --nocfprotection "${binary}"
  fi
done
