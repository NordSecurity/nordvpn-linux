FROM ghcr.io/canonical/snapcraft:8_core22@sha256:d6d2d32a1f87da38756f96145fc05683e60e612e08cc264eeba4cb824dbd4631

ARG DEBIAN_FRONTEND=noninteractive

# NOTE: Don't remove apt sources after installation here because snapcraft needs them.
RUN apt-get update && \
  apt-get install --no-install-recommends -y \
  elfutils=0.186-1ubuntu0.1 \
  git=1:2.34.1-1ubuntu1.12

# snapcraft is running in privileged user
# mode which does not match git repo owner
RUN git config --global --add safe.directory "/opt" && \
  git config --global --add safe.directory "/opt/parts/nordvpn/build"

WORKDIR /opt

# override the entrypoint from base image, because
# it runs long-running process which we don't need
ENTRYPOINT []
