#!/usr/bin/env bash
# CI_COMMIT_TAG variable is predefined if run by CI/CD. In case you want to build qa or
# prod builds locally, set these variables accordingly or use release/X.X.X branch for qa build
# dev and qa builds contain hash in version name
set -euxo pipefail

# if inside of docker container on CI || if inside of docker container on the host
if grep docker /proc/self/cgroup || [ "$(< /proc/self/cgroup)" == "0::/" ]; then
  # required for docker mounts to work correctly with mage targets
  git config --global --add safe.directory "${WORKDIR}"
fi

NAME=nordvpn
export NAME

COVERDIR="covdatafiles"
export COVERDIR

HASH=$(git rev-parse --short HEAD)
export HASH

LIBNORD_VERSION="0.5.1"
export LIBNORD_VERSION

OPENVPN_VERSION="2.5.8"
export OPENVPN_VERSION

OPENSSL_VERSION="1.1.1t"
export OPENSSL_VERSION

LZO_VERSION="2.10"
export LZO_VERSION

TUNNELBLICK_VERSION="v3.8.8"
export TUNNELBLICK_VERSION

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
  VERSION="$(git tag -l --sort=-v:refname | grep "^[0-9]\+\.[0-9]\+\.[0-9]\+$" | head -1)+${REVISION}"
  export VERSION
fi
