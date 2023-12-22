from lib import (
    allowlist,
    daemon,
    firewall,
    info,
    logging,
    login,
    network
)
import lib
import pytest
import sh
import timeout_decorator


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


@pytest.mark.parametrize("port", lib.PORTS + lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS + lib.PORTS_RANGE])
@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_allowlist_does_not_create_new_routes_when_adding_deleting_port_disconnected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        output_before_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
        allowlist.add_ports_to_allowlist([port], allowlist_alias)
        output_after_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
        allowlist.remove_ports_from_allowlist([port], allowlist_alias)
        output_after_delete = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)

        assert output_before_add == output_after_add
        assert output_after_add == output_after_delete


@pytest.mark.parametrize("port", lib.PORTS + lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS + lib.PORTS_RANGE])
@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_does_not_create_new_routes_when_adding_deleting_port_connected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()
            assert network.is_connected()

            output_before_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
            allowlist.add_ports_to_allowlist([port], allowlist_alias)
            output_after_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
            allowlist.remove_ports_from_allowlist([port], allowlist_alias)
            output_after_delete = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)

            assert output_before_add == output_after_add
            assert output_after_add == output_after_delete


@pytest.mark.parametrize("port", lib.PORTS + lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS + lib.PORTS_RANGE])
@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_allowlist_port_is_not_set_when_disconnected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        assert not firewall.is_active([port])
        allowlist.add_ports_to_allowlist([port], allowlist_alias)
        assert not firewall.is_active([port])


@pytest.mark.parametrize("port", lib.PORTS + lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS + lib.PORTS_RANGE])
@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_port_requires_connection(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()

            assert not firewall.is_active([port])
            allowlist.add_ports_to_allowlist([port], allowlist_alias)
            assert firewall.is_active([port])

        assert not firewall.is_active([port])
