#!/usr/bin/env bash
set -euxo pipefail

# This controls if a backup file is created when the pubspec.yaml file is changed
# this can be useful for local builds not to have the file marked as changed
USE_BACKUP_FILE=${1:-true}

VERSION_PATTERN="^[0-9]+\.[0-9]+\.[0-9]+$"
if [[ "${CI_COMMIT_TAG:-}" =~ ${VERSION_PATTERN} ]]; then
  # in this case expecting version to be e.g. 1.2.3
  VERSION="${CI_COMMIT_TAG}"
else
  # here add commit id to the version what is set as a last tag or default value
  VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.0.1")
  REVISION=$(git rev-parse --short HEAD)
  VERSION="${VERSION}+${REVISION}"
fi

# Extract current version number from pubspec.yaml
CURRENT_VERSION=$(grep 'version:' "${WORKDIR}"/gui/pubspec.yaml | cut -d ' ' -f 2 | cut -d '+' -f 1)

echo "Current pubspec.yaml: Version=${CURRENT_VERSION}"
echo "Given/determined version: Version=${VERSION}"

# Compare values and only replace file content when are different
if [[ "${CURRENT_VERSION}" != "${VERSION}" ]]; then
  if [ "${USE_BACKUP_FILE}" = "true" ]; then
    # store the previous version to revert
    cp -f "${WORKDIR}"/gui/pubspec.yaml "${WORKDIR}"/gui/pubspec.yaml.bak
  fi

  echo "Updating pubspec.yaml with new version."
  sed -i "s/^version:.*/version: ${VERSION}/" "${WORKDIR}"/gui/pubspec.yaml
else
  echo "No changes needed. Version is up to date."
fi
