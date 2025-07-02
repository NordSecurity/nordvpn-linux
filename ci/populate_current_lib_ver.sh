#!/bin/bash
set -euxo pipefail

function populate_current_ver() {
  current_ver_dir="${1}"
  lib_dir="${2}"
  so_file="${3}"

  for arch in "${ARCHS[@]}"; do
    mkdir -p "${current_ver_dir}/${arch}"
    ln -snfr "${lib_dir}/${arch}/${so_file}" "${current_ver_dir}/${arch}/${so_file}"
  done
}
