import os
import re
import zipfile

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


# ---------------------------------------------------------------------------
# `nordvpn troubleshoot`
# ---------------------------------------------------------------------------

EXPECTED_ZIP_ENTRIES = {
    "daemon.log",
    "system-info.txt",
    "network-info.txt",
    "dns-info.txt",
    "nftables-ruleset.txt",
    "log_extraction_report.log",
}


def _run_troubleshoot() -> str:
    """
    Run `nordvpn troubleshoot`, assert the success banner, and return the
    zip path the daemon reports.
    """
    output = str(sh.nordvpn.troubleshoot())
    assert "Diagnostics collected successfully." in output, output
    match = re.search(r"File saved to:\s*(\S+)", output)
    assert match, f"Could not find zip path in output:\n{output}"
    return match.group(1)


def _cleanup_zip(path: str | None) -> None:
    if not path:
        return
    try:
        os.remove(path)
    except FileNotFoundError:
        pass


def test_troubleshoot():
    """
    Single end-to-end check for `nordvpn troubleshoot`:
      - the file exists, ends in .zip, has the diagnostics naming scheme
      - the file is owned by the invoking user and readable
      - the zip is CRC-valid and contains every expected entry
      - log_extraction_report.log records the per-step start/finish lines
      - system-info.txt and dns-info.txt have their labelled blocks
      - a second invocation produces a different zip path (no overwrite)

    The success banner check is performed inside `_run_troubleshoot`.
    """
    zip_path = _run_troubleshoot()
    second_zip_path = None
    try:
        # File on disk
        assert os.path.isfile(zip_path), f"zip not found at {zip_path}"
        assert zip_path.endswith(".zip")
        assert "nordvpn-diagnostics-" in os.path.basename(zip_path)

        # File permissions: owned by the invoking user and readable by them.
        st = os.stat(zip_path)
        assert st.st_uid == os.getuid(), f"zip owner uid {st.st_uid} != caller uid {os.getuid()}"
        assert st.st_gid == os.getgid(), f"zip owner gid {st.st_gid} != caller gid {os.getgid()}"
        assert st.st_mode & 0o400, f"zip not readable by owner (mode={oct(st.st_mode)})"

        with zipfile.ZipFile(zip_path) as zf:
            assert zf.testzip() is None, "zip has CRC errors"
            names = set(zf.namelist())
            report = zf.read("log_extraction_report.log").decode("utf-8")
            sysinfo = zf.read("system-info.txt").decode("utf-8")
            dnsinfo = zf.read("dns-info.txt").decode("utf-8")

        # Entry list
        missing = EXPECTED_ZIP_ENTRIES - names
        assert not missing, f"missing entries: {missing}"

        # log_extraction_report.log
        assert "diagnostics collection started" in report
        assert "diagnostics collection finished" in report
        started = report.count("step started:")
        completed = report.count("step completed:")
        assert started >= 7, f"expected >=7 step-start entries, got {started}"
        # Some steps may be non-fatal failures on a minimal CI box, so
        # "completed" can be < started; just assert the logger wrote *something*.
        assert completed >= 1

        # system-info blocks
        for block in ("OS Release", "Kernel Version", "Desktop Environment", "Systemd Status"):
            assert f"=== {block} ===" in sysinfo, f"missing system-info block {block!r}"

        # dns-info blocks
        for block in (
            "/etc/resolv.conf",
            "NetworkManager DNS Mode (dbus)",
            "NetworkManager DNS Configuration (dbus)",
        ):
            assert f"=== {block} ===" in dnsinfo, f"missing dns-info block {block!r}"

        # Second invocation must not overwrite the first — names embed a
        # timestamp, so two consecutive runs should produce distinct paths.
        second_zip_path = _run_troubleshoot()
        assert second_zip_path != zip_path, (
            f"second invocation reused first path: {second_zip_path}"
        )
        assert os.path.isfile(second_zip_path)
        assert os.path.isfile(zip_path), "first zip was overwritten/removed"
    finally:
        _cleanup_zip(zip_path)
        _cleanup_zip(second_zip_path)
