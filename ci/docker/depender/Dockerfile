FROM golang:1.20.1-buster

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN GOBIN=/usr/bin go install github.com/google/go-licenses@v1.6.0 && apt install -y git
