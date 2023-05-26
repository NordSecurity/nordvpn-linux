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


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_standard_group(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.connect(group)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_additional_group(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.connect(group)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


@pytest.mark.parametrize("target_tech,target_proto,target_obfuscated", lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize("source_tech,source_proto,source_obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_reconnect_matrix_standard(
    source_tech,
    target_tech,
    source_proto,
    target_proto,
    source_obfuscated,
    target_obfuscated,
):
    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    output = sh.nordvpn.connect()
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    output = sh.nordvpn.connect()
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

@pytest.mark.parametrize("target_tech,target_proto,target_obfuscated", lib.OBFUSCATED_TECHNOLOGIES)
@pytest.mark.parametrize("source_tech,source_proto,source_obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_reconnect_matrix_obfuscated(
    source_tech,
    target_tech,
    source_proto,
    target_proto,
    source_obfuscated,
    target_obfuscated,
):
    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    output = sh.nordvpn.connect()
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    output = sh.nordvpn.connect()
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()




@pytest.mark.parametrize("country, city", list(zip(lib.COUNTRIES, lib.CITIES)))
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_country_and_city(country, city):
    lib.set_technology_and_protocol("nordlynx", "", "")

    output = sh.nordvpn.connect(country)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.connect(city)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.connect(country, city)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()
