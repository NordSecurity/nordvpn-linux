#!/usr/bin/env bash
# CI_COMMIT_TAG variable is predefined if run by CI/CD. In case you want to build qa or
# prod builds locally, set these variables accordingly or use release/X.X.X branch for qa build
# dev and qa builds contain hash in version name

# VERSION_DATE is expected by nfpm in package template.yaml. For the same version, version date 
# should be the same - to have reproducible builds/packages.
# Version date-time is taken from git version tag timestamp or changelog file last modification 
# timestamp, or, if those not available, use git HEAD commit timestamp.

set -euxo pipefail

NAME=nordvpn
HASH=$(git rev-parse --short HEAD)

format_date_utc() {
  local date_input="$1"
  # format date to ISO 8601 UTC format
  echo "$date_input" | xargs -I{} date -u -d "{}" +"%Y-%m-%dT%H:%M:%SZ"
}

get_head_commit_date() {
  format_date_utc "$(git log -1 --format=%aI HEAD)"
}

VERSION_PATTERN="^[0-9]+\.[0-9]+\.[0-9]+$"
if [[ "${CI_COMMIT_TAG:-}" =~ ${VERSION_PATTERN} ]]; then
  ENVIRONMENT="prod"
  REVISION=1
  VERSION="${CI_COMMIT_TAG}"

  if git rev-parse "${CI_COMMIT_TAG}" >/dev/null 2>&1; then
    VERSION_DATE="$(format_date_utc "$(git log -1 --format=%aI "${CI_COMMIT_TAG}")")"
  else
    echo "Warning: Git tag '${CI_COMMIT_TAG}' not found, using git HEAD instead"
    VERSION_DATE="$(get_head_commit_date)"
  fi
else
  ENVIRONMENT=${ENVIRONMENT:-"dev"}
  REVISION="${HASH}"

  # '+' character is chosen because '_' is not allowed in .deb packages and '-' is not allowed in .rpm packages
  CHLOG_VERSION="$(find "${WORKDIR}"/contrib/changelog/prod -maxdepth 1 -type f -name '*.md' -printf '%f\n' | sed -E 's/_.*//; s/\.md$//' | sort -V | tail -n1)"
  VERSION="${CHLOG_VERSION}+${REVISION}"

  CHLOG_FILE="${WORKDIR}/contrib/changelog/prod/${CHLOG_VERSION}.md"
  if [[ -f "${CHLOG_FILE}" ]]; then
    VERSION_DATE="$(format_date_utc "$(stat -c '%y' "${CHLOG_FILE}")")"
  else
    echo "Warning: Changelog file '${CHLOG_FILE}' not found, using git HEAD commit date instead"
    VERSION_DATE="$(get_head_commit_date)"
  fi
fi

export NAME HASH ENVIRONMENT REVISION VERSION VERSION_DATE
