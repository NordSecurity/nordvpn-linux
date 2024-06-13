FROM docker.io/library/golang:1.20.13-bullseye

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN apt-get update && \
    apt-get -y --no-install-recommends install \
        devscripts && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    # Install gosec
    wget --progress=dot:giga https://raw.githubusercontent.com/securego/gosec/master/install.sh -O /tmp/install.sh && \
    echo "2aafc038c170fb6141727d247ad5a17ded5bf195ad0fb8ed1b8e0aed326ac78f /tmp/install.sh" | sha256sum -c - && \
    sh -s v2.13.0 /tmp/install.sh && rm /tmp/install.sh

CMD ["exec", "$@"]
