#!/bin/bash
set -euo

source "${WORKDIR}"/ci/env.sh

go tool covdata percent -i="${COVERDIR}" 
go tool covdata textfmt -i="${COVERDIR}" -o coverage.txt
total_coverage=$(go tool cover -func=coverage.txt | grep 'total:' | awk '{print $3}')

echo "Total coverage: $total_coverage"
