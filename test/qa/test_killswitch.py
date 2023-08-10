from lib import (
    daemon,
    info,
    logging,
    login,
    network,
)
import lib
import pytest
import sh
import timeout_decorator


def setup_module(module):
    daemon.start()
    login.login_as("default")


def teardown_module(module):
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


# @TODO optimize takes > 1hour. using 1 technology for now
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES_BASIC1)
def test_killswitch_without_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    output = sh.nordvpn.set.killswitch("on")
    assert "Kill Switch is set to 'enabled' successfully." in output

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    sh.nordvpn.set.killswitch("off")
    assert network.is_available()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES_BASIC1)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_killswitch_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    output = sh.nordvpn.set.killswitch("on")
    assert "Kill Switch is set to 'enabled' successfully." in output

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

        with lib.ErrorDefer(sh.nordvpn.disconnect):
            output = sh.nordvpn.connect()
            print(output)
            assert lib.is_connect_successful(output)
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    sh.nordvpn.set.killswitch("off")
    assert network.is_available()


@pytest.mark.parametrize(
    "tech_from,proto_from,obfuscated_from", lib.TECHNOLOGIES_BASIC1
)
@pytest.mark.parametrize("tech_to,proto_to,obfuscated_to", lib.TECHNOLOGIES_BASIC2)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_killswitch_reconnect(
    tech_from, proto_from, obfuscated_from, tech_to, proto_to, obfuscated_to
):
    lib.set_technology_and_protocol(tech_from, proto_from, obfuscated_from)
    assert network.is_available()

    output = sh.nordvpn.set.killswitch("on")
    assert "Kill Switch is set to 'enabled' successfully." in output

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

        with lib.ErrorDefer(sh.nordvpn.disconnect):
            output = sh.nordvpn.connect()
            print(output)
            assert lib.is_connect_successful(output)
            assert network.is_connected()

            lib.set_technology_and_protocol(tech_to, proto_to, obfuscated_to)
            assert network.is_connected()
            output = sh.nordvpn.connect()
            assert lib.is_connect_successful(output)
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert daemon.is_disconnected()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    sh.nordvpn.set.killswitch("off")
    assert network.is_available()


# Test for 3.8.7 hotfix. Account and login commands would not work when killswitch is on
# Issue 441
def test_fancy_transport():
    sh.nordvpn.logout("--persist-token")
    output = sh.nordvpn.set.killswitch("on")
    assert "Kill Switch is set to 'enabled' successfully." in output

    output = login.login_as("default")
    print(output)
    assert "Welcome to NordVPN!" in output

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        output = sh.nordvpn.account()
        print(output)
        assert "Account Information:" in output


    sh.nordvpn.set.killswitch("off")
    assert network.is_available()