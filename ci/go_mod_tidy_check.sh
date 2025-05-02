#!/bin/sh
#
# This script checks whether go.mod and go.sum files are tidy and exits with exit code 1 if it is
# not.
# Note: go mod tidy does not have a dry run. Therefore, this script actually modifies the go.mod and
# go.sum files if needed by actually executing `go mod tidy`.
#
go mod tidy

if ! git diff --exit-code go.mod go.sum; then
	echo "::error::go.mod or go.sum is not tidy. Please run 'go mod tidy' and commit the changes."
	exit 1
fi
