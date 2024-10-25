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
excluded_packages=$excluded_packages"\|pb\|magefiles"
excluded_categories="root,link,firewall,route,file,integration"

tags="internal"

# In case 'full' was specified, do not exclude anything and run
# everything
if [ "${1:-""}" = "full" ]; then
	# Apply moose patch in case compiling with moose
	source "${WORKDIR}"/ci/add_moose.sh

	excluded_packages="thisshouldneverexist"
	excluded_categories="root,link"
	tags="internal,moose"
fi

# Execute tests in all the packages except the excluded ones

# SC2046 is disabled so that list of packages is not treated
# as a single argument for 'go test'

mkdir -p "${WORKDIR}"/coverage/unit

# single architecture for tests
export LD_LIBRARY_PATH="${WORKDIR}/bin/deps/lib/amd64/latest"

# shellcheck disable=SC2046
go test -tags "$tags" -v -race $(go list -tags "$tags" -buildvcs=false ./... | grep -v "${excluded_packages}") \
	-coverprofile "${WORKDIR}"/coverage.txt \
	-exclude "${excluded_categories}" \
	-args -test.gocoverdir="${WORKDIR}/coverage/unit"

grep -v "$excluded_packages" < "${WORKDIR}/coverage.txt" > "${WORKDIR}/cov.txt"
mv "${WORKDIR}/cov.txt" "${WORKDIR}/coverage.txt"

# Display code coverage report
go tool cover -func="${WORKDIR}"/coverage.txt

if [ "${1:-""}" = "full" ]; then
	# "gocover-cobertura" is used for test coverage visualization in the diff view.
	GOFLAGS=-tags="${tags}" gocover-cobertura < "$WORKDIR"/coverage.txt > coverage.xml
fi
