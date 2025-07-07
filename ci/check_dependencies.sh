#!/bin/bash
set -euxo pipefail

source "${WORKDIR}/ci/archs.sh"
source "${WORKDIR}/ci/export_lib_versions.sh"
source "${WORKDIR}/ci/populate_current_lib_ver.sh"

temp_dir="${WORKDIR}/bin/deps/artifacts"
lib_root="${WORKDIR}/bin/deps/lib"

libtelio_artifact_url="${LIBTELIO_ARTIFACTS_URL}/${LIBTELIO_VERSION}/linux.zip"
libdrop_artifact_url="${LIBDROP_ARTIFACTS_URL}/${LIBDROP_VERSION}/linux.zip"
libmoose_nordvpnapp_artifact_url="${LIBMOOSE_NORDVPNAPP_ARTIFACTS_URL}/${LIBMOOSE_NORDVPNAPP_VERSION}/linux.zip"
libmoose_worker_artifact_url="${LIBMOOSE_WORKER_ARTIFACTS_URL}/${LIBMOOSE_WORKER_VERSION}/linux.zip"
libquench_artifact_url="${LIBQUENCH_ARTIFACTS_URL}/${LIBQUENCH_VERSION}/linux.zip"

# Uncomment this when all libraries migrate artifacts
# if [[ ${CI+x} ]]; then
#   header="JOB-TOKEN:${CI_JOB_TOKEN}"
# else
#   header="PRIVATE-TOKEN:${GL_ACCESS_TOKEN}"
# fi

mkdir -p "${temp_dir}"

function fetch_gitlab_artifact() {
  if [[ $# -ne 2 ]]; then
    echo "Two parameters are required for fetch_gitlab_artifact function:"
    echo "fetch_gitlab_artifact <artifact_url> <out_file>"
    echo -e "\t artifact_url - URL to the artifact file on the GitLab"
    echo -e "\t out_file - destination path of the artifact"
    echo "You passed ${#} arg(s): ${*}"
    exit 1
  fi
  artifact_url="${1}"
  out_file="${2}"

  echo "Downloading artifact from ${artifact_url} to ${out_file}"
  # disable tracing - don't show the token
  set +x
  header="JOB-TOKEN:${CI_JOB_TOKEN}"
  for _ in $(seq 1 2)
  do
    curl \
      --retry 3 \
      --retry-delay 2 \
      --fail \
      --header "$header" \
      -o "${out_file}" \
      -L "${artifact_url}" && break || echo "Token failed"
    header="PRIVATE-TOKEN:${GL_ACCESS_TOKEN}"
  done
  # re-enable tracing
  set -x
}

function copy_to_libs() {
  if [[ $# -ne 4 ]]; then
    echo "copy_to_libs <zipfile> <so_prefix> <so_file> <out_dir>"
    echo -e "\t zipfile - path to a ZIP archive containing the artifacts"
    echo -e "\t so_prefix - prefix of the path to the so file _up to the architecture name_"
    echo -e "\t so_file - .so file name (including .so extension)"
    echo -e "\t out_dir - directory where to put the .so file"
    echo "You passed ${#} arg(s): ${*}"
    exit 1
  fi
  zipfile="${1}"
  so_file="${2}"
  so_prefix="${3}"
  out_dir="${4}"

  files_in_zip="$(unzip -l "${zipfile}")"
  for arch in "${!ARCHS_SO_REVERSE[@]}"; do
    # libraries have different names for the same architectures and
    # ARCHS_SO_REVERSE contains architecture names used in all three libraries,
    # so I check if the .so file exists.
    dir_in_zip="${so_prefix}/${arch}/${so_file}"
    if [[ "${files_in_zip}" != *"${dir_in_zip}"* ]]; then
	    continue
    fi
    so_out_dir="${out_dir}/${ARCHS_SO_REVERSE[${arch}]}"
    mkdir -p "${so_out_dir}"
    unzip -jo "${zipfile}" -d "${so_out_dir}" "${dir_in_zip}"
  done
}

function fetch_dependency() {
  if [[ $# -ne 5 ]]; then
    echo "fetch_dependency <name> <artifact_url> <so_file> <so_prefix> <version> [no_download]"
    echo -e "\t name - name of the dependency that is used to determine binary file paths"
    echo -e "\t artifact_url - URL to the artifact file on the GitLab"
    echo -e "\t so_prefix - prefix of the path to the so file _up to the architecture name_"
    echo -e "\t so_file - .so file name (including .so extension)"
    echo -e "\t version - version of the deendency"
    echo "You passed ${#} arg(s): ${*}"
    exit 1
  fi
  name=${1}
  artifact_url="${2}"
  so_file="${3}"
  so_prefix="${4}"
  version="${5}"

  out_dir="${lib_root}/${name}"
  version_out_dir="${out_dir}/${version}"
  checkout_completed_flag_file="${version_out_dir}/checkout-completed-flag"
  current_version_dir="${out_dir}/current"

  if [[ -e ${checkout_completed_flag_file} ]]; then
    echo "${name} ${version} already downloaded. Skipping download step."
  else
    zipfile="${temp_dir}/${name}-${version}.zip"
    fetch_gitlab_artifact "${artifact_url}" "${zipfile}"
    copy_to_libs "${zipfile}" "${so_file}" "${so_prefix}" "${version_out_dir}"
    touch "${checkout_completed_flag_file}"
  fi
  ln -sfn "${version}" "${current_version_dir}"
  populate_current_ver "${lib_root}/current" "${current_version_dir}" "${so_file}"
}

# Remove all the synlinks to libraries to avoid junk in the packages
rm -rf "${lib_root}/current"

if [[ "${FEATURES:-""}" == *telio* ]]; then
  fetch_dependency "libtelio" "${libtelio_artifact_url}" \
    "libtelio.so" "dist/linux/release" "${LIBTELIO_VERSION}"
fi

if [[ "${FEATURES:-""}" == *drop* ]]; then
  fetch_dependency "libdrop" "${libdrop_artifact_url}" \
    "libnorddrop.so" "libdrop/dist/linux/release" "${LIBDROP_VERSION}"
fi

if [[ "${FEATURES:-""}" == *quench* ]]; then
  fetch_dependency "libquench" "${libquench_artifact_url}" \
    "libquench.so" "dist/linux/release" "${LIBQUENCH_VERSION}"
fi

if [[ "${FEATURES:-""}" == *moose* ]]; then
  fetch_dependency "libmoose-worker" "${libmoose_worker_artifact_url}" \
    "libmooseworker.so" "out/worker/bin/worker/linux" "${LIBMOOSE_WORKER_VERSION}"

  fetch_dependency "libmoose-nordvpnapp" "${libmoose_nordvpnapp_artifact_url}" \
  "libmoosenordvpnapp.so" "out/nordvpnapp/bin/nordvpnapp/linux" "${LIBMOOSE_NORDVPNAPP_VERSION}"

  zipfile="${temp_dir}/libmoose-nordvpnapp-${LIBMOOSE_NORDVPNAPP_VERSION}.zip"
  if [[ -e "${zipfile}" ]]; then
    copy_to_libs "${zipfile}" "libsqlite3.so" "out/nordvpnapp/bin/common/linux" "${lib_root}/${name}/${LIBMOOSE_NORDVPNAPP_VERSION}"
  fi
  populate_current_ver "${lib_root}/current" "${lib_root}/libmoose-nordvpnapp/current" "libsqlite3.so"
fi

# remove leftovers
rm -rf "${temp_dir}"
