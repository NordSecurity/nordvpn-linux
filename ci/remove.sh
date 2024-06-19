#!/usr/bin/env bash
set -euxo pipefail

PACKAGE="${1}"

source "${WORKDIR}"/ci/archs.sh
source "${WORKDIR}"/ci/repository_name.sh "${PACKAGE}"
"${WORKDIR}"/ci/pulp_ca_certificate.sh

case "${PACKAGE}" in
    "deb")
        go run "${WORKDIR}"/cmd/pulp/main.go \
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
            go run "${WORKDIR}"/cmd/pulp/main.go \
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
