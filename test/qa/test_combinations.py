import pytest
import sh

import lib
from lib import (
    daemon,
    network
)
from test_connect import disconnect_base_test, get_alias

def connect_base_test(group: str = (), name: str = "", hostname: str = ""):
    """
    Connects to a NordVPN server and performs a series of checks to ensure the connection is successful.

    Parameters
    ----------
    group (str): The specific server name or group name to connect to. Default is an empty string.
    name (str): Used to verify the connection message. Default is an empty string.
    hostname (str): Used to verify the connection message. Default is an empty string.
    """

    output = sh.nordvpn.connect(group, _tty_out=False)
    print(output)

    assert lib.is_connect_successful(output, name, hostname)
    assert network.is_connected()


pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")


@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.TECHNOLOGIES)
def test_reconnect_matrix(
        source_tech,
        target_tech,
        source_proto,
        target_proto,
        source_obfuscated,
        target_obfuscated,
):
    """Manual TC: LVPN-8674"""

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)
    connect_base_test()

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)
    connect_base_test()

    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize(("country", "city"), list(zip(lib.COUNTRIES, lib.CITIES, strict=False)))
def test_connect_country_and_city(tech, proto, obfuscated, country, city):
    """Manual TC: LVPN-8610"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(country)
    connect_base_test(city)
    connect_base_test(f"{country} {city}")

    disconnect_base_test()


@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_status_change_technology_and_protocol(
        source_tech,
        target_tech,
        source_proto,
        target_proto,
        source_obfuscated,
        target_obfuscated,
):
    """Manual TC: LVPN-666"""

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    sh.nordvpn(get_alias())
    status_info = daemon.get_status_data()

    assert source_tech.upper() in status_info["current technology"]

    if source_tech == "openvpn":
        assert source_proto.upper() in status_info["current protocol"]
    elif source_tech == "nordwhisper":
        assert "Webtunnel" in status_info["current protocol"]
    else:
        assert "UDP" in status_info["current protocol"]

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)
    assert source_tech.upper() in status_info["current technology"]

    if source_tech == "openvpn":
        assert source_proto.upper() in status_info["current protocol"]
    elif source_tech == "nordwhisper":
        assert "Webtunnel" in status_info["current protocol"]
    else:
        assert "UDP" in status_info["current protocol"]

    disconnect_base_test()


@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_status_change_technology_and_protocol_reconnect(
        source_tech,
        target_tech,
        source_proto,
        target_proto,
        source_obfuscated,
        target_obfuscated,
):
    """Manual TC: LVPN-8694"""

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)
    sh.nordvpn(get_alias())

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    sh.nordvpn(get_alias())
    status_info = daemon.get_status_data()

    assert target_tech.upper() in status_info["current technology"]

    if target_tech == "openvpn":
        assert target_proto.upper() in status_info["current protocol"]
    elif target_tech == "nordwhisper":
        assert "Webtunnel" in status_info["current protocol"]
    else:
        assert "UDP" in status_info["current protocol"]

    disconnect_base_test()


@pytest.mark.parametrize("source_group", lib.STANDARD_GROUPS[-2:])
@pytest.mark.parametrize("target_group", lib.STANDARD_GROUPS[-2:])
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
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
    """Manual TC: LVPN-8681"""

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    connect_base_test(source_group)

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    connect_base_test(target_group)

    disconnect_base_test()


@pytest.mark.parametrize("source_group", lib.ADDITIONAL_GROUPS[-2:])
@pytest.mark.parametrize("target_group", lib.ADDITIONAL_GROUPS[-2:])
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
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
    """Manual TC: LVPN-8682"""

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    connect_base_test(source_group)

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    connect_base_test(target_group)

    disconnect_base_test()


@pytest.mark.parametrize("source_country", lib.COUNTRIES[-2:])
@pytest.mark.parametrize("target_country", lib.COUNTRIES[-2:])
@pytest.mark.parametrize(("source_tech", "source_proto", "source_obfuscated"), lib.STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize(("target_tech", "target_proto", "target_obfuscated"), lib.STANDARD_TECHNOLOGIES)
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
    """Manual TC: LVPN-8689"""

    lib.set_technology_and_protocol(source_tech, source_proto, source_obfuscated)

    connect_base_test(source_country)

    lib.set_technology_and_protocol(target_tech, target_proto, target_obfuscated)

    connect_base_test(target_country)

    disconnect_base_test()
