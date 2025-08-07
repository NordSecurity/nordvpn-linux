FROM ghcr.io/canonical/snapcraft:8_core22@sha256:f24fc9cd8a1f394b48dcd35f8e3f31b127023a4bb79134afeb2728b012bf2007

ARG DEBIAN_FRONTEND=noninteractive

# NOTE: Don't remove apt sources after installation here because snapcraft needs them.
RUN apt-get update && \
  apt-get install --no-install-recommends -y \
  elfutils \
  git

# This is a work around craftctl not available in docker
ENV PATH="/usr/libexec/snapcraft:$PATH"

# override the entrypoint from base image, because
# it runs long-running process which we don't need
ENTRYPOINT []
