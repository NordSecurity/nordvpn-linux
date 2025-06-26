# Table of contents
1. [How to setup the development environment](#how-to-setup-the-development-environment)
1. [Building](#building)
1. [Conditional compilation](#conditional-compilation)
1. [Building scripts/tooling](#building-scripts-tooling)
1. [Docker images](#docker-images)
# How to setup the development environment
Please follow the instructions in the following step for setting up the development environment. (Note: This process was tested on a virtual machine with Ubuntu 22.04):
1. Install [Go 1.22](https://go.dev/doc/install). Note that newer versions will not work if
   either `telio` or `drop` are included.
1. Install [Mage](https://github.com/magefile/mage#installation). Even though currently buiding
   using bash scripts directly works well, mage is the recommended way to go.
1. Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

## Docker builds
Mage targets for building with docker are named `build:xxxDocker`. The only extra dependency for
docker builds is [Docker](https://docs.docker.com/engine/install/ubuntu/).

### Idempotent docker builds
By default, mage targets will always pull docker images from the registry. If below entry is present in `.env` file, images will be pulled only if they are not present on the host system:
```
IDEMPOTENT_DOCKER=1
```

## Native builds
Mage targets for building natively are named `build:xxx` and they don't have `Docker` suffix.

This section contains full list of dependencies required in order to build the full application
package (deb, rpm, snap).

Native dependencies (Debian/Ubuntu):
* `dpkg-dev`
* `elfutils`
* `gcc`
* `gettext`
* `make`
* `pkg-config`
* `libnl-genl-3-dev`
* `libcap-ng-dev`
* `libxml2-dev`
* `libsqlite3-dev`

Go utils:
* [nfpm](https://github.com/goreleaser/nfpm)
* [go-licenses](https://github.com/google/go-licenses)

Rust:
* [rust](https://www.rust-lang.org/tools/install)

Note: when built natively, application will depend on the GLIBC version that is available on the
build environment. This means that in most cases application will not work if shipped to another
environment that has a lower GLIBC version than the one application was built on.

# Conditional compilation
Some parts of the app functionality can be conditionally compiled. See the list in
[.env.sample](.env.sample).

# Building scripts/tooling
Binaries found in `cmd/<binary>/main.go` are not shipped to the users and are built with:
```sh
go build -o "bin/<binary>" "cmd/<binary>"
```
- checkelf (ensures that nordvpn/nordvpnd executables don't exceed glibc version)
- downloader (downloads files from CDN used when building deb/rpm packages)

# Docker images
A list of docker images can be found in [ci/docker](ci/docker)

Images are stored in `ghcr.io/nordsecurity/nordvpn-linux` registry.

The building and tagging can be done in a single command like this:
```sh
docker build -t <registry>/<image>[:tag] ci/docker/<image>
```
The pushing can be done with:
```sh
docker push <registry>/<image>[:tag]
```
