#!/usr/bin/env bash
set -euxo pipefail

declare -A LIB_NAME_TO_PACKAGE=(
  [libtelio]=github.com/NordSecurity/libtelio-go
  [libdrop]=github.com/NordSecurity/libdrop-go
)

declare -A LIB_NAME_TO_VERSION
[[ -n "${LIBTELIO_VERSION:-}" ]] && LIB_NAME_TO_VERSION[libtelio]=${LIBTELIO_VERSION}
[[ -n "${LIBDROP_VERSION:-}" ]] && LIB_NAME_TO_VERSION[libdrop]=${LIBDROP_VERSION}

declare -A LIB_NAME_TO_BINDINGS_VERSION
[[ -n "${LIBTELIO_BINDINGS_VERSION:-}" ]] && LIB_NAME_TO_BINDINGS_VERSION[libtelio]=${LIBTELIO_BINDINGS_VERSION}
[[ -n "${LIBDROP_BINDINGS_VERSION:-}" ]] && LIB_NAME_TO_BINDINGS_VERSION[libdrop]=${LIBDROP_BINDINGS_VERSION}

lib_name=$1
repo_path="${LIB_NAME_TO_PACKAGE[${lib_name}]}"
lib_version="${LIB_NAME_TO_VERSION[${lib_name}]:-}"

bindings_version="${LIB_NAME_TO_BINDINGS_VERSION[${lib_name}]:-}"

new_version="${bindings_version:-${lib_version}}"
major_version=$(echo "${new_version}" | cut -d'.' -f1)

current_version=$(grep "^\s*${repo_path}/" go.mod | awk '{print $2}' | head -1) || true
current_major=$(echo "${current_version}" | cut -d'.' -f1)
new_module="${repo_path}/${major_version}"

if [[ -n "${current_major}" && -n "${major_version}" && "${current_major}" != "${major_version}" ]]; then
  # NOTE: GONOSUMDB allows ignoring module checksum check made by go. This is
  # not needed for regular builds, because we commit changed go.sum which is
  # trusted by go tooling, but when we are changing module on the fly, go
  # needs to recompute checksum database and we don't allow to do this in
  # docker, so this temporary override here allows module change on the fly
  # without compilation errors.
  GONOSUMDB="${repo_path}/*" go get "${new_module}@${new_version}"
  find . -name "*.go" -exec sed -i "s|\"${repo_path}/${current_major}|\"${repo_path}/${major_version}|g" {} +
fi
