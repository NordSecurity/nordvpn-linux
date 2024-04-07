#!/bin/bash
set -euxo pipefail

source "${WORKDIR}"/ci/env.sh
source "${WORKDIR}"/ci/archs.sh


# Since race detector has huge performance price and it works only on amd64 and does not
# work with pie executables, its enabled only for development builds.
# shellcheck disable=SC2153
if [ "${ENVIRONMENT}" = "dev" ]; then
	[ "${ARCH}" = "amd64" ] && [ "${RACE_DETECTOR_ENABLED:-""}" == "1" ] && BUILDMODE="-race"
else
	BUILDMODE="-buildmode=pie"
fi

ldflags="-X 'main.Version=${VERSION}' \
	-X 'main.Environment=${ENVIRONMENT}' \
	-X 'main.Hash=${HASH}' \
	-X 'main.Arch=${ARCH}' \
	-X 'main.PackageType=${PACKAGE:-deb}' \
	-X 'main.Salt=${SALT}' \
	-X 'main.FirebaseToken=${FIREBASE_TOKEN:-""}'"

declare -A names_map=(
	[cli]=nordvpn
	[daemon]=nordvpnd
	[downloader]=downloader
	[pulp]=pulp
	[fileshare]=nordfileshare
	[norduser]=norduserd
)

declare -A cross_compiler_map
declare -A cross_compiler_map_openwrt

# shellcheck disable=SC2034
cross_compiler_map=(
    [i386]=i686-linux-gnu-gcc
    [amd64]=x86_64-linux-gnu-gcc
    [armel]=arm-linux-gnueabi-gcc
    [armhf]=arm-linux-gnueabihf-gcc
    [aarch64]=aarch64-linux-gnu-gcc
)

cross_compiler_map_openwrt=(
    [amd64]="x86_64-openwrt-linux-musl-gcc"
  	[aarch64]="aarch64-openwrt-linux-musl-gcc"
)

# Required by Go when cross-compiling
export CGO_ENABLED=1

if [[ "${OS}" == "openwrt" ]]; then
	mkdir -p "$GO_BUILD_DIR/bin" "$GO_BUILD_CACHE_DIR" "$GO_MOD_CACHE_DIR" "$GO_BUILD_BIN_DIR"
else
	GOARCH="${ARCHS_GO["${ARCH}"]}"
	export GOARCH
fi

# C compiler flags for binary hardening.
export CGO_CFLAGS="${CGO_CFLAGS:-""} -g -O2 -D_FORTIFY_SOURCE=2"

# These C linker flags get appended to the ones specified in the source code
export CGO_LDFLAGS="${CGO_LDFLAGS:-""} -Wl,-z,relro,-z,now"

# Required by Go when cross-compiling to 32bit ARM architectures
[ "${ARCH}" == "armel" ] && export GOARM=5
[ "${ARCH}" == "armhf" ] && export GOARM=7

# In order to enable additional features, provide `FEATURES` environment variable
tags="${FEATURES:-"telio drop"}"

if [[ $tags == *"moose"* ]]; then
	# Set correct events domain in case compiling with moose
	if [[ "${ENVIRONMENT}" == "prod" ]]; then
		events_domain="${EVENTS_PROD_DOMAIN}"
	else
		events_domain="${EVENTS_STAGING_DOMAIN}"
	fi

	ldflags="${ldflags} \
		-X 'main.EventsDomain=${events_domain:-""}' \
		-X 'main.EventsSubdomain=${EVENTS_SUBDOMAIN:-""}'"

	# Apply moose patch in case compiling with moose
	git apply "${WORKDIR}"/contrib/patches/add_moose.diff || \
		# If applying fails try reverting and applying again
		(git apply -R "${WORKDIR}"/contrib/patches/add_moose.diff && \
		git apply "${WORKDIR}"/contrib/patches/add_moose.diff)
	function revert_moose_patch {
		cd "${WORKDIR}"
		git apply -R "${WORKDIR}"/contrib/patches/add_moose.diff
	}
	trap revert_moose_patch EXIT
fi

for program in ${!names_map[*]}; do # looping over keys
	pushd "${WORKDIR}/cmd/${program}"
	if [[ "${OS}" == "openwrt" ]]; then
		CC="${cross_compiler_map_openwrt[${ARCH}]}" \
			go build ${BUILD_FLAGS:+"${BUILD_FLAGS}"} ${BUILDMODE:-} -tags "${tags}" \
				-ldflags "-linkmode=external ${ldflags}" \
				-o "${WORKDIR}/bin/${ARCH}/${names_map[${program}]}"
		cp -r "${WORKDIR}/bin/${ARCH}/${names_map[${program}]}" "${GO_BUILD_BIN_DIR}"
	else
		# BUILDMODE can be no value and `go` does not like empty parameter ''
		# this is why surrounding double quotes are removed to not cause empty parameter i.e. ''
		# shellcheck disable=SC2086
		CC="${cross_compiler_map[${ARCH}]}" \
			go build ${BUILD_FLAGS:+"${BUILD_FLAGS}"} ${BUILDMODE:-}-tags "${tags}" \
				-ldflags "-linkmode=external ${ldflags}" \
				-o "${WORKDIR}/bin/${ARCH}/${names_map[${program}]}"
	fi
	popd
done
