#!/usr/bin/env bash
set -euxo pipefail

source "${WORKDIR}/ci/qa_tests_env.sh"

if ! "${WORKDIR}"/ci/install_snap.sh; then
    echo "failed to install snap"
    exit 1
fi

# cover dir is not configured for now because of snap confinement, it is a bit more complicated,
# and it cannot be set anywhere
rm -fr "${GOCOVERDIR}"

"${WORKDIR}/ci/qa_run_tests.sh" "$@"
