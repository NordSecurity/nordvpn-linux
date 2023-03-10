FROM debian:bullseye-slim

# Install global dependencies
RUN apt-get update && apt-get install -y \
  nodejs npm yarn \
  curl \
  bash \
  wget \
  git \
  python3-pip \
  openjdk-11-jre

# Create python symlink
RUN ln -s $(which python3) /usr/bin/python

# Add golang binaries
COPY --from=golang:latest /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

