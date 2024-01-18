from lib import (
    allowlist,
    daemon,
    firewall,
    info,
    logging,
    login,
    network,
)
import lib
import pytest
import sh
import socket
import timeout_decorator
from urllib.parse import urlparse


CIDR_32 = "/32"


def setup_module(module):
    daemon.start()
    login.login_as("default")
    firewall.add_and_delete_random_route()


def teardown_module(module):
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_allowlist_does_not_create_new_routes_when_adding_deleting_subnets_disconnected(allowlist_alias, tech, proto, obfuscated, subnet):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        output_before_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
        allowlist.add_subnet_to_allowlist([subnet], allowlist_alias)
        assert not firewall.is_active(None, [subnet])
        output_after_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
        allowlist.remove_subnet_from_allowlist([subnet], allowlist_alias)
        assert not firewall.is_active(None, [subnet])
        output_after_delete = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)

        assert output_before_add == output_after_add
        assert output_after_add == output_after_delete


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_does_not_create_new_routes_when_adding_deleting_subnets_connected(allowlist_alias, tech, proto, obfuscated, subnet):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()

            output_before_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
            allowlist.add_subnet_to_allowlist([subnet], allowlist_alias)
            output_after_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
            allowlist.remove_subnet_from_allowlist([subnet], allowlist_alias)
            output_after_delete = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)

            assert output_before_add == output_after_add
            assert output_after_add == output_after_delete


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_allowlist_subnet(allowlist_alias, tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    my_ip = network.get_external_device_ip()
    ip_addresses_with_subnet = None

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()
            assert network.is_connected()
            assert my_ip != network.get_external_device_ip()

            ip_provider_addresses = socket.gethostbyname_ex(urlparse(network.API_EXTERNAL_IP).netloc)[2]
            ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

            allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet, allowlist_alias)
            assert firewall.is_active(None, ip_addresses_with_subnet)
            assert my_ip == network.get_external_device_ip()
        assert not firewall.is_active(None, ip_addresses_with_subnet)


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_subnet_connect(allowlist_alias, tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    my_ip = network.get_external_device_ip()
    ip_addresses_with_subnet = None

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            ip_provider_addresses = socket.gethostbyname_ex(urlparse(network.API_EXTERNAL_IP).netloc)[2]
            ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

            allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet, allowlist_alias)
            assert not firewall.is_active(None, ip_addresses_with_subnet)

            sh.nordvpn.connect()
            assert firewall.is_active(None, ip_addresses_with_subnet)
            assert my_ip == network.get_external_device_ip()
    assert not firewall.is_active(None, ip_addresses_with_subnet)


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_allowlist_subnet(allowlist_alias, tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    my_ip = network.get_external_device_ip()
    ip_addresses_with_subnet = None

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()
            assert my_ip != network.get_external_device_ip()

            ip_provider_addresses = socket.gethostbyname_ex(urlparse(network.API_EXTERNAL_IP).netloc)[2]
            ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

            allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet, allowlist_alias)
            assert firewall.is_active(None, ip_addresses_with_subnet)
            assert my_ip == network.get_external_device_ip()
    assert not firewall.is_active(None, ip_addresses_with_subnet)


@pytest.mark.parametrize("subnet", lib.SUBNETS)
@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_allowlist_subnet_twice_disconnected(allowlist_alias, tech, proto, obfuscated, subnet):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        allowlist.add_subnet_to_allowlist([subnet], allowlist_alias)

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn(allowlist_alias, "add", "subnet", subnet)

        expected_message = allowlist.MSG_ALLOWLIST_SUBNET_ADD_ERROR % subnet
        assert expected_message in str(ex)
        assert str(sh.nordvpn.settings()).count(subnet) == 1
        assert not firewall.is_active(None, [subnet])


@pytest.mark.parametrize("subnet", lib.SUBNETS)
@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_subnet_twice_connected(allowlist_alias, tech, proto, obfuscated, subnet):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()

            allowlist.add_subnet_to_allowlist([subnet], allowlist_alias)

            with pytest.raises(sh.ErrorReturnCode_1) as ex:
                sh.nordvpn(allowlist_alias, "add", "subnet", subnet)

            expected_message = allowlist.MSG_ALLOWLIST_SUBNET_ADD_ERROR % subnet
            assert expected_message in str(ex)
            assert str(sh.nordvpn.settings()).count(subnet) == 1
            assert firewall.is_active(None, [subnet])
    assert not firewall.is_active(None, [subnet])


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_allowlist_subnet_and_remove_disconnected(allowlist_alias, tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        ip_provider_addresses = socket.gethostbyname_ex(urlparse(network.API_EXTERNAL_IP).netloc)[2]
        ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

        allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet, allowlist_alias)
        assert not firewall.is_active(None, ip_addresses_with_subnet)

        allowlist.remove_subnet_from_allowlist(ip_addresses_with_subnet, allowlist_alias)
        assert not firewall.is_active(None, ip_addresses_with_subnet)


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_subnet_and_remove_connected(allowlist_alias, tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    my_ip = network.get_external_device_ip()

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()
            assert my_ip != network.get_external_device_ip()

            ip_provider_addresses = socket.gethostbyname_ex(urlparse(network.API_EXTERNAL_IP).netloc)[2]
            ip_addresses_with_subnet = [ip + CIDR_32 for ip in ip_provider_addresses]

            allowlist.add_subnet_to_allowlist(ip_addresses_with_subnet, allowlist_alias)
            assert firewall.is_active(None, ip_addresses_with_subnet)
            assert my_ip == network.get_external_device_ip()

            allowlist.remove_subnet_from_allowlist(ip_addresses_with_subnet, allowlist_alias)
            assert firewall.is_active() and not firewall.is_active(None, ip_addresses_with_subnet)

            assert my_ip != network.get_external_device_ip()
    assert not firewall.is_active()


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
def test_allowlist_subnet_remove_nonexistant_disconnected(allowlist_alias, tech, proto, obfuscated, subnet):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(allowlist_alias, "remove", "subnet", subnet)

    expected_message = allowlist.MSG_ALLOWLIST_SUBNET_REMOVE_ERROR % subnet
    assert expected_message in str(ex)


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_subnet_remove_nonexistant_connected(allowlist_alias, tech, proto, obfuscated, subnet):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn(allowlist_alias, "remove", "subnet", subnet)

        expected_message = allowlist.MSG_ALLOWLIST_SUBNET_REMOVE_ERROR % subnet
        assert expected_message in str(ex)
