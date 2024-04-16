#!/bin/bash
set -euo pipefail

source "${WORKDIR}/ci/env.sh"

libnord_version="0.5.4"
libnord_id="6385"

if [[ "${FEATURES:-""}" == *internal* ]]; then
	"${WORKDIR}"/ci/download_from_remote.sh \
		-O nord -p "${libnord_id}" -v "${libnord_version}" ${ARCH:+-a ${ARCH}} libnord.a
fi
