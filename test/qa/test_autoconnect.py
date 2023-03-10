from lib import daemon, info, logging, login, network, server
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


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(20)
def test_autoconnect_default(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.set.autoconnect.on()
    print(output)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        daemon.restart()
        daemon.wait_for_autoconnect()
        with lib.ErrorDefer(sh.nordvpn.set.autoconnect.off):
            assert network.is_connected()

    output = sh.nordvpn.set.autoconnect.off()
    print(output)

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(20)
def test_not_autoconnect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.set.autoconnect.off()
    print(output)

    daemon.restart()
    assert network.is_disconnected()


@pytest.mark.parametrize("country", lib.COUNTRIES)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(20)
def test_autoconnect_to_country(tech, proto, obfuscated, country):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.set.autoconnect.on(country)
    print(output)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        daemon.restart()
        daemon.wait_for_autoconnect()
        with lib.ErrorDefer(sh.nordvpn.set.autoconnect.off):
            assert network.is_connected()

    output = sh.nordvpn.set.autoconnect.off()
    print(output)

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


@pytest.mark.parametrize("city", lib.CITIES)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(20)
def test_autoconnect_to_city(tech, proto, obfuscated, city):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.set.autoconnect.on(city)
    print(output)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        daemon.restart()
        daemon.wait_for_autoconnect()
        with lib.ErrorDefer(sh.nordvpn.set.autoconnect.off):
            assert network.is_connected()

    output = sh.nordvpn.set.autoconnect.off()
    print(output)

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(20)
def test_autoconnect_to_random_server_by_name(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    _, hostname = server.get_hostname_by(tech, proto, obfuscated)
    name = hostname.split(".")[0]

    output = sh.nordvpn.set.autoconnect.on(name)
    print(output)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        daemon.restart()
        daemon.wait_for_autoconnect()
        with lib.ErrorDefer(sh.nordvpn.set.autoconnect.off):
            assert network.is_connected()

    output = sh.nordvpn.set.autoconnect.off()
    print(output)

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()
