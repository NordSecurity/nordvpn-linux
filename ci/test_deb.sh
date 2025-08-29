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

if [[ -n ${RC_FILES:-} ]]; then
    if ! sudo grep -q "export RC_USE_LOCAL_CONFIG=1" "/etc/init.d/nordvpn"; then
        sudo sed -i "1a export RC_USE_LOCAL_CONFIG=1" "/etc/init.d/nordvpn"
    fi
    sudo mkdir /var/lib/nordvpn/conf
    # Config directory is only created on usage, we need to put files on install
    sudo cp -r $RC_FILES '/var/lib/nordvpn/conf/'
fi


"${WORKDIR}/ci/qa_run_tests.sh" "$@"
