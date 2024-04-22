import os

import pytest
import sh
import timeout_decorator

import lib
from lib import (
    daemon,
    logging,
    login,
    network,
)


def setup_module(module):  # noqa: ARG001
    daemon.start()
    login.login_as("default")


def teardown_module(module):  # noqa: ARG001
    network.unblock()
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_api_call_after_vpn_connect():
    # call api
    output = sh.nordvpn.account()
    print(output)
    assert "Account Information:" in output
    # connect vpn
    output = sh.nordvpn.connect(_tty_out=False)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()
    # call api again
    output = sh.nordvpn.account()
    print(output)
    assert "Account Information:" in output
    # disconnect vpn
    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


def test_daemon_socket_permissions():
    socket_dir = "/run/nordvpn"
    check_info = "nordvpn 750"
    if daemon.is_under_snap():
        socket_dir = "/var/snap/nordvpn/common/run/nordvpn"
        check_info = "root 755"
    cmd_str = f"sudo stat -c '%G %a' {socket_dir}"
    out = os.popen(cmd_str).read()
    assert check_info in out
