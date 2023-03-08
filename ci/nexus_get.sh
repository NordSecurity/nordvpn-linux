#!/bin/bash
set -eu

usage() {
    echo "Usage:"
    echo -e "\tnexus_get.sh <arguments> <filepath>"
    echo "Args:"
    echo -e "\t-o output filename"
    echo -e "\t-r <qa/releases> repository, default release."
    echo -e "\t-v verbose"
    exit 1
}

download_file() {
    local filepath=$1
    filename="${OUTPUT_FILE:-$(basename "${filepath}")}"
    curl "${VERBOSITY}" -f -u "${NEXUS_QA_READ_USER}:${NEXUS_QA_READ_CRED}" "https://${NEXUS_SERVER}/repository/gitlab-${REPOSITORY_TYPE}/${filepath}" -o "${filename}" || { echo "Failed to download file!" ; exit 1 ; }
}

[[ "$#" == 0 ]] && usage

VERBOSITY='-s'
REPOSITORY_TYPE='releases'

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
