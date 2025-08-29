#!/usr/bin/env bash
# Set the needed env variables to be able to build packages for distribution
set -euxo pipefail

NAME=nordvpn-gui
export NAME

DISPLAY_NAME="NordVPN GUI"
export DISPLAY_NAME

# Extracts the app version from pubspecs.yaml
VERSION=$(grep 'version:' pubspec.yaml | cut -d ' ' -f 2)
echo "Version: $VERSION"

export VERSION
