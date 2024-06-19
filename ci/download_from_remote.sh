#!/bin/bash
# Download is idempotent and will not redownload if the file already exists
set -exo pipefail

source "${WORKDIR}"/ci/archs.sh

usage() {
    echo "Usage:"
    echo -e "\ndownload_from_remote.sh -a <arch> -O <output_dir> -p <project_id> -v <package_version> <file_name>"
    echo "Args:"
    echo -e "\t-a binary architecture"
    echo -e "\t-O output directory name"
    echo -e "\t-p project ID"
    echo -e "\t-v package version"
    exit 1
}

FILE=''
DIR_NAME=''
ARCHS=''

while [[ $# -gt 0 ]] ; do
    cmd=$1
    case "$cmd" in
        -p)
            shift
            PROJECT_ID="$1"
            [[ -z $1 ]] && { echo "No project ID provided!" ; exit 1 ; }
            shift
            ;;
        -O)
            shift
            DIR_NAME="$1"
            [[ -z $1 ]] && { echo "No directory provided!" ; exit 1 ; }
            shift
            ;;
        -v)
            shift
            PACKAGE_VERSION="$1"
            [[ -z $1 ]] && { echo "No binary version to download is provided!" ; exit 1 ; }
            shift
            ;;
        -a)
            shift
            ARCHS=${1,,}
            [[ -z $1 ]] && { echo "No binary architecture is provided!" ; exit 1 ; }
            shift
            ;;
        -h | --help)
            usage
            ;;
        *)
            FILE="$1"
            [[ -z $1 ]] && { echo "No file name provided!" ; exit 1 ; }
            shift
            ;;
    esac
done

if [[ -n "${DIR_NAME}" ]]; then
    DOWNLOAD_DIR="${WORKDIR}/bin/deps/${DIR_NAME}"
else
    DOWNLOAD_DIR="${WORKDIR}/bin/deps"
fi

mkdir -p "${DOWNLOAD_DIR}"

for arch in ${ARCHS} ; do
    output_arch=${ARCHS_REVERSE[$arch]}
    arch_dir="${DOWNLOAD_DIR}/${output_arch}"
    output_dir="${arch_dir}/${PACKAGE_VERSION}"
    latest_dir="${arch_dir}/latest"
    output_file="${output_dir}/${FILE}"
    mkdir -p "${output_dir}"
    # Create a symlink so that path to the newest binary could be used statically (e.g. in IDEs)
    # Symlink is relative so it would work in Docker containers as well
    ln -fsnr "${output_dir}" "${latest_dir}" 
    [[ -e "${output_file}" ]] && continue
    echo "Downloading ${arch}/${PACKAGE_VERSION}/${FILE}..."
    curl \
        -fL \
        -o "$output_file" \
        -u "${NVPN_LINUX_GL_DEPS_CREDS:-}" \
        "${CI_API_V4_URL}/projects/${PROJECT_ID}/packages/generic/${arch}/${PACKAGE_VERSION}/${FILE}"

done

echo 'Done!'
