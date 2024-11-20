import os

import pytest
import sh

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
    logging.log("IP: " + str(network.get_external_device_ip()))
    pytest.skip()


def teardown_module(module):  # noqa: ARG001
    network.unblock()
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log()


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


def test_cmd_not_found_error():
    invalid_cmd = "kinect"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(invalid_cmd)

    print(ex.value)
    assert f"Command '{invalid_cmd}' doesn't exist." in ex.value.stdout.decode()
    assert network.is_disconnected()
