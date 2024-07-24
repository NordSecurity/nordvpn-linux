#!/bin/bash
set -euxo pipefail

# Excluded packages are directly related to C packages, therefore they
# complicate the compilation process. It is fine to exclude them for
# testing/development purposes.
# If the tests fail because of C dependencies, the failing packages
# must be separated from the rest of the code and added here.
excluded_packages="moose\|cmd\/daemon\|telio\|daemon\/vpn\/openvpn"
excluded_packages=$excluded_packages"\|meshnet\/mesh\/nordlynx\|fileshare\/drop"
excluded_packages=$excluded_packages"\|events\/moose"
excluded_categories="root,link,firewall,route,file,integration"

# In case 'full' was specified, do not exclude anything and run
# everything
if [ "${1:-""}" = "full" ]; then
	# Apply moose patch in case compiling with moose
	git apply "${WORKDIR}"/contrib/patches/add_moose.diff
	function revert_moose_patch {
		git apply -R "${WORKDIR}"/contrib/patches/add_moose.diff
	}
	trap revert_moose_patch EXIT

	excluded_packages="thisshouldneverexist"
	excluded_categories="root,link"
fi

# Execute tests in all the packages except the excluded ones

# SC2046 is disabled so that list of packages is not treated
# as a single argument for 'go test'

mkdir -p "${WORKDIR}"/coverage/unit

ARCH="amd64"

# for compile-time
export LIBRARY_PATH="${WORKDIR}/bin/deps/lib/${ARCH}/latest"
# for run-time
export LD_LIBRARY_PATH="${WORKDIR}/bin/deps/lib/${ARCH}/latest"

# shellcheck disable=SC2046
go test -tags internal -v -race $(go list ./... | grep -v "${excluded_packages}") \
	-coverprofile "${WORKDIR}"/coverage.txt \
	-exclude "${excluded_categories}" \
	-args -test.gocoverdir="${WORKDIR}/coverage/unit"

# Display code coverage report
go tool cover -func="${WORKDIR}"/coverage.txt

if [ "${1:-""}" = "full" ]; then
	# "gocover-cobertura" is used for test coverage visualization in the diff view.
	gocover-cobertura < "$WORKDIR"/coverage.txt > coverage.xml
fi
