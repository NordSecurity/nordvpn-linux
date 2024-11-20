import time

import sh

from lib import daemon, login


def setup_module(module):  # noqa: ARG001
    sh.sudo.apt.purge("-y", "nordvpn")

    sh.sh(_in=sh.curl("-sSf", "https://downloads.nordcdn.com/apps/linux/install.sh"))

    daemon.start()
    while not daemon.is_running():
        time.sleep(1)

    login.login_as("default")
