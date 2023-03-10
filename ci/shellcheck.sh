#!/bin/sh
set -euo

# Without + at the end, find will return 0 even when exec fails.
for dir in ci contrib; do
	find "${CI_PROJECT_DIR}"/"${dir}" -type f -name "*.sh" \
		-exec shellcheck --exclude=SC1090,SC2016,SC2207 -x "{}" +
done
