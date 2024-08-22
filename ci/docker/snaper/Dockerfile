# build image: docker build --no-cache -f ci/docker/snaper/Dockerfile --tag ghcr.io/nordsecurity/nordvpn-linux/snaper:0.0.1 .
# build snap: docker run -v "$PWD":/build -w /build ghcr.io/nordsecurity/nordvpn-linux/snaper:0.0.1 snapcraft --destructive-mode
# cleanup: sudo rm -rf ./parts ./stage ./prime

ARG RISK=stable
# ubuntu 22.04
ARG UBUNTU=jammy

FROM ubuntu:$UBUNTU as builder
ARG RISK
ARG UBUNTU
RUN echo "Building snapcraft:$RISK in ubuntu:$UBUNTU"

# Grab dependencies
RUN apt-get update \
        && apt-get dist-upgrade --yes \
        && apt-get install --yes curl jq squashfs-tools 

# Grab the core22 snap (which snapcraft uses as a base) from the stable channel
# and unpack it in the proper place.
RUN curl -L $(curl -H 'X-Ubuntu-Series: 16' 'https://api.snapcraft.io/api/v1/snaps/details/core22' | jq '.download_url' -r) --output core22.snap\
        && mkdir -p /snap/core22 \
        && unsquashfs -d /snap/core22/current core22.snap

# Grab the snapcraft snap from the $RISK channel and unpack it in the proper place.
RUN curl -L $(curl -H 'X-Ubuntu-Series: 16' 'https://api.snapcraft.io/api/v1/snaps/details/snapcraft?channel='$RISK | jq '.download_url' -r) --output snapcraft.snap \
        && mkdir -p /snap/snapcraft \
        && unsquashfs -d /snap/snapcraft/current snapcraft.snap

# Fix Python3 installation: Make sure we use the interpreter from the snapcraft snap:
RUN unlink /snap/snapcraft/current/usr/bin/python3 \
        && ln -s /snap/snapcraft/current/usr/bin/python3.* /snap/snapcraft/current/usr/bin/python3 \
        && echo /snap/snapcraft/current/lib/python3.*/site-packages >> /snap/snapcraft/current/usr/lib/python3/dist-packages/site-packages.pth

# Create a snapcraft runner (TODO: move version detection to the core of snapcraft).
RUN mkdir -p /snap/bin \
        && echo "#!/bin/sh" > /snap/bin/snapcraft \
        && snap_version="$(awk '/^version:/{print $2}' /snap/snapcraft/current/meta/snap.yaml | tr -d \')" && echo "export SNAP_VERSION=\"$snap_version\"" >> /snap/bin/snapcraft \
        && echo 'exec "$SNAP/usr/bin/python3" "$SNAP/bin/snapcraft" "$@"' >> /snap/bin/snapcraft \
        && chmod +x /snap/bin/snapcraft

# Multi-stage build, only need the snaps from the builder. Copy them one at a time so they can be cached.
FROM ubuntu:$UBUNTU
COPY --from=builder /snap/core22 /snap/core22
COPY --from=builder /snap/snapcraft /snap/snapcraft
COPY --from=builder /snap/bin/snapcraft /snap/bin/snapcraft

RUN mkdir -p /.cache && chmod -R 777 /.cache
RUN mkdir -p /.local/state && chmod -R 777 /.local/state

# Generate locale and install dependencies.
RUN apt-get update && apt-get dist-upgrade --yes \
        && apt-get install --yes snapd locales \
        && locale-gen en_US.UTF-8 \
        && apt-get install --yes git elfutils

# Set the proper environment.
ENV LANG="en_US.UTF-8"
ENV LANGUAGE="en_US:en"
ENV LC_ALL="en_US.UTF-8"
ENV PATH="/snap/bin:/snap/snapcraft/current/usr/bin:$PATH"
ENV SNAP="/snap/snapcraft/current"
ENV SNAP_NAME="snapcraft"
ENV SNAP_ARCH="amd64"
