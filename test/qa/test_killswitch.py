import pytest
import sh

import lib
from lib import (
    daemon,
    info,
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
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()


MSG_KILLSWITCH_ON = "Kill Switch is set to 'enabled' successfully."
MSG_KILLSWITCH_OFF = "Kill Switch is set to 'disabled' successfully."


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_killswitch_on_disconnected(tech, proto, obfuscated):
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
