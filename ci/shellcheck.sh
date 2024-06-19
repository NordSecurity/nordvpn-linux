#!/bin/sh
set -euxo

# Without + at the end, find will return 0 even when exec fails.
for dir in ci contrib; do
	find "${WORKDIR}"/"${dir}" -type f -name "*.sh" \
		-exec shellcheck -x "{}" +
done