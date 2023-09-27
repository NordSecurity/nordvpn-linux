#!/usr/bin/env bash
set -euxo

PACKAGE="${1}"
FILES=$(find "${WORKDIR}"/dist/app/"${PACKAGE}" -type f -name "*.${PACKAGE}")

source "${WORKDIR}"/ci/repository_name.sh "${PACKAGE}"

case "${PACKAGE}" in
    "deb")
        for FILE in $FILES; do
            echo "Uploading ${FILE}"
            pulp-admin deb repo uploads deb --repo-id="${REPOSITORY}" --file="${FILE}"
            echo "Uploaded ${FILE}"
        done
        ;;
    "rpm")
        for FILE in $FILES; do
            echo "Uploading ${FILE}"
            ARCH=$(echo "${FILE}" | awk -F'[..]' '{print $(NF-1)}') # cut arch part
            pulp-admin rpm repo uploads rpm --repo-id="${REPOSITORY}"-"${ARCH}" --file="${FILE}"
            echo "Uploaded ${FILE}"
        done
        ;;
esac
