#!/bin/bash
set -euxo pipefail

# parameters: 
# $1: deb / rpm
# $2: target dir (default /tmp/test-reproducible)

PACKAGE_TYPE=${1:-"deb"}
TARGET_PATH=${2:-"/tmp/test-reproducible"}

# number of iterations 
LOOP_START=1
LOOP_COUNT=10

# initialize target path
if [ -d "$TARGET_PATH" ]; then
    rm -rf "$TARGET_PATH"
fi
mkdir -p "$TARGET_PATH"

BASE_DIR=${WORKDIR:-$(pwd)}
DIFF_COUNT=0
TARGET_PACKAGE_PREV_SIZE=0

echo "BASE_DIR: $BASE_DIR"

for idx in $(seq $LOOP_START $LOOP_COUNT); do

    echo "~~~~~~~~ $idx ~~~~~~~~~~~"

    TARGET_DIR=$TARGET_PATH/t$idx
    echo "TARGET DIR: $TARGET_DIR"
    mkdir -p "$TARGET_DIR"

    echo "BUILD [$PACKAGE_TYPE][$ARCH] PACKAGE..."
    "$BASE_DIR"/ci/nfpm/build_packages_resources.sh "$PACKAGE_TYPE"

    PACKAGE_FILE=$(find "$BASE_DIR/dist/app" -name "*.$PACKAGE_TYPE" -type f)
    if [ -z "$PACKAGE_FILE" ]; then
        echo "Package file not found!"
        exit 1
    fi
    echo "BUILT PACKAGE: $PACKAGE_FILE"
    PACKAGE_NAME=$(basename -- "$PACKAGE_FILE")

    echo "COPY [$PACKAGE_TYPE][$ARCH] PACKAGE..."
    cp "$PACKAGE_FILE" "$TARGET_DIR"
    TARGET_PACKAGE_SIZE=$(stat --printf="%s" "$TARGET_DIR/$PACKAGE_NAME")

    echo "PACKAGE SIZE: $TARGET_PACKAGE_SIZE"

    if [ "$TARGET_PACKAGE_PREV_SIZE" -eq "0" ]; then
        TARGET_PACKAGE_PREV_SIZE=$TARGET_PACKAGE_SIZE
    fi
    if [ "$TARGET_PACKAGE_PREV_SIZE" != "$TARGET_PACKAGE_SIZE" ]; then
        echo " ::: got different package size!!! previous package size: [$TARGET_PACKAGE_PREV_SIZE]"
        ((DIFF_COUNT++))
    fi
    TARGET_PACKAGE_PREV_SIZE=$TARGET_PACKAGE_SIZE

    echo "Some sleep..."
    sleep 2
    echo ""

done

echo "ALL PACKAGES:"
for f in $(find "$TARGET_PATH" -name "*.$PACKAGE_TYPE" -type f | sort); do
    ls -l "$f"
done

RC=0
if [ "$DIFF_COUNT" -gt "0" ]; then
    echo "Builds are non-reproducable, got diffs [$DIFF_COUNT] out of [$LOOP_COUNT] tries!"
    RC=1
else
    echo "All packages are the same size - builds are reproducable! Tried [$LOOP_COUNT] times."
fi

echo "DONE."
exit $RC
