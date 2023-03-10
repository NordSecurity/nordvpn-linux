"""Logging utilities to make it easier to write to daemon logs. """
import os
import sh

# log file location
FILE = f"{os.environ['CI_PROJECT_DIR']}/dist/logs/daemon.log"


def log(data=None):
    """log test name to the daemon logs or data if provided, but not both"""
    if data:
        sh.sudo.bash("-c", f"echo \"{data}\" >> {FILE}")
    else:
        test_name = os.environ["PYTEST_CURRENT_TEST"]
        sh.sudo.bash("-c", f"echo '{test_name}' >> {FILE}")
        print(test_name)
