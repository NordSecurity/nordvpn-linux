#!/usr/bin/env bash
set -euo pipefail

PUSH_TO_REGISTRY=false
if [ "${1:-}" = "--push" ]; then
	PUSH_TO_REGISTRY=true
	shift
fi

if [ "$#" -ne 4 ]; then
	echo "Usage: $0 [--push] <image name> <version> <platforms> <dockerfile folder>"
	exit 1
fi

IMAGE_NAME=$1
IMAGE_VERSION=$2
PLATFORMS=$3
DOCKERFILE_FOLDER=$4
IMAGE="$IMAGE_NAME:$IMAGE_VERSION"

if [ "$PUSH_TO_REGISTRY" = true ]; then
	# never overwrite an image that is already in the registry
	if docker buildx imagetools inspect "$IMAGE" >/dev/null 2>&1; then
		echo "$IMAGE already exists in the registry, skipping"
		exit 0
	fi
	OUTPUT=--push
else
	# local docker can hold only one platform per image
	if [[ "$PLATFORMS" == *,* ]]; then
		echo "without --push only one platform can be built, e.g. linux/amd64"
		exit 1
	fi
	OUTPUT=--load
fi

# the default docker builder cannot build for several platforms at once,
# so we need to use a separate builder that can
# it will be created only on the first run
BUILDER=nordvpn-multiplatform
if ! docker buildx inspect "$BUILDER" >/dev/null 2>&1; then
	docker buildx create --name "$BUILDER" --driver docker-container
fi

echo "Building $IMAGE for $PLATFORMS"
docker buildx build \
	--builder "$BUILDER" \
	--platform "$PLATFORMS" \
	--tag "$IMAGE" \
	"$OUTPUT" \
	"$DOCKERFILE_FOLDER"

if [ "$PUSH_TO_REGISTRY" = true ]; then
	# list the digest of each platform's image
	docker buildx imagetools inspect "$IMAGE"
fi
