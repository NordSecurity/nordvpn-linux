#!/bin/bash
set -euxo pipefail

category="${1}"
pattern="${2}"

export COVERDIR="covdatafiles"

if ! "${WORKDIR}"/ci/install_snap.sh; then
    echo "failed to install snap"
    exit 1
fi

mkdir -p "${WORKDIR}"/dist/logs

cd "${WORKDIR}"/test/qa || exit

args=()

case "${category}" in
    "all")
        ;;
    *)
	args+=("test_${category}.py")
        ;;
esac

case "${pattern}" in
    "")
        ;;
    *)
	args+=("-k ${pattern}")
        ;;
esac


# mkdir -p "${WORKDIR}"/"${COVERDIR}" 

# if ! sudo grep -q "export GOCOVERDIR=${WORKDIR}/${COVERDIR}" "/etc/init.d/nordvpn"; then
#     sudo sed -i "1a export GOCOVERDIR=${WORKDIR}/${COVERDIR}" "/etc/init.d/nordvpn"
# fi

# if [[ -n ${LATTE:-} ]]; then
#     if ! sudo grep -q "export IGNORE_HEADER_VALIDATION=1" "/etc/init.d/nordvpn"; then
#         sudo sed -i "1a export IGNORE_HEADER_VALIDATION=1" "/etc/init.d/nordvpn"
#     fi

#     if ! sudo grep -q "export HTTP_TRANSPORTS=http1" "/etc/init.d/nordvpn"; then
#         sudo sed -i "1a export HTTP_TRANSPORTS=http1" "/etc/init.d/nordvpn"
#     fi
# fi

python3 -m pytest -v --disable-pytest-warnings --timeout 180 -x -rsx --timeout-method=signal -o log_cli=true \
--html=artifacts/report.html --self-contained-html  --junitxml=artifacts/results.xml "${args[@]}"

# if ! sudo grep -q "export GOCOVERDIR=${WORKDIR}/${COVERDIR}" "/etc/init.d/nordvpn"; then
#     sudo sed -i "2d" "/etc/init.d/nordvpn"
# fi

# # To print goroutine profile when debugging:
# RET=$?
# if [ $RET != 0 ]; then
#     curl http://localhost:6960/debug/pprof/goroutine?debug=1
# fi
# exit $RET
