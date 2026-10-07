#!/usr/bin/env bash
set -euxo pipefail

source "${WORKDIR}/ci/qa_tests_env.sh"

if ! "${WORKDIR}"/ci/install_snap.sh; then
    echo "failed to install snap"
    exit 1
fi

# because of snap confinement is not possible to save anywhere the cover files
# for now set the cover dir to /tmp to prevent CLI errors.
# If later this will be needed also the snap nordvpn service file needs to be changed.
rm -fr "${GOCOVERDIR}"
ORIG_COVERDIR="$GOCOVERDIR"
GOCOVERDIR="/tmp/"

"${WORKDIR}/ci/qa_run_tests.sh" "$@"

# restore to original value
GOCOVERDIR="$ORIG_COVERDIR"
