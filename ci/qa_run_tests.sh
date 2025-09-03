#!/usr/bin/env bash
set -euxo pipefail

# Run the QA tests with pytest. 
# Parameters:
# 1. test_categories - one or more categories for which to run the tests. Possible values
#     - "all" - for running all the tests
#     - "category_1 category_2" - the python file names minus test_, e.g. test_category1.py test_category2.py
# 2. pattern[optional] - what tests to run from the given categories
#     - when missing is running all the tests in that categories
#     - "test" - for running all the tests in that categories from mage command
#     - "test_function_name" - to run a specific test function from the given categories, e.g.: test_check_routing_table_for_lan

if [[ $# -gt 2 ]]; then
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


# check that the nordvpn group exists in the system and that the current user is part of it
GROUP="nordvpn"
if ! getent group "$GROUP" > /dev/null; then
	# application installer must create the group
	echo "Group '$GROUP' does not exist."
	exit 1
fi

# if user variable is not set, set it with whoami result
: "${USER:=$(whoami)}"

# add current user into nordvpn group if needed
if id -nG "$USER" | grep -qw "$GROUP"; then
	echo "User '$USER' is part of the in group '$GROUP'."
else
	echo "Adding user '$USER'to group '$GROUP'..."
	sudo usermod -aG "$GROUP" "$USER"
fi

ARTIFACTS_FOLDER="${WORKDIR}"/dist/test_artifacts

mkdir -p "${LOGS_FOLDER}"
mkdir -p "${ARTIFACTS_FOLDER}"
mkdir -p "${GOCOVERDIR}"

python3 -m pytest -v -x -rsx --setup-timeout 60 --execution-timeout 180 --teardown-timeout 25 -o log_cli=true \
--html="${WORKDIR}"/dist/test_artifacts/report.html --self-contained-html  --junitxml="${WORKDIR}"/dist/test_artifacts/report.xml "${args[@]}"

# # To print goroutine profile when debugging:
# RET=$?
# if [ $RET != 0 ]; then
#     curl http://localhost:6960/debug/pprof/goroutine?debug=1
# fi
# exit $RET
