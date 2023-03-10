# QA-Peer Docker image

Image used as a separate peer when doing QA tests (meshnet for example). Designed to be used as a Gitlab service https://docs.gitlab.com/ee/ci/services/

The image doesn't contain nordvpn anymore because it was a burden to update the image with new nordvpn versions. Upon starting the container user should use ssh to upload desired nordvpn deb, install it and login to be able to use it.

## Issues

There is (almost) no way to see logs of service containers in Gitlab https://gitlab.com/gitlab-org/gitlab-runner/-/issues/2119

Redirecting logs to file in repository didn't work for me because repository is being pulled only after the service container is started.

Thankfully, when the service container is started there's always some healthcheck error which prints logs of very beginning of service container output, so that can be used to debug some container issues.

## Usage
`docker run --cap-add=NET_ADMIN -it --rm --name=qa-peer ghcr.io/nordsecurity/nordvpn-linux/qa-peer`
