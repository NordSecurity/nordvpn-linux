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
    assert not firewall.is_active(None, [subnet])
    output_after_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    allowlist.remove_subnet_from_allowlist([subnet])
    assert not firewall.is_active(None, [subnet])
    output_after_delete = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)

    assert output_before_add == output_after_add
    assert output_after_add == output_after_delete


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_allowlist_subnet(tech, proto, obfuscated):
    """Manual TC: LVPN-801"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    my_ip = network.get_external_device_ip()

    sh.nordvpn.connect()
    assert network.is_connected()
    assert my_ip != network.get_external_device_ip()

    ip_provider_addresses = socket.gethostbyname_ex(urlparse(lib.API_EXTERNAL_IP).netloc)[2]
    ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]
    allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet)
    assert firewall.is_active(None, ip_addresses_with_subnet)
    assert my_ip == network.get_external_device_ip()

    sh.nordvpn.disconnect()
    assert not firewall.is_active(None, ip_addresses_with_subnet)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_allowlist_subnet_connect(tech, proto, obfuscated):
    """Manual TC: LVPN-8785"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    my_ip = network.get_external_device_ip()

    ip_provider_addresses = socket.gethostbyname_ex(urlparse(lib.API_EXTERNAL_IP).netloc)[2]
    ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

    allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet)
    assert not firewall.is_active(None, ip_addresses_with_subnet)

    sh.nordvpn.connect()
    assert firewall.is_active(None, ip_addresses_with_subnet)
    assert my_ip == network.get_external_device_ip()

    sh.nordvpn.disconnect()
    assert not firewall.is_active(None, ip_addresses_with_subnet)


@pytest.mark.parametrize("subnet", lib.SUBNETS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_allowlist_subnet_twice_disconnected(tech, proto, obfuscated, subnet):
    """Manual TC: LVPN-3766"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    allowlist.add_subnet_to_allowlist([subnet])

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(allowlist.get_alias(), "add", "subnet", subnet)

    expected_message = allowlist.MSG_ALLOWLIST_SUBNET_ADD_ERROR % subnet
    assert expected_message in str(ex)
    assert str(sh.nordvpn.settings()).count(subnet) == 1
    assert not firewall.is_active(None, [subnet])


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
    assert expected_message in str(ex)
    assert str(sh.nordvpn.settings()).count(subnet) == 1
    assert firewall.is_active(None, [subnet])

    sh.nordvpn.disconnect()
    assert not firewall.is_active(None, [subnet])


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_allowlist_subnet_and_remove_disconnected(tech, proto, obfuscated):
    """Manual TC: LVPN-8788"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    ip_provider_addresses = socket.gethostbyname_ex(urlparse(lib.API_EXTERNAL_IP).netloc)[2]
    ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

    allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet)
    assert not firewall.is_active(None, ip_addresses_with_subnet)

    allowlist.remove_subnet_from_allowlist(ip_addresses_with_subnet)
    assert not firewall.is_active(None, ip_addresses_with_subnet)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_allowlist_subnet_and_remove_connected(tech, proto, obfuscated):
    """Manual TC: LVPN-786"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    my_ip = network.get_external_device_ip()

    sh.nordvpn.connect()
    assert my_ip != network.get_external_device_ip()

    ip_provider_addresses = socket.gethostbyname_ex(urlparse(lib.API_EXTERNAL_IP).netloc)[2]
    ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

    allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet)
    assert firewall.is_active(None, ip_addresses_with_subnet)
    assert my_ip == network.get_external_device_ip()

    allowlist.remove_subnet_from_allowlist(ip_addresses_with_subnet)
    assert firewall.is_active() and not firewall.is_active(None, ip_addresses_with_subnet)
    assert my_ip != network.get_external_device_ip()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
def test_allowlist_subnet_remove_nonexistent_disconnected(tech, proto, obfuscated, subnet):
    """Manual TC: LVPN-3768"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(allowlist.get_alias(), "remove", "subnet", subnet)

    expected_message = allowlist.MSG_ALLOWLIST_SUBNET_REMOVE_ERROR % subnet
    assert expected_message in str(ex)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
def test_allowlist_subnet_remove_nonexistent_connected(tech, proto, obfuscated, subnet):
    """Manual TC: LVPN-8787"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(allowlist.get_alias(), "remove", "subnet", subnet)

    expected_message = allowlist.MSG_ALLOWLIST_SUBNET_REMOVE_ERROR % subnet
    assert expected_message in str(ex)
