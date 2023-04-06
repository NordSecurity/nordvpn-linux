#!/bin/bash
set -euo

source "${CI_PROJECT_DIR}"/ci/env.sh

go tool covdata percent -i=./coverage/unit,./"${COVERDIR}" -o coverage.txt
total_coverage=$(go tool cover -func=coverage.txt | grep 'total:' | awk '{print $3}')

echo "Total coverage: $total_coverage"
