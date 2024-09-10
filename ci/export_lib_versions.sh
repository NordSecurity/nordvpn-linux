#!/bin/bash
set -euxo pipefail

# Read `lib-versions.env` file line by line and export every env variable
# defined in that file only if the env variable with such name does not exist
# yet. This allows overriding the variables by the pipeline, but there is always
# default defined in one place - `lib-versions.env`.
while read -r line || [[ -n "$line" ]]; do
  # if the line is not a comment and is not empty; then
  if ! [[ "$line" =~ ^#.*$ ]] && [[ -n "$line" ]]; then
    key=$(echo "${line}" | cut -d'=' -f1)
    echo "key = ${key}"
    # only export the variable if it is not already set
    if [[ -z "${!key:-}" ]]; then
      # shellcheck disable=SC2163
      export "${line}"
    fi
  fi
done <"${WORKDIR}/lib-versions.env"
