#!/bin/bash
# Download is idempotent and will not redownload if the file already exists
set -eo

source "${CI_PROJECT_DIR}"/ci/archs.sh

usage() {
    echo "Usage:"
    echo -e "\ndownload_from_remote.sh <download_repository_id> <version> <os> <arch> <extension>"
    echo "Args:"
    echo -e "\t-a binary architecture"
    echo -e "\t-i repository ID to download"
    echo -e "\t-d download directory name"
    echo -e "\t-o operating system"
    echo -e "\t-r <qa/releases> repository, default release."
    echo -e "\t-v binary version to download"
    echo -e "\t-x binary extension"
    exit 1
}

REPOSITORY_TYPE='releases'
REPOSITORY_NAME=''
FILE_EXTENSION=''
DIR_NAME=''
ARCHS=''

while [[ $# -gt 0 ]] ; do
    cmd=${1,,}
    case "$cmd" in
        -r)
            shift
            REPOSITORY_TYPE="$1"
            [[ -z $1 ]] && { echo "No repository type is provided!" ; exit 1 ; }
            shift
            ;;
        -n)
            shift
            REPOSITORY_NAME=${1,,}
            [[ -z $1 ]] && { echo "No repository name is provided!" ; exit 1 ; }
            shift
            ;;
        -i)
            shift
            REPOSITORY_ID="$1"
            [[ -z $1 ]] && { echo "No repository to download is provided!" ; exit 1 ; }
            shift
            ;;
        -d)
            shift
            DIR_NAME="$1"
            shift
            ;;
        -v)
            shift
            BINARY_VERSION="$1"
            [[ -z $1 ]] && { echo "No binary version to download is provided!" ; exit 1 ; }
            shift
            ;;
        -a)
            shift
            ARCHS=${1,,}
            [[ -z $1 ]] && { echo "No binary architecture is provided!" ; exit 1 ; }
            shift
            ;;
        -o)
            shift
            OS=${1,,}
            [[ -z $1 ]] && { echo "No operating system is provided!" ; exit 1 ; }
            shift
            ;;
        -x)
            shift
            FILE_EXTENSION=${1,,}
            [[ -z $1 ]] && { echo "No file extension is provided!" ; exit 1 ; }
            shift
            ;;
        -h | --help)
            usage
            ;;
        *)
            echo "No repository to download is provided!" ; exit 1
            ;;
    esac
done

if [[ -n "${DIR_NAME}" ]]; then
    DOWNLOAD_DIR="${CI_PROJECT_DIR}/bin/deps/${DIR_NAME}"
else
    DOWNLOAD_DIR="${CI_PROJECT_DIR}/bin/deps/${REPOSITORY_NAME}"
fi

mkdir -p "${DOWNLOAD_DIR}"

for arch in ${ARCHS} ; do
    output_arch=${ARCHS_REVERSE[$arch]}
    arch_dir="${DOWNLOAD_DIR}/${output_arch}"
    output_dir="${arch_dir}/${BINARY_VERSION}"
    latest_dir="${arch_dir}/latest"
    output_file="${output_dir}/${REPOSITORY_NAME}${FILE_EXTENSION}"
    mkdir -p "${output_dir}"
    # Create a symlink so that path to the newest binary could be used statically (e.g. in IDEs)
    # Symlink is relative so it would work in Docker containers as well
    ln -fsnr "${output_dir}" "${latest_dir}" 
    [[ -e "${output_file}" ]] && continue
    echo "Downloading ${REPOSITORY_NAME}-${arch} ${BINARY_VERSION}..."
    "${CI_PROJECT_DIR}"/ci/nexus_get.sh -r "${REPOSITORY_TYPE}" -o "${output_file}" \
        "${REPOSITORY_ID}/${BINARY_VERSION}/${OS}/${arch}/${REPOSITORY_NAME}${FILE_EXTENSION}"

done

echo 'Done!'
