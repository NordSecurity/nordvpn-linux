#!/usr/bin/env bash
set -euxo pipefail

# using nordvpn status connect the needed snap interfaces
snap_connect_interfaces() {
    # List all connections for the Snap
    echo "Make the snap connections"

    wait_for_daemon
    SNAP_CONNECT_CMDS=$(nordvpn status | grep -o '^sudo snap connect .*' || true)

    echo "${SNAP_CONNECT_CMDS}" | while read -r CMD; do
        echo "Executing: ${CMD}"
        # Run the command
        ${CMD}
    done
    echo "All permissions granted successfully."

    wait_for_daemon

    # recheck that the nordvpn status is successful
    nordvpn status

    echo "Snap connection process completed."
}

# wait until daemon is running
wait_for_daemon() {
    for i in {1..10}; do
        STATUS=$(nordvpn status || true)
        # cannot use return code 0 because snap interfaces are not connected
        # instead check that the output have "nordvpnd.sock not found"
        if [[ "${STATUS}" != *"nordvpnd.sock"* ]]; then
            return
        fi
        echo "Attempt ${i} failed: ${STATUS}"
        sleep 1
    done
}
