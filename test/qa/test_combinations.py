import pytest
import sh
import timeout_decorator

import lib
from lib import (
    daemon,
    info,
    logging,
    login,
    network,
)
from test_connect import connect_base_test, disconnect_base_test


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


@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.TECHNOLOGIES)
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

    output = sh.nordvpn.connect(_tty_out=False)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    output = sh.nordvpn.connect(_tty_out=False)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.TECHNOLOGIES)
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

    output = sh.nordvpn.connect(_tty_out=False)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    output = sh.nordvpn.connect(_tty_out=False)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


@pytest.mark.parametrize(("country", "city"), list(zip(lib.COUNTRIES, lib.CITIES, strict=False)))
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_country_and_city(country, city):
    lib.set_technology_and_protocol("nordlynx", "", "")

    output = sh.nordvpn.connect(country, _tty_out=False)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.connect(city, _tty_out=False)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.connect(country, city, _tty_out=False)
    print(output)
    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_status_change_technology_and_protocol(
        source_tech,
        target_tech,
        source_proto,
        target_proto,
        source_obfuscated,
        target_obfuscated,
):
    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        assert source_tech.upper() in sh.nordvpn.status()

        if source_tech == "openvpn":
            assert source_proto.upper() in sh.nordvpn.status()
        else:
            assert "UDP" in sh.nordvpn.status()

        lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)
        assert source_tech.upper() in sh.nordvpn.status()

        if source_tech == "openvpn":
            assert source_tech.upper() in sh.nordvpn.status()
        else:
            assert "UDP" in sh.nordvpn.status()

    assert network.is_disconnected()


@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_status_change_technology_and_protocol_reconnect(
        source_tech,
        target_tech,
        source_proto,
        target_proto,
        source_obfuscated,
        target_obfuscated,
):
    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        assert target_tech.upper() in sh.nordvpn.status()

        if target_tech == "openvpn":
            assert target_proto.upper() in sh.nordvpn.status()
        else:
            assert "UDP" in sh.nordvpn.status()

    assert network.is_disconnected()


@pytest.mark.parametrize("source_group", lib.STANDARD_GROUPS[-2:])
@pytest.mark.parametrize("target_group", lib.STANDARD_GROUPS[-2:])
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@timeout_decorator.timeout(40)
def test_reconnect_to_standard_group(
    source_tech,
    target_tech,
    source_proto,
    target_proto,
    source_obfuscated,
    target_obfuscated,
    source_group,
    target_group,
):

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    connect_base_test((source_tech, source_proto, source_obfuscated), source_group)

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    connect_base_test((target_tech, target_proto, target_obfuscated), target_group)

    disconnect_base_test()


@pytest.mark.parametrize("source_group", lib.ADDITIONAL_GROUPS[-2:])
@pytest.mark.parametrize("target_group", lib.ADDITIONAL_GROUPS[-2:])
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@timeout_decorator.timeout(40)
def test_reconnect_to_additional_group(
    source_tech,
    target_tech,
    source_proto,
    target_proto,
    source_obfuscated,
    target_obfuscated,
    source_group,
    target_group,
):

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    connect_base_test((source_tech, source_proto, source_obfuscated), source_group)

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    connect_base_test((target_tech, target_proto, target_obfuscated), target_group)

    disconnect_base_test()


@pytest.mark.parametrize("source_country", lib.COUNTRIES[-2:])
@pytest.mark.parametrize("target_country", lib.COUNTRIES[-2:])
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@timeout_decorator.timeout(40)
def test_reconnect_to_server_by_country_name(
    source_tech,
    target_tech,
    source_proto,
    target_proto,
    source_obfuscated,
    target_obfuscated,
    source_country,
    target_country,
):

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    connect_base_test((source_tech, source_proto, source_obfuscated), source_country)

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    connect_base_test((target_tech, target_proto, target_obfuscated), target_country)

    disconnect_base_test()
