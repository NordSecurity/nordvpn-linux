#!/usr/bin/env bash
set -euxo pipefail

# Runs snap's review-tools against the built snap package, the same checks
# performed by the Snap Store on upload. Fails the build if review-tools
# reports any warning or error, e.g. plugs that require manual review.

source "${WORKDIR}"/ci/env.sh

ARCH=$([ "${ARCH}" == "aarch64" ] && echo arm64 || echo "${ARCH}")

mapfile -t FILES < <(find "${WORKDIR}/dist/app/snap/" -type f -name "*_${ARCH}.snap")

if [[ ${#FILES[@]} -ne 1 ]]; then
  echo "ERROR: expected exactly 1 snap file, found ${#FILES[@]}: ${FILES[*]}" >&2
  exit 1
fi

echo "review snap package: ${FILES[0]}"

# review-tools is a confined snap and can only access files under $HOME,
# so copy the package there before handling it
TMP_SNAP="$(mktemp -p "${HOME}" --suffix=.snap)"
trap 'rm -f "${TMP_SNAP}"' EXIT
cp "${FILES[0]}" "${TMP_SNAP}"

review-tools.snap-review "${TMP_SNAP}"

echo "DONE: review-tools found no issues"
