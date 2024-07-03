#!/usr/bin/env bash

if [[ ! "${ARCHS:-}" ]]; then
ARCHS=(
    i386
    amd64
    armel
    armhf
    aarch64
)
fi

declare -A ARCHS_REVERSE=(
    [i386]=i386
    [i686]=i386
    [amd64]=amd64
    [x86_64]=amd64
    [armel]=armel
    [armv5l]=armel
    [armv5_eabi]=armel
    [armhf]=armhf
    [armhfp]=armhf
    [armv7_eabihf]=armhf
    [arm64]=aarch64
    [aarch64]=aarch64
)

# Key is one of ARCHS
declare -A ARCHS_DEB=(
    [i386]=i386
    [amd64]=amd64
    [armel]=armel
    [armhf]=armhf
    [aarch64]=arm64
)

# Key is one of ARCHS
declare -A ARCHS_RPM=(
    [i386]=i386
    [amd64]=x86_64
    [armel]=armv5l
    [armhf]=armhfp
    [aarch64]=aarch64
)

# Key is one of ARCHS
declare -A ARCHS_GO=(
    [i386]=386
    [amd64]=amd64
    [armel]=arm
    [armhf]=arm
    [aarch64]=arm64
)

# for .so files comming from libtelio, libdrop and libmoose
declare -A ARCHS_SO_REVERSE=(
    [i686]=i386
    [x86_64]=amd64
    [aarch64]=arm64
    [armv5]=armel
    [armv5_eabi]=armel
    [armv7hf]=armhf
    [armv7]=armhf
    [armv7_eabihf]=armhf
)

export ARCHS
export ARCHS_REVERSE
export ARCHS_DEB
export ARCHS_RPM
export ARCHS_GO
export ARCHS_SO_REVERSE
