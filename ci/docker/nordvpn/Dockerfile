FROM ubuntu:22.04

ARG version
RUN apt-get update && \
    apt-get install -y wget iputils-ping curl && \
    wget -O /tmp/nordrepo.deb https://repo.nordvpn.com/deb/nordvpn/debian/pool/main/nordvpn-release_1.0.0_all.deb && \
    apt-get install -y /tmp/nordrepo.deb && \
    apt-get update && \
    apt-get install -y nordvpn=$version && \
    apt-get remove -y wget nordvpn-release && \
    rm /tmp/nordrepo.deb && \
    apt-get clean

COPY nordvpn/entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
CMD bash