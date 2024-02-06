import pytest
import sh
import timeout_decorator

import lib
from lib import daemon, info, logging, login, network, server


def setup_module(module):  # noqa: ARG001
    daemon.start()
    login.login_as("default")


def teardown_module(module):  # noqa: ARG001
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()


def autoconnect_base_test(group):
    output = sh.nordvpn.set.autoconnect.on(group)
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_autoconnect_default(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test("")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_not_autoconnect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.set.autoconnect.off()
    print(output)

    daemon.restart()
    assert network.is_disconnected()


@pytest.mark.parametrize("group", lib.COUNTRIES + lib.COUNTRY_CODES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_autoconnect_to_country(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.CITIES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_autoconnect_to_city(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_autoconnect_to_random_server_by_name(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    _, hostname = server.get_hostname_by(tech, proto, obfuscated)
    name = hostname.split(".")[0]

    autoconnect_base_test(name)


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_autoconnect_to_standard_group(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_autoconnect_to_additional_group(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)
