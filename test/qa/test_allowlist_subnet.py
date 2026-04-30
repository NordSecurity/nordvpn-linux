import socket
from urllib.parse import urlparse

import pytest
import sh

import lib
from lib import (
    allowlist,
    firewall,
    network,
)

pytestmark = pytest.mark.usefixtures("add_and_delete_random_route", "nordvpnd_scope_function")

CIDR_32 = "/32"


@pytest.mark.parametrize("subnet", lib.SUBNETS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_allowlist_does_not_create_new_routes_when_adding_deleting_subnets_disconnected(tech, proto, obfuscated, subnet):
    """Manual TC: LVPN-8789"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output_before_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    allowlist.add_subnet_to_allowlist([subnet])
    assert not firewall.is_active(), "Subnet should not be active when disconnected"
    output_after_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    allowlist.remove_subnet_from_allowlist([subnet])
    assert not firewall.is_active(), "Subnet should not be active when disconnected"

    output_after_delete = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)

    assert output_before_add == output_after_add, "Route table should not change after adding subnet"
    assert output_after_add == output_after_delete, "Route table should not change after removing subnet"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_allowlist_subnet(tech, proto, obfuscated):
    """Manual TC: LVPN-801"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    my_ip = network.get_external_device_ip()

    sh.nordvpn.connect()
    assert network.is_connected(), "VPN should be connected"
    assert my_ip != network.get_external_device_ip(), "IP should change when connected"

    ip_provider_addresses = socket.gethostbyname_ex(urlparse(lib.API_EXTERNAL_IP).netloc)[2]
    ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]
    allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet)
    assert not firewall.is_ip_routed_via_VPN(ip_addresses_with_subnet), "Subnet should be active when connected"
    assert my_ip == network.get_external_device_ip(), "IP should return to original when subnet is allowlisted"

    sh.nordvpn.disconnect()
    assert not firewall.is_active(), "Subnet should not be active when disconnected"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_allowlist_subnet_connect(tech, proto, obfuscated):
    """Manual TC: LVPN-8785"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    my_ip = network.get_external_device_ip()

    ip_provider_addresses = socket.gethostbyname_ex(urlparse(lib.API_EXTERNAL_IP).netloc)[2]
    ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

    allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet)
    assert not firewall.is_active(), "Firewall should not be active when disconnected"

    sh.nordvpn.connect()
    assert not firewall.is_ip_routed_via_VPN(ip_addresses_with_subnet), "Whitelisted address is not routed thru VPN"
    assert my_ip == network.get_external_device_ip(), "IP should return to original when subnet is allowlisted"

    sh.nordvpn.disconnect()
    assert not firewall.is_active(), "Firewall is not active after disconnect"


@pytest.mark.parametrize("subnet", lib.SUBNETS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_allowlist_subnet_twice_disconnected(tech, proto, obfuscated, subnet):
    """Manual TC: LVPN-3766"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    allowlist.add_subnet_to_allowlist([subnet])

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(allowlist.get_alias(), "add", "subnet", subnet)

    expected_message = allowlist.MSG_ALLOWLIST_SUBNET_ADD_ERROR % subnet
    assert expected_message in ex.value.stdout.decode("utf-8"), "Error message should indicate subnet add failed"
    assert str(sh.nordvpn.settings()).count(subnet) == 1, "Subnet should appear once in settings"
    assert not firewall.is_active(), "Expected to have firewall empty after disconnect"


@pytest.mark.parametrize("subnet", lib.SUBNETS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_allowlist_subnet_twice_connected(tech, proto, obfuscated, subnet):
    """Manual TC: LVPN-8786"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    allowlist.add_subnet_to_allowlist([subnet])

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(allowlist.get_alias(), "add", "subnet", subnet)

    expected_message = allowlist.MSG_ALLOWLIST_SUBNET_ADD_ERROR % subnet
    assert expected_message in ex.value.stdout.decode("utf-8"), "Error message should indicate subnet add failed"
    assert str(sh.nordvpn.settings()).count(subnet) == 1, "Subnet should appear once in settings"
    assert not firewall.is_ip_routed_via_VPN([subnet]), "Whitelisted IP is not routed thru the tunnel"

    sh.nordvpn.disconnect()
    assert not firewall.is_active(), "Firewall is not active after VPN disconnect"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_allowlist_subnet_and_remove_disconnected(tech, proto, obfuscated):
    """Manual TC: LVPN-8788"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    ip_provider_addresses = socket.gethostbyname_ex(urlparse(lib.API_EXTERNAL_IP).netloc)[2]
    ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

    allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet)
    assert not firewall.is_active(), "Firewall is not configured"

    allowlist.remove_subnet_from_allowlist(ip_addresses_with_subnet)
    assert not firewall.is_active(), "Firewall is not configured"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_allowlist_subnet_and_remove_connected(tech, proto, obfuscated):
    """Manual TC: LVPN-786"""

    # TODO: remove conditional timeout, once LVPN-10169 gets fixed
    timeout = 30 if tech.lower() == "nordwhisper" else 5

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    my_ip = network.get_external_device_ip(timeout)

    sh.nordvpn.connect()
    assert my_ip != network.get_external_device_ip(timeout), "IP should change when connected"

    ip_provider_addresses = socket.gethostbyname_ex(urlparse(lib.API_EXTERNAL_IP).netloc)[2]
    ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

    allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet)
    assert not firewall.is_ip_routed_via_VPN(ip_addresses_with_subnet), "Whitelisted IP is not routet thru VPN"
    assert my_ip == network.get_external_device_ip(timeout), "IP should return to original when subnet is allowlisted"

    allowlist.remove_subnet_from_allowlist(ip_addresses_with_subnet)
    assert firewall.is_active() and firewall.is_ip_routed_via_VPN(ip_addresses_with_subnet), "Firewall is configured and whitelisted IP is routed thru the tunnel"
    assert my_ip != network.get_external_device_ip(timeout), "IP is not the real IP address"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
def test_allowlist_subnet_remove_nonexistent_disconnected(tech, proto, obfuscated, subnet):
    """Manual TC: LVPN-3768"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(allowlist.get_alias(), "remove", "subnet", subnet)

    expected_message = allowlist.MSG_ALLOWLIST_SUBNET_REMOVE_ERROR % subnet
    assert expected_message in ex.value.stdout.decode("utf-8"), "Error message should indicate subnet remove failed"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
def test_allowlist_subnet_remove_nonexistent_connected(tech, proto, obfuscated, subnet):
    """Manual TC: LVPN-8787"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(allowlist.get_alias(), "remove", "subnet", subnet)

    expected_message = allowlist.MSG_ALLOWLIST_SUBNET_REMOVE_ERROR % subnet
    assert expected_message in ex.value.stdout.decode("utf-8"), "Error message should indicate subnet remove failed"
