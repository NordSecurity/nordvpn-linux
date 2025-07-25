import os

import pytest
import sh
import json

import lib
from lib import (
    daemon,
    network,
)

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_module", "unblock_network", "collect_logs")


def test_static_install_file_creation():
    static_config_dir = "/var/lib/nordvpn/data/"
    if daemon.is_under_snap():
        static_config_dir = "/var/snap/nordvpn/common/var/lib/nordvpn/data/"
    static_config_path = f"{static_config_dir}/install_static.dat"

    rollout_group_field_name = "rollout_group"

    rollout_group = 0
    # use popen because config file requires need sudo privileges
    static_config_json = os.popen(f"sudo cat {static_config_path}")
    static_config = json.load(static_config_json)
    rollout_group = static_config[rollout_group_field_name]

    assert rollout_group != 0, "Rollout group was not configured on startup."

    os.popen(f"rm {static_config_path}")
    daemon.restart()

    static_config_json = os.popen(f"sudo cat {static_config_path}")
    static_config = json.load(static_config_json)
    rollout_group_after_restart = static_config[rollout_group_field_name]

    assert rollout_group == rollout_group_after_restart, "Different rollout group was generated after restarting the deamon."


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
