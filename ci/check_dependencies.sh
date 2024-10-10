#!/bin/bash
set -euxo pipefail

source "${WORKDIR}/ci/archs.sh"
source "${WORKDIR}/ci/export_lib_versions.sh"

temp_dir="${WORKDIR}/bin/deps/artifacts"
lib_root="${WORKDIR}/bin/deps/lib"

libtelio_artifact_url="${LIBTELIO_ARTIFACTS_URL}/${LIBTELIO_VERSION}/linux.zip"
libtelio_zipfile="${temp_dir}/libtelio-${LIBTELIO_VERSION}.zip"
libtelio_dst="${temp_dir}/libtelio-${LIBTELIO_VERSION}"

libdrop_artifact_url="${LIBDROP_ARTIFACTS_URL}/${LIBDROP_VERSION}/linux.zip"
libdrop_zipfile="${temp_dir}/libdrop-${LIBDROP_VERSION}.zip"
libdrop_dst="${temp_dir}/libdrop-${LIBDROP_VERSION}"

libmoose_nordvpnapp_artifact_url="${LIBMOOSE_NORDVPNAPP_ARTIFACTS_URL}/${LIBMOOSE_NORDVPNAPP_VERSION}/linux.zip"
libmoose_nordvpnapp_zipfile="${temp_dir}/libmoose-nordvpnapp-${LIBMOOSE_NORDVPNAPP_VERSION}.zip"
libmoose_nordvpnapp_dst="${temp_dir}/libmoose-nordvpnapp-${LIBMOOSE_NORDVPNAPP_VERSION}"

libmoose_worker_artifact_url="${LIBMOOSE_WORKER_ARTIFACTS_URL}/${LIBMOOSE_WORKER_VERSION}/linux.zip"
libmoose_worker_zipfile="${temp_dir}/libmoose-worker-${LIBMOOSE_WORKER_VERSION}.zip"
libmoose_worker_dst="${temp_dir}/libmoose-worker-${LIBMOOSE_WORKER_VERSION}"

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
  curl \
    --retry 3 \
    --retry-delay 2 \
    --header "PRIVATE-TOKEN: ${GL_ACCESS_TOKEN}" \
    -o "${out_file}" \
    -L "${artifact_url}"
  # re-enable tracing
  set -x
}

# This function makes a copy of .so file. It requires two parameters:
# - so_prefix - path to the .so file in unzipped artifact directory up to architecture name
# - so_file - name of the .so file to copy
#
# Those two parameters are joined with architecture and all together create
# full path. Example:
#
# `copy_to_libs "${libtelio_dst}/linux/release" "libtelio.so"`
#
# makes a copy of ${libtelio_dst}/linux/release/${architecture}/libtelio.so to the
# `libroot/${architecture}/latest directory`.
function copy_to_libs() {
  if [[ $# -ne 2 ]]; then
    echo "Two parameters are required for copy_to_libs function:"
    echo "copy_to_libs <so_prefix> <so_file>"
    echo -e "\t so_prefix - prefix of the path to the so file _up to the architecture name_"
    echo -e "\t so_file - .so file name (including .so extension)"
    echo "You passed ${#} arg(s): ${*}"
    exit 1
  fi
  so_prefix="${1}"
  so_file="${2}"
  for arch in "${!ARCHS_SO_REVERSE[@]}"; do
    so_path="${so_prefix}/${arch}/${so_file}"
    # libraries have different names for the same architectures and
    # ARCHS_SO_REVERSE contains architecture names used in all three libraries,
    # so I check if the .so file exists.
    if [[ -e ${so_path} ]]; then
      cp "${so_path}" "${lib_root}/${ARCHS_SO_REVERSE[${arch}]}/latest"
    fi
  done
}

# Artifacts zips are pretty big. Skip downloading if it was already done.
if [[ -e "${lib_root}/checkout-completed-flag" ]]; then
  echo "Dependencies already downloaded. Skipping download step."
  exit 0
fi

# ====================[  Download artifacts ]=========================

fetch_gitlab_artifact "${libtelio_artifact_url}" "${libtelio_zipfile}"
fetch_gitlab_artifact "${libdrop_artifact_url}" "${libdrop_zipfile}"

if [[ "${FEATURES:-""}" == *internal* ]]; then
  fetch_gitlab_artifact "${libmoose_nordvpnapp_artifact_url}" "${libmoose_nordvpnapp_zipfile}"
  fetch_gitlab_artifact "${libmoose_worker_artifact_url}" "${libmoose_worker_zipfile}"
fi

# ====================[  Unzip files ]=========================

unzip -o "${libtelio_zipfile}" -d "${temp_dir}" && mv "${temp_dir}/dist" "${libtelio_dst}"
unzip -o "${libdrop_zipfile}" -d "${temp_dir}" && mv "${temp_dir}/libdrop" "${libdrop_dst}"

if [[ "${FEATURES:-""}" == *internal* ]]; then
  unzip -o "${libmoose_nordvpnapp_zipfile}" -d "${temp_dir}" && mv "${temp_dir}/out" "${libmoose_nordvpnapp_dst}"
  unzip -o "${libmoose_worker_zipfile}" -d "${temp_dir}" && mv "${temp_dir}/out" "${libmoose_worker_dst}"
fi

# ====================[  Copy to bin/deps/libs ]=========================

mkdir -p "${lib_root}/"{amd64,aarch64,armel,armhf,i386}/latest/

# libtelio
copy_to_libs "${libtelio_dst}/linux/release" "libtelio.so"

# sqlite3
copy_to_libs "${libtelio_dst}/linux/release" "libsqlite3.so"

# libdrop
copy_to_libs "${libdrop_dst}/dist/linux/release" "libnorddrop.so"

if [[ "${FEATURES:-""}" == *internal* ]]; then
  # moose nordvpnapp
  copy_to_libs "${libmoose_nordvpnapp_dst}/nordvpnapp/bin/nordvpnapp/linux" "libmoosenordvpnapp.so"

  # moose worker
  copy_to_libs "${libmoose_worker_dst}/worker/bin/worker/linux" "libmooseworker.so"
fi

# remove leftovers
rm -rf "${temp_dir}"

touch "${lib_root}/checkout-completed-flag"
