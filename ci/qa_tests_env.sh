#!/usr/bin/env bash

set -euxo pipefail

# Defines env variables common when running pytests

# Disable the TUI loader indicator to prevent interference during automated tests
export DISABLE_TUI_LOADER=1

# go code cover dir
export COVERDIR="covdatafiles"
export GOCOVERDIR="${WORKDIR}/$COVERDIR"
