import pytest
import sh
import os

import lib
import time
from lib import (
    daemon,
    login,
    network,
)


pytestmark = pytest.mark.usefixtures("nordvpnd_scope_module", "collect_logs")


MSG_KILLSWITCH_ON = "Kill Switch is set to 'enabled' successfully."
MSG_KILLSWITCH_OFF = "Kill Switch is set to 'disabled' successfully."


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_killswitch_on_disconnected(tech, proto, obfuscated):
    """Manual TC: LVPN-419"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
    assert daemon.is_killswitch_on()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    assert MSG_KILLSWITCH_OFF in sh.nordvpn.set.killswitch("off")
    assert not daemon.is_killswitch_on()
    assert network.is_available()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_killswitch_on_connect(tech, proto, obfuscated):
    """Manual TC: LVPN-8707"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
    assert daemon.is_killswitch_on()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

        with lib.ErrorDefer(sh.nordvpn.disconnect):
            output = sh.nordvpn.connect()
            print(output)
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    assert MSG_KILLSWITCH_OFF in sh.nordvpn.set.killswitch("off")
    assert not daemon.is_killswitch_on()
    assert network.is_available()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_killswitch_on_connected(tech, proto, obfuscated):
    """Manual TC: LVPN-1394"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    with lib.Defer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect()
        print(output)
        assert network.is_connected()

        assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
        assert daemon.is_killswitch_on()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    assert MSG_KILLSWITCH_OFF in sh.nordvpn.set.killswitch("off")
    assert not daemon.is_killswitch_on()
    assert network.is_available()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_killswitch_off_connected(tech, proto, obfuscated):
    """Manual TC: LVPN-2195"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
    assert daemon.is_killswitch_on()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

        with lib.Defer(sh.nordvpn.disconnect):
            output = sh.nordvpn.connect()
            print(output)
            assert network.is_connected()

            assert MSG_KILLSWITCH_OFF in sh.nordvpn.set.killswitch("off")
            assert not daemon.is_killswitch_on()
            assert network.is_available()

    assert network.is_available()


@pytest.mark.parametrize(("tech_from", "proto_from", "obfuscated_from"), lib.TECHNOLOGIES)
@pytest.mark.parametrize(("tech_to", "proto_to", "obfuscated_to"), lib.TECHNOLOGIES)
def test_killswitch_reconnect(tech_from, proto_from, obfuscated_from, tech_to, proto_to, obfuscated_to):
    """Manual TC: LVPN-8716"""

    lib.set_technology_and_protocol(tech_from, proto_from, obfuscated_from)
    assert network.is_available()

    assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
    assert daemon.is_killswitch_on()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

        with lib.ErrorDefer(sh.nordvpn.disconnect):
            output = sh.nordvpn.connect()
            print(output)
            assert network.is_connected()

            lib.set_technology_and_protocol(tech_to, proto_to, obfuscated_to)
            assert network.is_connected()
            output = sh.nordvpn.connect()
            print(output)
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert daemon.is_disconnected()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    assert MSG_KILLSWITCH_OFF in sh.nordvpn.set.killswitch("off")
    assert not daemon.is_killswitch_on()
    assert network.is_available()


# Test for 3.8.7 hotfix. Account and login commands would not work when killswitch is on
# Issue 441
def test_fancy_transport():
    """Manual TC: LVPN-8717"""

    sh.nordvpn.logout("--persist-token")
    output = sh.nordvpn.set.killswitch("on")
    assert MSG_KILLSWITCH_ON in output

    output = login.login_as("default")
    print(output)
    assert "Welcome to NordVPN!" in output

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        output = sh.nordvpn.account()
        print(output)
        assert "Account Information:" in output

    sh.nordvpn.set.killswitch("off")
    assert network.is_available()


def test_killswitch_enabled_does_not_affect_cdn_with_firewall_mark():
    """
    Test for a scenario where despite killswitch being enabled the network is still accessible for remote config.

    Test steps:
        1. Set up required environment variables
        2. Enable killswitch
        3. Remove previously fetched config files
        4. Attempt to fetch remote config
        5. Verify that the config is fetched successfully

    Jira ID: LVPN-8626
    """

    conf_dir = "/var/lib/nordvpn/conf/"
    expected_files = [
        "libtelio.json",
        "meshnet.json",
    ]

    # enabling killswitch should not affect http transport of the remote config
    assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
    assert daemon.is_killswitch_on()

    sh.nordvpn.disconnect()
    assert daemon.is_disconnected()
    # remove previously fetched files
    # upon restart, they should be loaded again
    os.system(f"sudo rm -rf {conf_dir}")

    # make sure the rc files are gone
    for fname in expected_files:
        path = os.path.join(conf_dir, fname)
        res = os.popen(f"sudo test -f {path} && echo exists || echo missing").read().strip()
        assert res == "missing", f"File {path} should not exist"

    daemon.restart()
    time.sleep(3)

    for fname in expected_files:
        path = os.path.join(conf_dir, fname)
        res = os.popen(f"sudo test -f {path} && echo exists || echo missing").read().strip()
        assert res == "exists", f"File {os.path} should exist after kill-switch was enabled"


