#!/bin/bash
set -euxo pipefail

# implemented it this way, because expansion did not work
binaries=(
  "${WORKDIR}/bin/deps/openvpn/${ARCH}/latest/openvpn"
)

for binary in "${binaries[@]}"; do
  if [[ -f "${binary}" ]]; then
    hardening-check -q --nocfprotection "${binary}"
  fi
done

# Stack protection is forcibly disabled when building Go apps
# https://github.com/golang/go/blob/202a1a57064127c3f19d96df57b9f9586145e21c/src/runtime/cgo/cgo.go#L28
binaries=(
  "${WORKDIR}/bin/${ARCH}/nordvpnd"
  "${WORKDIR}/bin/${ARCH}/nordfileshare"
  "${WORKDIR}/bin/${ARCH}/nordvpn"
  "${WORKDIR}/bin/${ARCH}/norduserd"
)

for binary in "${binaries[@]}"; do
  if [[ -f "${binary}" ]]; then
    hardening-check -q --nostackprotector --nocfprotection "${binary}"
  fi
done

# GUI
GUI_DIR="${WORKDIR}/bin/${ARCH}/gui"
if [[ -d "${GUI_DIR}" ]]; then
  while IFS= read -r -d '' binary; do
    description=$(file -b "${binary}")
    if [[ "${description}" != ELF* ]]; then
      continue
    fi

    case "${binary##*/}" in
    # prebuilt Flutter engine, uses libc functions that cannot be fortified
    libflutter_linux_gtk.so) flags=(--nofortify) ;;
    # contains only Dart compiled code and data
    libapp.so) flags=(--nostackprotector --norelro --nobindnow) ;;
    *) flags=() ;;
    esac
    hardening-check -q --nocfprotection "${flags[@]}" "${binary}"

    runpath=$(readelf -dW "${binary}" | sed -nE 's/.*\(R(UN)?PATH\).*\[(.*)\]/\2/p')
    IFS=':' read -ra entries <<<"${runpath}"
    for entry in "${entries[@]}"; do
      # shellcheck disable=SC2016
      if [[ "${entry}" != '$ORIGIN' && "${entry}" != '$ORIGIN/'* ]]; then
        echo "${binary}: contains run-time search path '${entry}'"
        exit 1
      fi
    done

    if [[ "${description}" == *"not stripped"* ]]; then
      echo "${binary}: is not stripped"
      exit 1
    fi
  done < <(
    find "${GUI_DIR}" -maxdepth 1 -type f -print0
    find "${GUI_DIR}/lib" -type f -print0
  )
fi
