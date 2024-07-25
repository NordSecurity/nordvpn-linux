#!/bin/bash
set -euo pipefail

libnord_version="0.5.8"
libnord_id="6385"

if [[ "${FEATURES:-""}" == *internal* ]]; then
	"${WORKDIR}"/ci/download_from_remote.sh \
		-O nord -p "${libnord_id}" -v "${libnord_version}" ${ARCH:+-a ${ARCH}} libnord.a
fi
