#!/usr/bin/env bash
set -euxo pipefail

find . -name '*.dart' \
    -not -name '*.tailor.dart' \
    -not -name '*.freezed.dart' \
    -not -name '*.g.dart' \
    -not -path './lib/pb/*' \
    -not -path './.dart_tool/*' \
    -print0 \
    | xargs -0 dart format