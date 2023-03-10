#!/usr/bin/env bash
set -euxo

PACKAGE="${1}"

source "${CI_PROJECT_DIR}"/ci/archs.sh
source "${CI_PROJECT_DIR}"/ci/repository_name.sh "${PACKAGE}"
"${CI_PROJECT_DIR}"/ci/pulp_ca_certificate.sh

case "${PACKAGE}" in
    "deb")
        go run "${CI_PROJECT_DIR}"/cmd/pulp/main.go \
            --hostname "https://${PULP_HOST}" \
            --username "${PULP_USER}" \
            --password "${PULP_PASS}" \
            --certificate /tmp/pulp.crt \
            --keep 3 \
            --package "${PACKAGE}" \
            --repository "${REPOSITORY}"
        ;;
    "rpm")
        for arch in "${ARCHS[@]}"; do
            go run "${CI_PROJECT_DIR}"/cmd/pulp/main.go \
                --hostname "https://${PULP_HOST}" \
                --username "${PULP_USER}" \
                --password "${PULP_PASS}" \
                --certificate /tmp/pulp.crt \
                --keep 3 \
                --package "${PACKAGE}" \
                --repository "${REPOSITORY}-${ARCHS_RPM[$arch]}"
        done
        ;;
esac
