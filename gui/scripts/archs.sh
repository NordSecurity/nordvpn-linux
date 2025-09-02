#!/usr/bin/env bash

if [[ ! "${ARCHS:-}" ]]; then
    ARCHS=(
        i386
        amd64
        aarch64
    )
fi

# Key is one of ARCHS
declare -A ARCHS_DEB=(
    [i386]=i386
    [amd64]=amd64
    [arm64]=arm64
)

# Key is one of ARCHS
declare -A ARCHS_RPM=(
    [i386]=i386
    [amd64]=x86_64
    [arm64]=aarch64
)

export ARCHS
export ARCHS_DEB
export ARCHS_RPM
