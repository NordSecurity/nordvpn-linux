from itertools import product

import pytest
import sh

import lib
from lib import (
    daemon,
    network,
    IS_NIGHTLY
)
from lib.dynamic_parametrize import dynamic_parametrize
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

    assert lib.is_connect_successful(output, name, hostname), "Connection should be successful"
    assert network.is_connected(), "Network should be connected"


pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")


@dynamic_parametrize(
    [
        "target_tech", "target_proto",
        "source_tech", "source_proto",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[lib.TECHNOLOGIES],
    generate_all=IS_NIGHTLY,
    id_pattern="{source_tech}-{source_proto}-"
              "{target_tech}-{target_proto}",
)
def test_reconnect_matrix(
        source_tech,
        target_tech,
        source_proto,
        target_proto,
):
    """Manual TC: LVPN-8674, LVPN-8694, LVPN-676"""
    lib.set_technology_and_protocol(source_tech, source_proto)
    connect_base_test()

    lib.set_technology_and_protocol(target_tech, target_proto)
    connect_base_test()

    status_info = daemon.get_status_data()

    assert status_info["protocol"] == lib.vpn_protocol_display_name(target_tech, target_proto), \
        "Current protocol should match target protocol"

    disconnect_base_test()


@dynamic_parametrize(
    [
        "tech", "proto", "country", "city",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[list(zip(lib.COUNTRIES, lib.CITIES, strict=False))],
    generate_all=IS_NIGHTLY,
    id_pattern="{country}-{city}-"
               "{tech}-{proto}",
)
def test_connect_country_and_city(tech, proto, country, city):
    """Manual TC: LVPN-8610"""

    lib.set_technology_and_protocol(tech, proto)

    connect_base_test(country)
    connect_base_test(city)
    connect_base_test(f"{country} {city}")

    disconnect_base_test()


@dynamic_parametrize(
    [
        "target_tech", "target_proto",
        "source_tech", "source_proto",
    ],
    ordered_source=[lib.STANDARD_TECHNOLOGIES],
    randomized_source=[lib.STANDARD_TECHNOLOGIES],
    generate_all=IS_NIGHTLY,
    id_pattern="{source_tech}-{source_proto}-"
              "{target_tech}-{target_proto}",
)
def test_status_change_technology_and_protocol(
        source_tech,
        target_tech,
        source_proto,
        target_proto,
):
    """Manual TC: LVPN-676"""

    lib.set_technology_and_protocol(source_tech, source_proto)

    sh.nordvpn(get_alias())
    status_info = daemon.get_status_data()

    source_name = lib.vpn_protocol_display_name(source_tech, source_proto)
    assert status_info["protocol"] == source_name, "Current protocol should match source protocol"

    lib.set_technology_and_protocol(target_tech, target_proto)
    status_info = daemon.get_status_data()
    assert status_info["protocol"] == source_name, "Connection protocol should remain source protocol until reconnect"

    disconnect_base_test()


@dynamic_parametrize(
    [
        "target_tech", "target_proto", "target_group",
        "source_tech", "source_proto", "source_group",
    ],
    ordered_source=[[(*tech, group) for tech, group in product(lib.STANDARD_TECHNOLOGIES, lib.ADDITIONAL_GROUPS[-1:])]],
    randomized_source=[[(*tech, group) for tech, group in product(lib.STANDARD_TECHNOLOGIES, lib.ADDITIONAL_GROUPS[-1:])]],
    generate_all=IS_NIGHTLY,
    id_pattern="{source_tech}-{source_proto}-"
               "{target_tech}-{target_proto}-"
               "{source_group}-{target_group}",
)
def test_reconnect_to_additional_group(
    source_tech,
    target_tech,
    source_proto,
    target_proto,
    source_group,
    target_group,
):
    """Manual TC: LVPN-8682"""

    lib.set_technology_and_protocol(source_tech, source_proto)

    connect_base_test(source_group)

    lib.set_technology_and_protocol(target_tech, target_proto)

    connect_base_test(target_group)

    disconnect_base_test()


@dynamic_parametrize(
    [
        "target_tech", "target_proto", "target_country",
        "source_tech", "source_proto", "source_country",
    ],
    ordered_source=[[(*tech, group) for tech, group in product(lib.STANDARD_TECHNOLOGIES, lib.COUNTRIES[-2:])]],
    randomized_source=[[(*tech, group) for tech, group in product(lib.STANDARD_TECHNOLOGIES, lib.COUNTRIES[-2:])]],
    generate_all=IS_NIGHTLY,
    id_pattern="{source_tech}-{source_proto}-"
               "{target_tech}-{target_proto}-"
               "{source_country}-{target_country}",
)
def test_reconnect_to_server_by_country_name(
    source_tech,
    target_tech,
    source_proto,
    target_proto,
    source_country,
    target_country,
):
    """Manual TC: LVPN-8689"""

    lib.set_technology_and_protocol(source_tech, source_proto)

    connect_base_test(source_country)

    lib.set_technology_and_protocol(target_tech, target_proto)

    connect_base_test(target_country)

    disconnect_base_test()
