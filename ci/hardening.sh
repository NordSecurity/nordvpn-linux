#!/usr/bin/env bash
set -euxo pipefail

check_hardening() {
  local binary="$1"
  shift
  hardening-check -q --nocfprotection "$@" "${binary}"
}

check_hardening_if_exists() {
  if [[ -f "$1" ]]; then
    check_hardening "$@"
  fi
}

check_runpath() {
  local binary="$1" runpath entry entries
  runpath=$(readelf -dW "${binary}" | sed -nE 's/.*\(R(UN)?PATH\).*\[(.*)\]/\2/p')
  IFS=':' read -ra entries <<<"${runpath}"
  for entry in "${entries[@]}"; do
    # shellcheck disable=SC2016
    if [[ "${entry}" != '$ORIGIN' && "${entry}" != '$ORIGIN/'* ]]; then
      echo "${binary}: contains run-time search path '${entry}'" >&2
      exit 1
    fi
  done
}

check_stripped() {
  local binary="$1"
  if [[ "$(file -b "${binary}")" == *"not stripped"* ]]; then
    echo "${binary}: is not stripped" >&2
    exit 1
  fi
}

check_gui_binary() {
  local binary="$1" flags=()
  case "${binary##*/}" in
  # prebuilt Flutter engine, uses libc functions that cannot be fortified
  libflutter_linux_gtk.so) flags=(--nofortify) ;;
  # contains only Dart compiled code and data
  libapp.so) flags=(--nostackprotector --norelro --nobindnow) ;;
  esac
  check_hardening "${binary}" "${flags[@]}"
  check_runpath "${binary}"
  check_stripped "${binary}"
}

check_gui_bundle() {
  local bundle="$1" binary
  while IFS= read -r -d '' binary; do
    if [[ "$(file -b "${binary}")" == ELF* ]]; then
      check_gui_binary "${binary}"
    fi
  done < <(
    find "${bundle}" -maxdepth 1 -type f -print0
    find "${bundle}/lib" -type f -print0
  )
}

check_hardening_if_exists "${WORKDIR}/bin/deps/openvpn/${ARCH}/latest/openvpn"

# Stack protection is forcibly disabled when building Go apps
# https://github.com/golang/go/blob/202a1a57064127c3f19d96df57b9f9586145e21c/src/runtime/cgo/cgo.go#L28
for binary in nordvpnd nordfileshare nordvpn norduserd; do
  check_hardening_if_exists "${WORKDIR}/bin/${ARCH}/${binary}" --nostackprotector
done

GUI_DIR="${WORKDIR}/bin/${ARCH}/gui"
if [[ -d "${GUI_DIR}" ]]; then
  check_gui_bundle "${GUI_DIR}"
fi
