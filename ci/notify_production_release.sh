#!/usr/bin/env bash
set -euxo

WEBHOOK=${SLACK_WEBHOOK_URL}
CHANNELS="newsfeed"
USER="Everyone's favorite penguin"

file=("${CI_PROJECT_DIR}/contrib/changelog/${ENVIRONMENT}/${VERSION}"_*.md)
if [[ ! -f "${file[0]}" ]]; then
    echo "Error! Release notes not found" 1>&2
    exit 1
fi

FILE=$(awk '{printf "%s\\n", $0}' "${file[0]}")

AT="--------------------------------\n
$(echo "$SLACK_TAGGEES" | jq -r 'join(" ")')\n"
TEMPLATE="*Linux NordVPN app version $VERSION is released!*\n
--------------------------------\n
*Release notes:*\n"

for CHANNEL in $CHANNELS
do
    "${CI_PROJECT_DIR}"/ci/slack.sh "${WEBHOOK}" "${CHANNEL}" "${USER}" "${TEMPLATE}${FILE}${AT}"
done
