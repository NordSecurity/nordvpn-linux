#!/bin/bash
set -eu

usage() {
    echo "Usage:"
    echo -e "\tnexus_get.sh <arguments> <filepath>"
    echo "Args:"
    echo -e "\t-o output filename"
    echo -e "\t-r <qa/release> repository, defaults to release."
    echo -e "\t-v verbose"
    exit 1
}

download_file() {
    local filepath=$1
    filename="${OUTPUT_FILE:-$(basename "${filepath}")}"
    NEXUS_REPOSITORY="ll-gitlab-${REPOSITORY_TYPE}"
    curl "${VERBOSITY}" -f -u "${NEXUS_CREDENTIALS}" "${NEXUS_URL}/repository/${NEXUS_REPOSITORY}/${filepath}" -o "${filename}" || { echo "Failed to download file!" ; exit 1 ; }
}

[[ "$#" == 0 ]] && usage

VERBOSITY='-s'
REPOSITORY_TYPE='release'

while [[ $# -gt 0 ]] ; do
    cmd=${1,,}
    case "$cmd" in
            -o)
                shift
                [[ -z "$1" ]] && { echo "No output is file provided!" ; exit 1 ; }
                OUTPUT_FILE="$1"
                shift
                ;;
            -r)
                shift
                REPOSITORY_TYPE="$1"
                [[ -z "$1" ]] && { echo "No repository type is provided!" ; exit 1 ; }
                shift
                ;;
            -v)
                VERBOSITY='-v'
                shift
                ;;
            -h | --help)
                usage
                ;;
            *)
                FILE_TO_DOWNLOAD="$1"
                break
                ;;
    esac
done

download_file "${FILE_TO_DOWNLOAD}"
