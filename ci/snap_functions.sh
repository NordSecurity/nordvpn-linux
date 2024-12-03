#!/bin/bash

snap_connect_interfaces() {
    local SNAP_NAME=nordvpn

    # List all connections for the Snap
    echo "Checking connections for Snap package: ${SNAP_NAME}"
    connections=$(snap connections "${SNAP_NAME}")

    # Display and process unconnected interfaces
    echo
    echo "Unconnected connections for ${SNAP_NAME}:"
    unconnected=$(echo "${connections}" | awk '
    NR > 1 && $3 == "-" {
        print $1
    }')

    if [ -z "${unconnected}" ]; then
        echo "All connections are already connected."
        return
    fi

    echo "${unconnected}"

    # Attempt to connect each unconnected interface
    echo
    for interface in ${unconnected}; do
        echo "Connecting interface: ${interface}"
        if sudo snap connect "${SNAP_NAME}:${interface}"; then
            echo "Successfully connected: ${interface}"
        else
            echo "Failed to connect: ${interface}"
        fi
    done

    echo
    echo "Connection process completed for ${SNAP_NAME}."
    # Show current connections state
    snap connections "${SNAP_NAME}"
}
