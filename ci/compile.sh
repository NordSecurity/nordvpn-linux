#!/bin/bash
set -euox

source "${CI_PROJECT_DIR}"/ci/env.sh
source "${CI_PROJECT_DIR}"/ci/archs.sh

"${CI_PROJECT_DIR}"/ci/check_dependencies.sh

# Since race detector has huge performance price and it works only on amd64 and does not
# work with pie executables, its enabled only for development builds.
branch="${CI_COMMIT_REF_NAME:=$(git describe --contains --all HEAD)}"
# shellcheck disable=SC2153
[ "${branch}" != "master" ] && [ "${ENVIRONMENT}" = "dev" ] && [ "${ARCH}" = "amd64" ] && BUILDMODE="-race" || BUILDMODE="-buildmode=pie"

if [[ "${ENVIRONMENT}" == "prod" ]]; then
	EVENTS_DOMAIN="${EVENTS_PROD_DOMAIN}"
else
	EVENTS_DOMAIN="${EVENTS_STAGING_DOMAIN}"
fi

LDFLAGS="-X 'main.Version=${VERSION}' \
	-X 'main.Environment=${ENVIRONMENT}' \
	-X 'main.Hash=${HASH}' \
	-X 'main.Arch=${ARCH}' \
	-X 'main.PackageType=${PACKAGE:-deb}' \
	-X 'main.Salt=${SALT}' \
	-X 'main.EventsDomain=${EVENTS_DOMAIN}' \
	-X 'main.EventsSubdomain=${EVENTS_SUBDOMAIN}' \
	-X 'main.FirebaseToken=${FIREBASE_TOKEN:-""}'"

declare -A names_map=(
	[cli]=nordvpn
	[daemon]=nordvpnd
	[downloader]=downloader
	[pulp]=pulp
)

# shellcheck disable=SC2034
declare -A cross_compiler_map=(
    [i386]=i686-linux-gnu-gcc
    [amd64]=x86_64-linux-gnu-gcc
    [armel]=arm-linux-gnueabi-gcc
    [armhf]=arm-linux-gnueabihf-gcc
    [aarch64]=aarch64-linux-gnu-gcc
)

# Required by Go when cross-compiling
export CGO_ENABLED=1
GOARCH="${ARCHS_GO["${ARCH}"]}"
export GOARCH="${GOARCH}"

# C compiler flags for binary hardening.
export CGO_CFLAGS="-g -O2 -D_FORTIFY_SOURCE=2"

# These C linker flags get appended to the ones specified in the source code
export CGO_LDFLAGS="-Wl,-z,relro,-z,now"

# Required by Go when cross-compiling to 32bit ARM architectures
[ "${ARCH}" == "armel" ] && export GOARM=5
[ "${ARCH}" == "armhf" ] && export GOARM=7

# In order to enable additional features, provide `FEATURES` environment variable
tags="${FEATURES:-"telio drop"}"

# Only compile fileshare daemon if feature flag is present
if [[ $tags == *"drop"* ]]; then 
	names_map[fileshare]+=nordfileshared
fi

# Apply moose patch in case compiling with moose
if [[ $tags == *"moose"* ]]; then 
	git apply "${CI_PROJECT_DIR}"/contrib/patches/add_moose.diff
	function revert_moose_patch {
		cd "${CI_PROJECT_DIR}"
		git apply -R "${CI_PROJECT_DIR}"/contrib/patches/add_moose.diff
	}
	trap revert_moose_patch EXIT
fi

for program in ${!names_map[*]}; do # looping over keys
	pushd "${CI_PROJECT_DIR}/cmd/${program}"
	CC="${cross_compiler_map[${ARCH}]}" \
		go build "${BUILDMODE}" -tags "${tags}" \
		-ldflags "-linkmode=external ${LDFLAGS}" \
		-o "${CI_PROJECT_DIR}/bin/${ARCH}/${names_map[${program}]}"
	popd
done

