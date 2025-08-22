#!/usr/bin/env bash
set -euxo pipefail

# Run the QA tests with pytest. 
# Parameters:
# 1. test_categories - one or more categories for which to run the tests. Possible values
#     - "all" - for running all the tests
#     - "category_1 category_2" - the python file names minus test_, e.g. test_category1.py test_category2.py
# 2. pattern - what tests to run from the given categories
#     - "test" - for running all the tests in that categories
#     - "test_function_name" - to run a specific test function from the given categories, e.g.: test_check_routing_table_for_lan

if [[ $# -ne 2 ]]; then
    echo "Usage: $0 \"<test_categories>\" \"<pattern>\""
    exit 1
fi

categories="${1}"
pattern="${2:-}"

# check that some env variables are set before running the tests
: "${DISABLE_TUI_LOADER:?DISABLE_TUI_LOADER must be set to disable ANSI loading indicator in CLI commands}"

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


ARTIFACTS_FOLDER="${WORKDIR}"/dist/test_artifacts
LOGS_FOLDER="${WORKDIR}"/dist/logs

mkdir -p "${LOGS_FOLDER}"
mkdir -p "${ARTIFACTS_FOLDER}"
mkdir -p "${GOCOVERDIR}"

if [[ -n ${NORD_CDN_URL:-} ]]; then
    if ! sudo grep -q "export NORD_CDN_URL=$NORD_CDN_URL" "/etc/init.d/nordvpn"; then
        sudo sed -i "1a export NORD_CDN_URL=$NORD_CDN_URL" "/etc/init.d/nordvpn"
    fi
    if ! sudo grep -q "export IGNORE_HEADER_VALIDATION=1" "/etc/init.d/nordvpn"; then
        sudo sed -i "1a export IGNORE_HEADER_VALIDATION=1" "/etc/init.d/nordvpn"
    fi
fi

python3 -m pytest -v -x -rsx --setup-timeout 60 --execution-timeout 180 --teardown-timeout 25 -o log_cli=true \
--html="${WORKDIR}"/dist/test_artifacts/report.html --self-contained-html  --junitxml="${WORKDIR}"/dist/test_artifacts/report.xml "${args[@]}"

# # To print goroutine profile when debugging:
# RET=$?
# if [ $RET != 0 ]; then
#     curl http://localhost:6960/debug/pprof/goroutine?debug=1
# fi
# exit $RET
