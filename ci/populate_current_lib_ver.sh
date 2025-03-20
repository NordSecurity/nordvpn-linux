#!/bin/bash
set -euxo pipefail

function populate_current_ver() {
  current_ver_dir="${1}"
  lib_dir="${2}"
  so_file="${3}"

  for path in "${lib_dir}"/*; do
    arch="${path##*/}"
    if [[ "${arch}" = "checkout-completed-flag" ]]; then
	    continue
    fi
    mkdir -p "${current_ver_dir}/${arch}"
    ln -sfnr "${path}/${so_file}" "${current_ver_dir}/${arch}/${so_file}"
  done
}
