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

redirect_logs() {
    local log_path="${LOGS_FOLDER}"/daemon.log

    mkdir -p "$(dirname "$log_path")"
    touch "$log_path"
    chmod 777 "$log_path"

    sudo mkdir -p /etc/systemd/system/snap.nordvpn.nordvpnd.service.d

    cat <<EOF | sudo tee /etc/systemd/system/snap.nordvpn.nordvpnd.service.d/override.conf > /dev/null
[Service]
StandardOutput=append:$log_path
StandardError=append:$log_path
EOF

    sudo systemctl daemon-reload
    sudo systemctl restart snap.nordvpn.nordvpnd.service
}
