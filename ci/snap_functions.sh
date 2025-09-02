#!/usr/bin/env bash
set -euxo pipefail

# using nordvpn status connect the needed snap interfaces
snap_connect_interfaces() {
    # List all connections for the Snap
    echo "Make the snap connections"

    wait_for_daemon
    output=$(nordvpn status || true)

    echo "$output" | grep -o '^sudo snap connect .*' | while read -r cmd; do
        echo "Executing: $cmd"
        # Run the command
        $cmd
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
        output=$(nordvpn status || true)
        # cannot use return code 0 because snap interfaces are not connected
        # instead check that the output have "nordvpnd.sock not found"
        if [[ "$output" != *"nordvpnd.sock"* ]]; then
            return
        fi
        echo "Attempt $i failed: $output"
        sleep 1
    done
}
