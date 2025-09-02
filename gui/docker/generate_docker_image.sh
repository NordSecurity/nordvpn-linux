#!/usr/bin/env bash
set -euox pipefail

# build a docker image if it doesn't exists
if [ "$#" -ne 4 ]; then
    echo "missing parameters. Expected: image name, version, platform(e.g. linux/amd64,linux/arm64) and folder containing Dockerfile"
    exit 1
fi

IMAGE_NAME=$1
IMAGE_VERSION=$2
PLATFORM=$3
DOCKER_FILE_FOLDER=$4

pushd ${DOCKER_FILE_FOLDER}
# check if multi-platform build is supported
if docker info -f '{{ .DriverStatus }}' | grep -vq containerd; then
    echo "Enable multi-platform build for docker https://docs.docker.com/build/building/multi-platform/#simple-multi-platform-build-using-emulation"
    exit 1
fi

echo Generating image $IMAGE_NAME:$IMAGE_VERSION for ${PLATFORM} from ${DOCKER_FILE_FOLDER}
docker build --platform ${PLATFORM} -t ${IMAGE_NAME}:${IMAGE_VERSION} .

popd

#TODO: upload the images