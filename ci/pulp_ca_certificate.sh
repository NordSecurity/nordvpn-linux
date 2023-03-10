#!/usr/bin/env bash
set -euxo

# Pulp uses self-signed certificates issued by Alpha SSL, which are not
# in OS' certificate chain, so we need to download their CA cert for later use.
# Also, CA certificate expires eventually, so just update the environment
# variable in the CI when it happens(next expiration is February 24th in 2024).
curl -o /tmp/pulp.crt "${PULP_CA_URL}"
