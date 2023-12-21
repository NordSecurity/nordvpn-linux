"""Logging utilities to make it easier to write to daemon logs."""
import os

import sh

# log file location
FILE = f"{os.environ['WORKDIR']}/dist/logs/daemon.log"


def log(data=None):
    """log test name to the daemon logs or data if provided, but not both"""
    # Printing this way prints the pure data into a file, going the bash -c echo route
    # is vulnerable to double quotes character being found and subsequent lines being taken
    # as pure bash code (and failing as it begins to list processes taking them as commands)
    if data:
        sh.sudo.bash("-c", f"""cat <<EOF >> {FILE}
{data}
EOF
""")
    else:
        test_name = os.environ["PYTEST_CURRENT_TEST"]
        sh.sudo.bash("-c", f"echo '{test_name}' >> {FILE}")
        print(test_name)
