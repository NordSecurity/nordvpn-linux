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

  # version date should be always the same
  VERSION_DATE="$(git log -1 --format=%aI ${CI_COMMIT_TAG} | xargs -I{} date -u -d "{}" +"%Y-%m-%dT%H:%M:%SZ")"
  export VERSION_DATE
else
  ENVIRONMENT=${ENVIRONMENT:-"dev"}
  export ENVIRONMENT

  REVISION="${HASH}"
  export REVISION

  # '+' character is chosen because '_' is not allowed in .deb packages and '-' is not allowed in .rpm packages
  CHLOG_VERSION="$(find "${WORKDIR}"/contrib/changelog/prod -maxdepth 1 -type f -name '*.md' -printf '%f\n' | sed -E 's/_.*//; s/\.md$//' | sort -V | tail -n1)"
  VERSION="${CHLOG_VERSION}+${REVISION}"
  export VERSION

  # version date should be always the same
  chlog_file="${WORKDIR}/contrib/changelog/prod/${CHLOG_VERSION}.md"
  VERSION_DATE="$(stat -c '%y' "${chlog_file}" | xargs -I{} date -u -d "{}" +"%Y-%m-%dT%H:%M:%SZ")"
  export VERSION_DATE
fi
