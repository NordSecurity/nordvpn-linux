# Table of contents
1. [How to setup the development environment](#how-to-setup-the-development-environment)
1. [Building](#building)
1. [Testing](#testing)
1. [Branches](#branches)
1. [CI/CD](#cicd)
1. [Linting](#linting)
# How to setup the development environment
Please follow the instructions in the following step for setting up the development environment. (Note: This process was tested on a virtual machine with Ubuntu 22.04):
1. Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
1. Install [Go 1.20+](https://go.dev/doc/install)
1. Install [Rust 1.64+](https://www.rust-lang.org/tools/install)
1. Install [Mage](https://github.com/magefile/mage#installation).
    If you are installing `Mage` using GOPATH, make sure that your `$GOPATH` variable is set correctly.
1. Install [Docker](https://docs.docker.com/engine/install/ubuntu/).
    After the installation is completed, add the user to the `docker` group by running `sudo usermod -aG docker $USER` on the terminal.
    Restart desktop session by relogging in order to the make changes effective.
1. Optionally setup environment for testing application across various distros.
    1. Install [Vagrant](https://www.vagrantup.com/docs/installation).
    1. Install [Vagrant-Libvirt](https://github.com/vagrant-libvirt/vagrant-libvirt#installation).
    How to use vagrant (from the project root):
    ```sh
    vagrant up fedora36
    vagrant ssh fedora36
    vagrant destroy fedora36 # when the work is done
    ```
    List of available distros:
    - debian10
    - debian11
    - fedora35
    - fedora36
    - ubuntu18
    - ubuntu20
    - ubuntu22
    Virtual machine mounts `dist/app` directory from the host at `/vagrant`.
1. (optional) Install [protoc](https://grpc.io/docs/languages/go/quickstart/#prerequisites) to be able to compile protobuf files.
    1. Might need to rename `protoc-gen-go-grpc` binary to `protoc-gen-go_grpc` to work.
1. (optional) Install Act to run Github jobs locally https://github.com/nektos/act
1. Run `mage` to discover and execute build targets.
    1. To use non-Docker targets please refer to `ci/docker/*/Dockerfile` Dockerfiles for necessary dependencies to be installed.
# Building
## Building with mage and docker
Convenient way for building the application is available using the [mage](https://github.com/magefile/mage#installation)
and optionally having a [docker](https://docs.docker.com/engine/install/) daemon running.
#### Building all binaries natively
```sh
mage build:binaries
```
#### Building all binaries in docker
```sh
mage build:binariesDocker
```
## Building parts of the application manually
Below steps can be used to build the app without mage and build scripts.
### Dependencies
#### Compile time
- Go 1.22+
- libxml2
#### Runtime
- iptables
- iproute2
### Building nordvpn cli utility
Building requires injecting the following variables via linker:
```sh
# Used to derive keys when encrypting/decrypting configuration files.
# Changing the Salt means changing the keys, which means losing access
# to previously encrypted files. While this is fine during development,
# production builds should never change the Salt between versions.
main.Salt
# Used by `nordvpn --version` command.
# Usually taken from the git commit tag.
main.Version
# Used to implement feature toggles.
# It's either dev or prod.
main.Environment
# Used by `nordvpn --version` command.
# Equal to the git commit hash the build was made from.
main.Hash
```
#### Development builds
```sh
LDFLAGS="-X main.Version=${} \
	-X main.Environment=${} \
	-X main.Hash=${} \
	-X main.Salt=${} \
  go build -ldflags "${LDFLAGS}" \
  -o "bin/nordvpn" "cmd/cli"
```
#### Production (hardened) builds
Since Go 1.17 building position independent executables requires using C linker.
C linker also allows adding additional protection known as hardening:
- [Fortification](https://www.redhat.com/en/blog/enhance-application-security-fortifysource)
- [Relocation](https://www.redhat.com/en/blog/hardening-elf-binaries-using-relocation-read-only-relro)
```sh
CGO_CFLAGS="-g -O2 -D_FORTIFY_SOURCE=2" \
CGO_LDFLAGS="-Wl,-z,relro,-z,now" \
LDFLAGS="-X main.Version=${} \
	-X main.Environment=${} \
	-X main.Hash=${} \
	-X main.Salt=${} \
  go build -buildmode=pie -ldflags "-s -w -linkmode=external ${LDFLAGS}" \
  -o "bin/nordvpn" "cmd/cli"
```
### Building nordvpn daemon
Building requires injecting the following variables via linker:
```sh
# Used to derive keys when encrypting/decrypting configuration files.
# Changing the Salt means changing the keys, which means losing access
# to previously encrypted files. While this is fine during development,
# production builds should never change the Salt between versions.
main.Salt
# Used by `nordvpn --version` command.
# Usually taken from the git commit tag.
main.Version
# Used to implement feature toggles.
# It's either dev or prod.
main.Environment
# Used by deb/rpm repository checker to find out if there is a new version available.
main.Arch
# Used by deb/rpm repository checker to find out if there is a new version available.
main.PackageType
```
#### Conditional compilation
Application features implemented in FFI libraries are hidden behind build tags.
For development builds, it's acceptable to omit build tags, which means that the
application is compiled without CGo dependencies. For production builds, the
following tags are used:
- drop (filesharing feature)
- moose (telemetry feature)
- telio (meshnet feature)
#### Development builds
```sh
LDFLAGS="-X main.Version=${} \
	-X main.Environment=${} \
	-X main.Arch=${} \
	-X main.Salt=${} \
	-X main.PackageType=${} \
  go build -ldflags "${LDFLAGS}" \
	-o "bin/nordvpnd" "cmd/daemon"
```
#### Production (hardened) builds
Since Go 1.17 building position independent executables requires using C linker.
Production builds also use CGo, so the library path has to be given to the C linker.
Due to Rust's unstable ABI, multiple Rust libraries are compiled into a single library
called `libfoss.a` and linked into the application.
C linker also allows adding additional protection known as hardening:
- [Fortification](https://www.redhat.com/en/blog/enhance-application-security-fortifysource)
- [Relocation](https://www.redhat.com/en/blog/hardening-elf-binaries-using-relocation-read-only-relro)
```sh
CGO_CFLAGS="-g -O2 -D_FORTIFY_SOURCE=2" \
LDFLAGS="-X main.Version=${} \
	-X main.Environment=${} \
	-X main.Arch=${} \
	-X main.Salt=${} \
	-X main.PackageType=${} \
  go build -buildmode=pie -tags=drop,moose,telio \
	-ldflags "-s -w -linkmode=external ${LDFLAGS}" \
	-o "bin/nordvpnd" "cmd/daemon"
```
### Building scripts/tooling
Binaries found in `cmd/<binary>/main.go` are not shipped to the users and are built with:
```sh
go build -o "bin/<binary>" "cmd/<binary>"
```
- checkelf (ensures that nordvpn/nordvpnd executables don't exceed glibc version)
- downloader (downloads files from CDN used when building deb/rpm packages)
- pulp (removes outdated packages from deb/rpm package repository)

# Branches
* `main` - main branch which always contains latest changes. Direct commits are strictly forbidden.
# CI/CD
## Releases
Released packages of Linux App can be found in https://repo.nordvpn.com/deb and https://repo.nordvpn.com/yum
## Docker images
### Building images
The following images can be built:
- builder
- generator
- notifier
- packager
- qa-peer
- ruster
- scanner
- snaper
- tester
- uploader

Images are stored in `ghcr.io/nordsecurity/nordvpn-linux` registry.

The building and tagging can be done in a single command like this:
```sh
docker build -t <registry>/<image>[:tag] ci/docker/<image>
```
The pushing can be done with:
```sh
docker push <registry>/<image>[:tag]
```
### Idempotent docker builds
By default, mage targets will always pull docker images from the registry. If below entry is present in `.env` file, images will be pulled only if they are not present on the host system:
```
IDEMPOTENT_DOCKER=1
```
# Linting
We run [golangci-lint](https://github.com/golangci/golangci-lint) for our changes. You can find our linter setting in [.golangci-lint.yml](.golangci-lint.yml).
