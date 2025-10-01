#!/usr/bin/env bash
# CI_COMMIT_TAG variable is predefined if run by CI/CD. In case you want to build qa or
# prod builds locally, set these variables accordingly or use release/X.X.X branch for qa build
# dev and qa builds contain hash in version name
set -euxo pipefail

NAME=nordvpn
export NAME

HASH=$(git rev-parse --short HEAD)
export HASH

VERSION_PATTERN="^[0-9]+\.[0-9]+\.[0-9]+$"
if [[ "${CI_COMMIT_TAG:-}" =~ ${VERSION_PATTERN} ]]; then
  ENVIRONMENT="prod"
  export ENVIRONMENT

  REVISION=1
  export REVISION

  VERSION="${CI_COMMIT_TAG}"
  export VERSION
else
  ENVIRONMENT=${ENVIRONMENT:-"dev"}
  export ENVIRONMENT

  REVISION="${HASH}"
  export REVISION

  # '+' character is chosen because '_' is not allowed in .deb packages and '-' is not allowed in .rpm packages
  VERSION="$(find "${WORKDIR}"/contrib/changelog/prod -maxdepth 1 -type f -name '*.md' -printf '%f\n' | sed -E 's/_.*//; s/\.md$//' | sort -V | tail -n1)+${REVISION}"
  export VERSION
fi
