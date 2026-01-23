import pytest
import sh
import os
import glob

from pathlib import Path

import lib
from lib import (
    daemon,
    login,
    logging,
    network,
    IS_NIGHTLY
)
from lib.dynamic_parametrize import dynamic_parametrize

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_module", "collect_logs")

PROJECT_ROOT = os.environ['WORKDIR']
DEB_PATH = glob.glob(f'{PROJECT_ROOT}/dist/app/deb/nordvpn_*amd64.deb')[0]
MSG_KILLSWITCH_ON = "Kill Switch has been successfully set to 'enabled'."
MSG_KILLSWITCH_OFF = "Kill Switch has been successfully set to 'disabled'."


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


@dynamic_parametrize(
    [
        "tech_from", "proto_from", "obfuscated_from",
        "tech_to", "proto_to", "obfuscated_to",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[lib.TECHNOLOGIES],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech_from}-{proto_from}-{obfuscated_from}-"
               "{tech_to}-{proto_to}-{obfuscated_to}",
)
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
        assert "Account information" in output

    sh.nordvpn.set.killswitch("off")
    assert network.is_available()


# This test assumes being run on docker
def test_killswitch_on_after_update():
    assert Path("/.dockerenv").exists(), "Test must be executed in docker"
    # Mocking ps to pretend as if we are in an initd system
    sh.sudo.mv("/usr/bin/ps", "/usr/bin/pso")
    sh.sudo.cp("/etc/mock_ps.sh", "/usr/bin/ps")

    sh.nordvpn.set.killswitch.on()
    assert daemon.is_killswitch_on()
    logging.log(f"Settings before update {sh.nordvpn.settings()}")
    assert network.is_not_available(2)
    sh.sudo.dpkg("-i", DEB_PATH)
    daemon.wait_until_daemon_is_running()
    logging.log(f"Settings after app update {sh.nordvpn.settings()}")
    assert network.is_not_available(2)
    assert daemon.is_killswitch_on()
    sh.nordvpn.set.killswitch.off()
    assert network.is_available()
    # Restore to normal if more tests are run afterwards
    sh.sudo.mv("/usr/bin/pso", "/usr/bin/ps")
