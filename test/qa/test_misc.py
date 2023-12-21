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


# noinspection PyUnusedLocal
def setup_module(module):
    daemon.start()
    login.login_as("default")


# noinspection PyUnusedLocal
def teardown_module(module):
    network.unblock()
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


# noinspection PyUnusedLocal
def setup_function(function):
    logging.log()


# noinspection PyUnusedLocal
def teardown_function(function):
    logging.log()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_api_call_after_vpn_connect():
    # call api
    output = sh.nordvpn.account()
    print(output)
    assert "Account Information:" in output
    # connect vpn
    output = sh.nordvpn.connect()
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
    assert "nordvpn 750" in sh.sudo.stat(socket_dir, "-c", "%G %a")
