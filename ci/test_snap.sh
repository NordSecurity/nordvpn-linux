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

add_daemon_logs() {
    # append the snap daemon logs into the daemon.log file
    echo "----------------------------------------- " >> "${LOGS_FOLDER}/daemon.log"
    echo "----------- start daemon log ------------ " >> "${LOGS_FOLDER}/daemon.log"
    echo "----------------------------------------- " >> "${LOGS_FOLDER}/daemon.log"
    sudo journalctl -b -u snap.nordvpn.nordvpnd.service >> "${LOGS_FOLDER}/daemon.log"
}

trap add_daemon_logs EXIT INT TERM

"${WORKDIR}/ci/qa_run_tests.sh" "$@"

# restore to original value
GOCOVERDIR="$ORIG_COVERDIR"
