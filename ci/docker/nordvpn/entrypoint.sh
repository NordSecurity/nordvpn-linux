#!/bin/bash
set -eu

# To prevent leaking IP prior to nordvpn start
if ! iptables -I INPUT -j DROP || ! iptables -I OUTPUT -j DROP; then 
    echo "Make sure to pass '--cap-add=NET_ADMIN' to 'docker run'"
    exit 1
fi

/etc/init.d/nordvpn start

# Wait until nordvpn actually starts
set +e
for i in {1..5} 
do
    sleep 1
    nordvpn status > /dev/null 2>&1 && break
    if [[ $i == 5 ]] 
    then
        cat /var/log/nordvpn/daemon.log
        echo "Cannot start NordVPN daemon"
        exit 1
    fi
done
set -e

nordvpn set killswitch on
iptables -D INPUT -j DROP
iptables -D OUTPUT -j DROP

nordvpn login --token "$NORDVPN_LOGIN_TOKEN"

/bin/bash -c "$@"

tail -fn +1 --pid "$(pidof nordvpnd)" /var/log/nordvpn/daemon.log
