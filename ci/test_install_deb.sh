#!/usr/bin/env bash
set -euxo

default_tool=curl

apt-get update && apt-get -y install apt-utils "${1-$default_tool}"
mkdir -p "${REPO_DIR}" && cp -t "${REPO_DIR}" "${WORKDIR}"/dist/app/deb/*.deb
cd "${REPO_DIR}" && apt-ftparchive packages . > Packages
"${WORKDIR}"/test/qa/install.sh -n -b "" -k "" -d "[trusted=true] file:///$REPO_DIR/" -v "./"
