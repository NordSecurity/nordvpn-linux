#!/bin/bash
set -euxo pipefail

categories="${1}"
pattern="${2:-}"

export COVERDIR="covdatafiles"

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


mkdir -p "${WORKDIR}"/dist/logs

cd "${WORKDIR}"/test/qa || exit

args=()
read -ra array <<< "$categories"
for category in "${array[@]}"
do 
    case "${category}" in
        "all")
            ;;
        *)
        args+=("test_${category}.py")
            ;;
    esac
done

case "${pattern}" in
    "")
        ;;
    *)
	args+=("-k ${pattern}")
        ;;
esac


mkdir -p "${WORKDIR}"/"${COVERDIR}" 

if ! sudo grep -q "export GOCOVERDIR=${WORKDIR}/${COVERDIR}" "/etc/init.d/nordvpn"; then
    sudo sed -i "1a export GOCOVERDIR=${WORKDIR}/${COVERDIR}" "/etc/init.d/nordvpn"
fi

if [[ -n ${LATTE:-} ]]; then
    if ! sudo grep -q "export IGNORE_HEADER_VALIDATION=1" "/etc/init.d/nordvpn"; then
        sudo sed -i "1a export IGNORE_HEADER_VALIDATION=1" "/etc/init.d/nordvpn"
    fi

    if ! sudo grep -q "export HTTP_TRANSPORTS=http1" "/etc/init.d/nordvpn"; then
        sudo sed -i "1a export HTTP_TRANSPORTS=http1" "/etc/init.d/nordvpn"
    fi
fi

python3 -m pytest -v -x -rsx --setup-timeout 60 --execution-timeout 180 --teardown-timeout 25 -o log_cli=true --html=reports/report.html "${args[@]}"

if ! sudo grep -q "export GOCOVERDIR=${WORKDIR}/${COVERDIR}" "/etc/init.d/nordvpn"; then
    sudo sed -i "2d" "/etc/init.d/nordvpn"
fi

# # To print goroutine profile when debugging:
# RET=$?
# if [ $RET != 0 ]; then
#     curl http://localhost:6960/debug/pprof/goroutine?debug=1
# fi
# exit $RET
