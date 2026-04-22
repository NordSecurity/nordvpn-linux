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
    """Manual TC: LVPN-8775"""

    # call api
    output = sh.nordvpn.account()
    print(output)
    assert "Account information" in output, "Account command should display account information"
    # connect vpn
    output = sh.nordvpn.connect(_tty_out=False)
    print(output)
    assert lib.is_connect_successful(output), "VPN connection should be successful"
    assert network.is_connected(), "Network should be connected"
    # call api again
    output = sh.nordvpn.account()
    print(output)
    assert "Account information" in output, "Account command should display account information"
    # disconnect vpn
    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output), "VPN disconnection should be successful"
    assert network.is_disconnected(), "Network should be disconnected"


def test_daemon_socket_permissions():
    """Manual TC: LVPN-8774"""

    socket_dir = "/run/nordvpn"
    check_info = "nordvpn 750"
    if daemon.is_under_snap():
        socket_dir = "/var/snap/nordvpn/common/run/nordvpn"
        check_info = "root 755"
    cmd_str = f"sudo stat -c '%G %a' {socket_dir}"
    out = os.popen(cmd_str).read()
    assert check_info in out, f"Socket directory should have permissions {check_info}"


def test_cmd_not_found_error():
    """Manual TC: LVPN-8773"""

    invalid_cmd = "kinect"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(invalid_cmd)

    print(ex.value)
    assert f"Command '{invalid_cmd}' doesn't exist." in ex.value.stdout.decode(), "Invalid command should show error message"
    assert network.is_disconnected(), "Network should be disconnected"


def test_account_output_contains_all_fields():
    """
    Verify that account command output contains all expected fields.

    This test ensures that the account command displays all required fields
    without checking their specific values.
    """
    output = sh.nordvpn.account()

    # Header
    assert "Account information" in output, "Account command should display account information"

    # Required fields (always present)
    assert "Email address:" in output, "Account output should contain email address field"
    assert "Account created:" in output, "Account output should contain account created field"
    assert "Subscription:" in output, "Account output should contain subscription field"
    assert "Dedicated IP:" in output, "Account output should contain dedicated IP field"
    assert "Multi-factor authentication (MFA):" in output, "Account output should contain MFA field"
    assert "Terms of Service" in output, "Account output should contain Terms of Service"
    assert "Auto-renewal terms" in output, "Account output should contain auto-renewal terms"
    assert "Privacy Policy" in output, "Account output should contain Privacy Policy"


def test_account_not_logged_in():
    """Verify account command shows error when not logged in."""
    sh.nordvpn.logout("--persist-token", _ok_code=[0, 1])

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.account()

    assert "You're not logged in." in ex.value.stdout.decode("utf-8"), "Account command should show error when not logged in"
