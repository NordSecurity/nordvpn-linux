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

OPENVPN_VERSION="2.5.8"
export OPENVPN_VERSION
OPENVPN_SHA256SUM="a6f315b7231d44527e65901ff646f87d7f07862c87f33531daa109fb48c53db2"
export OPENVPN_SHA256SUM

OPENSSL_VERSION="1.1.1t"
export OPENSSL_VERSION
OPENSSL_SHA256SUM="8dee9b24bdb1dcbf0c3d1e9b02fb8f6bf22165e807f45adeb7c9677536859d3b"
export OPENSSL_SHA256SUM

LZO_VERSION="2.10"
export LZO_VERSION
LZO_SHA256SUM="c0f892943208266f9b6543b3ae308fab6284c5c90e627931446fb49b4221a072"
export LZO_SHA256SUM

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
