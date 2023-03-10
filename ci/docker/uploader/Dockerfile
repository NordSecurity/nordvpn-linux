FROM amd64/debian:buster-20200607

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN apt-get update --allow-releaseinfo-change && \
    apt upgrade -y && apt-get install -y \
    git python-m2crypto python-pip libssh-dev python-rpm

RUN mkdir -p /build \
    /etc/pulp/admin/conf.d \
    /usr/lib/pulp/admin/extensions \
    /root/.pulp && \
    touch /etc/pulp/admin/admin.conf

ADD setup.sh  /build/setup.sh

RUN cd /build/ && \
    chmod 755 setup.sh && \
    ./setup.sh

RUN /usr/bin/apt-get remove git -y && \
    /usr/bin/apt-get clean && \
    /usr/bin/apt-get -y autoremove && \
    rm -rf /tmp/* /var/tmp/* /build

ADD env.sh /env.sh

RUN chmod 755 /env.sh

ENTRYPOINT ["/env.sh"]
