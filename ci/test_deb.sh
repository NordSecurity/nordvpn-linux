#!/usr/bin/env bash
set -euxo pipefail


# load the env vars needed for tests
source "${WORKDIR}/ci/qa_tests_env.sh"

if ! "${WORKDIR}"/ci/install_deb.sh; then
    echo "failed to install deb"
    exit 1
fi

echo "~~~Diagnose wireguard for possible problems on gitlab runner"
echo "~~~ ip a"
ip a
echo "~~~~~~~~~~~~~~"
echo "lsmod"
lsmod
echo "~~~~~~~~~~~~~~"
echo "sudo ip link add dev wg0diagnose type wireguard"
sudo ip link add dev wg0diagnose type wireguard
echo "~~~~~~~~~~~~~~"
echo "~~~ ip a"
ip a
echo "~~~~~~~~~~~~~~"
echo "sudo ip link del wg0diagnose"
sudo ip link del wg0diagnose
echo "~~~~~~~~~~~~~~"


rm -fr "${GOCOVERDIR}"
mkdir -p "${GOCOVERDIR}"

if ! sudo grep -q "export GOCOVERDIR=${GOCOVERDIR}" "/etc/init.d/nordvpn"; then
    sudo sed -i "1a export GOCOVERDIR=${GOCOVERDIR}" "/etc/init.d/nordvpn"

    revert_go_cov() {
        sudo sed -i "2d" "/etc/init.d/nordvpn"
    }

    trap revert_go_cov EXIT INT TERM
fi

if [[ -n ${NORD_CDN_URL:-} ]]; then
    if ! sudo grep -q "export NORD_CDN_URL=$NORD_CDN_URL" "/etc/init.d/nordvpn"; then
        sudo sed -i "1a export NORD_CDN_URL=$NORD_CDN_URL" "/etc/init.d/nordvpn"
    fi
    if ! sudo grep -q "export IGNORE_HEADER_VALIDATION=1" "/etc/init.d/nordvpn"; then
        sudo sed -i "1a export IGNORE_HEADER_VALIDATION=1" "/etc/init.d/nordvpn"
    fi
fi


"${WORKDIR}/ci/qa_run_tests.sh" "$@"
