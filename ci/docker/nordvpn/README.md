# NordVPN Docker image

## Usage

Default behavior is to open bash session to configure nordvpn. When bash session is closed container keeps running and printing daemon logs to the screen.

Killswitch is enabled by default so that the image could be safely used to route other containers through VPN without leaking IP.

### Building
`docker build -f ci/docker/nordvpn/Dockerfile -t nordvpn --build-arg version=x.x.x ci/docker`
* `version` - desired NordVPN version available in APT (example: `version=3.14.1`)

#### Using dev nordvpn build
`docker build -f ci/docker/nordvpn/Dockerfile.dev -t nordvpn .`
* Must have .deb built in dist/app/deb (`mage build:debDocker`)
* Be aware that root of project must be passed as build context to be able to copy .deb into image

### Running
`docker run -e NORDVPN_LOGIN_TOKEN=0123456789abcdef --cap-add=NET_ADMIN -it --rm --name=nordvpn nordvpn`

#### Running custom command
When custom command finishes container keeps running and printing daemon logs to the screen.

`docker run -e NORDVPN_LOGIN_TOKEN=0123456789abcdef --cap-add=NET_ADMIN -it --rm --name=vpn vpn "nordvpn set killswitch off && nordvpn set meshnet on && nordvpn connect"`

### Routing from other container through VPN
`docker run --net=container:nordvpn -it --rm ubuntu`

### Example docker-compose.yaml
```
version: "3"
services:
  nordvpn:
    image: nordvpn
    cap_add:
      - NET_ADMIN
    environment:
      - NORDVPN_LOGIN_TOKEN=0123456789abcdef
    command: nordvpn set technology openvpn && nordvpn connect
  ubuntu:
    image: ubuntu
    network_mode: service:nordvpn
    depends_on:
      - nordvpn
```