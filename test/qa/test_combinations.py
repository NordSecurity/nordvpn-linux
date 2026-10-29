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
from lib import capture_utils
from test_routing import get_network_interface
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
        "target_vpn_protocol",
        "source_vpn_protocol",
    ],
    ordered_source=[lib.VPN_PROTOCOLS],
    randomized_source=[lib.VPN_PROTOCOLS],
    generate_all=IS_NIGHTLY,
    id_pattern="{source_vpn_protocol}-"
              "{target_vpn_protocol}",
)
def test_reconnect_matrix(
        source_vpn_protocol,
        target_vpn_protocol,
):
    """Manual TC: LVPN-8674, LVPN-8694, LVPN-676"""
    def verify_ovpn_traffic(vpn_protocol):
        """Verify OpenVPN tunnel traffic uses the expected transport protocol."""
        cap = capture_utils.BackgroundCapture("any", display_filter=f"ip.addr == {daemon.get_status_data()['ip']}")
        cap.start()
        network.generate_traffic()
        cap.stop()

        assert len(cap.packets) > 0, f"No packets captured for {vpn_protocol}"

        for p in cap.packets:
            if vpn_protocol == "openvpn_tcp":
                assert p.ip.proto == capture_utils.IPProtocol.TCP, "TCP packets should travel to VPN server"
            else:
                assert p.ip.proto == capture_utils.IPProtocol.UDP, "UDP packets should travel to VPN server"

    lib.set_vpn_protocol(source_vpn_protocol)
    connect_base_test()
    assert get_network_interface(source_vpn_protocol) in sh.ip.a(), f"{get_network_interface(source_vpn_protocol)} should be present in `ip a` output"
    if source_vpn_protocol in lib.OVPN_VPN_PROTOCOLS:
        verify_ovpn_traffic(source_vpn_protocol)

    lib.set_vpn_protocol(target_vpn_protocol)
    connect_base_test()
    assert get_network_interface(target_vpn_protocol) in sh.ip.a(), f"{get_network_interface(target_vpn_protocol)} should be present in `ip a` output"
    if target_vpn_protocol in lib.OVPN_VPN_PROTOCOLS:
        verify_ovpn_traffic(target_vpn_protocol)

    status_info = daemon.get_status_data()

    assert status_info["protocol"] == lib.vpn_protocol_display_name(target_vpn_protocol), \
        "Current protocol should match target protocol"

    disconnect_base_test()


@dynamic_parametrize(
    [
        "vpn_protocol", "country", "city",
    ],
    ordered_source=[lib.VPN_PROTOCOLS],
    randomized_source=[list(zip(lib.COUNTRIES, lib.CITIES, strict=False))],
    generate_all=IS_NIGHTLY,
    id_pattern="{country}-{city}-"
               "{vpn_protocol}",
)
def test_connect_country_and_city(vpn_protocol, country, city):
    """Manual TC: LVPN-8610"""

    lib.set_vpn_protocol(vpn_protocol)

    connect_base_test(country)
    assert country.replace("_", " ") in sh.nordvpn.status(), "We should be connected to country, that was specified during connect"
    connect_base_test(city)
    assert city.replace("_", " ") in sh.nordvpn.status(), "We should be connected to city, that was specified during connect"
    connect_base_test(f"{country} {city}")
    status = sh.nordvpn.status()
    assert city.replace("_", " ") in status, "We should be connected to city, that was specified during connect"
    assert country.replace("_", " ") in status, "We should be connected to country, that was specified during connect"

    disconnect_base_test()


@dynamic_parametrize(
    [
        "target_vpn_protocol",
        "source_vpn_protocol",
    ],
    ordered_source=[lib.STANDARD_VPN_PROTOCOLS],
    randomized_source=[lib.STANDARD_VPN_PROTOCOLS],
    generate_all=IS_NIGHTLY,
    id_pattern="{source_vpn_protocol}-"
              "{target_vpn_protocol}",
)
def test_status_change_protocol(
        source_vpn_protocol,
        target_vpn_protocol,
):
    """Manual TC: LVPN-676"""

    lib.set_vpn_protocol(source_vpn_protocol)

    sh.nordvpn(get_alias())
    status_info = daemon.get_status_data()

    source_name = lib.vpn_protocol_display_name(source_vpn_protocol)
    assert status_info["protocol"] == source_name, "Current protocol should match source protocol"

    lib.set_vpn_protocol(target_vpn_protocol)
    status_info = daemon.get_status_data()
    assert status_info["protocol"] == source_name, "Connection protocol should remain source protocol until reconnect"

    disconnect_base_test()


@dynamic_parametrize(
    [
        "target_vpn_protocol", "target_group",
        "source_vpn_protocol", "source_group",
    ],
    ordered_source=[[(vpn_protocol, group) for vpn_protocol, group in product(lib.STANDARD_VPN_PROTOCOLS, lib.ADDITIONAL_GROUPS[-1:])]],
    randomized_source=[[(vpn_protocol, group) for vpn_protocol, group in product(lib.STANDARD_VPN_PROTOCOLS, lib.ADDITIONAL_GROUPS[-1:])]],
    generate_all=IS_NIGHTLY,
    id_pattern="{source_vpn_protocol}-"
               "{target_vpn_protocol}-"
               "{source_group}-{target_group}",
)
def test_reconnect_to_additional_group(
    source_vpn_protocol,
    target_vpn_protocol,
    source_group,
    target_group,
):
    """Manual TC: LVPN-8682"""

    lib.set_vpn_protocol(source_vpn_protocol)

    connect_base_test(source_group)

    # Standard_VPN_Servers does not show in status output
    if "Standard_VPN_Servers" not in source_group:
        assert str(source_group).replace("_", " ") in sh.nordvpn.status(), "We should be connected to group, that was specified during connect"

    lib.set_vpn_protocol(target_vpn_protocol)

    connect_base_test(target_group)

    if "Standard_VPN_Servers" not in target_group:
        assert str(target_group).replace("_", " ") in sh.nordvpn.status(), "We should be connected to group, that was specified during connect"

    disconnect_base_test()


@dynamic_parametrize(
    [
        "target_vpn_protocol", "target_country",
        "source_vpn_protocol", "source_country",
    ],
    ordered_source=[[(vpn_protocol, group) for vpn_protocol, group in product(lib.STANDARD_VPN_PROTOCOLS, lib.COUNTRIES[-2:])]],
    randomized_source=[[(vpn_protocol, group) for vpn_protocol, group in product(lib.STANDARD_VPN_PROTOCOLS, lib.COUNTRIES[-2:])]],
    generate_all=IS_NIGHTLY,
    id_pattern="{source_vpn_protocol}-"
               "{target_vpn_protocol}-"
               "{source_country}-{target_country}",
)
def test_reconnect_to_server_by_country_name(
    source_vpn_protocol,
    target_vpn_protocol,
    source_country,
    target_country,
):
    """Manual TC: LVPN-8689"""

    lib.set_vpn_protocol(source_vpn_protocol)

    connect_base_test(source_country)

    assert source_country.replace("_", " ") in sh.nordvpn.status(), "We should be connected to country, that was specified during connect"

    lib.set_vpn_protocol(target_vpn_protocol)

    connect_base_test(target_country)

    assert target_country.replace("_", " ") in sh.nordvpn.status(), "We should be connected to country, that was specified during connect"

    disconnect_base_test()
