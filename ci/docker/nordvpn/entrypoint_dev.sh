#!/bin/bash
set -eu

/etc/init.d/nordvpn start

# Wait until nordvpn actually starts
set +e
for i in {1..10} 
do
    sleep 1
    nordvpn status > /dev/null 2>&1 && break
    if [[ $i == 10 ]] 
    then
        cat /var/log/nordvpn/daemon.log
        echo "Cannot start NordVPN daemon"
        exit 1
    fi
done
set -e

nordvpn login --token "$NORDVPN_LOGIN_TOKEN"

/bin/bash -c "$@"
