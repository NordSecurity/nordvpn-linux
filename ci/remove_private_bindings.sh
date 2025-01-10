#!/bin/bash
# This script removes private module dependencies from go.mod by dropping the 'require' and 'replace' directives for a given module name.
# Usage: ./ci/remove_private_bindings.sh <MODULE>

bindings_name=$1

if [ -z "$bindings_name" ]; then
  echo "error no bindings name provided"
  exit 1
fi

go mod edit -droprequire="$bindings_name" || { echo "failed to drop require for $bindings_name"; exit 1; }
go mod edit -dropreplace="$bindings_name" || { echo "failed to drop replace for $bindings_name"; exit 1; }