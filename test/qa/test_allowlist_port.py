import random
from lib import (
    allowlist,
    daemon,
    firewall,
    info,
    logging,
    login
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
        assert not firewall.is_active([port])
        output_after_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
        allowlist.remove_ports_from_allowlist([port], allowlist_alias)
        assert not firewall.is_active([port])
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

            output_before_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
            allowlist.add_ports_to_allowlist([port], allowlist_alias)
            assert firewall.is_active([port])
            output_after_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
            allowlist.remove_ports_from_allowlist([port], allowlist_alias)
            assert not firewall.is_active([port])
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


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS])
def test_allowlist_port_twice_disconnected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        allowlist.add_ports_to_allowlist([port], allowlist_alias)

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            if port.protocol == lib.Protocol.ALL:
                sh.nordvpn(allowlist_alias, "add", "port", port.value)
            else:
                sh.nordvpn(allowlist_alias, "add", "port", port.value, "protocol", port.protocol)

        expected_message = allowlist.MSG_ALLOWLIST_PORT_ADD_ERROR % (port.value, port.protocol)
        assert expected_message in str(ex)
        assert str(sh.nordvpn.settings()).count(port.value) == 1
        assert not firewall.is_active([port])


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_port_twice_connected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()

            allowlist.add_ports_to_allowlist([port], allowlist_alias)

            with pytest.raises(sh.ErrorReturnCode_1) as ex:
                if port.protocol == lib.Protocol.ALL:
                    sh.nordvpn(allowlist_alias, "add", "port", port.value)
                else:
                    sh.nordvpn(allowlist_alias, "add", "port", port.value, "protocol", port.protocol)

            expected_message = allowlist.MSG_ALLOWLIST_PORT_ADD_ERROR % (port.value, port.protocol)
            assert expected_message in str(ex)
            assert str(sh.nordvpn.settings()).count(port.value) == 1
            assert firewall.is_active([port])
    assert not firewall.is_active([port])


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS + lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS + lib.PORTS_RANGE])
def test_allowlist_port_and_remove_disconnected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        allowlist.add_ports_to_allowlist([port], allowlist_alias)
        assert not firewall.is_active([port])

        allowlist.remove_ports_from_allowlist([port], allowlist_alias)
        assert not firewall.is_active([port])


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS + lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS + lib.PORTS_RANGE])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_port_and_remove_connected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()

            allowlist.add_ports_to_allowlist([port], allowlist_alias)
            assert firewall.is_active([port])

            allowlist.remove_ports_from_allowlist([port], allowlist_alias)
            assert firewall.is_active() and not firewall.is_active([port])


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS])
def test_allowlist_port_remove_nonexistant_disconnected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if port.protocol == lib.Protocol.ALL:
            sh.nordvpn(allowlist_alias, "remove", "port", port.value)
        else:
            sh.nordvpn(allowlist_alias, "remove", "port", port.value, "protocol", port.protocol)

    expected_message = allowlist.MSG_ALLOWLIST_PORT_REMOVE_ERROR % (port.value, port.protocol)
    assert expected_message in str(ex)


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_port_remove_nonexistant_connected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            if port.protocol == lib.Protocol.ALL:
                sh.nordvpn(allowlist_alias, "remove", "port", port.value)
            else:
                sh.nordvpn(allowlist_alias, "remove", "port", port.value, "protocol", port.protocol)

        expected_message = allowlist.MSG_ALLOWLIST_PORT_REMOVE_ERROR % (port.value, port.protocol)
        assert expected_message in str(ex)


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS_RANGE])
def test_allowlist_port_range_remove_nonexistant_disconnected(allowlist_alias, tech, proto, obfuscated, port):
    port_range = port.value.split(":")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if port.protocol == lib.Protocol.ALL:
            sh.nordvpn(allowlist_alias, "remove", "ports", port_range[0], port_range[1])
        else:
            sh.nordvpn(allowlist_alias, "remove", "ports", port_range[0], port_range[1], "protocol", port.protocol)

    expected_message = allowlist.MSG_ALLOWLIST_PORT_RANGE_REMOVE_ERROR % (port.value.replace(":", " - "), port.protocol)
    assert expected_message in str(ex)


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS_RANGE])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_port_range_remove_nonexistant_connected(allowlist_alias, tech, proto, obfuscated, port):
    port_range = port.value.split(":")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            if port.protocol == lib.Protocol.ALL:
                sh.nordvpn(allowlist_alias, "remove", "ports", port_range[0], port_range[1])
            else:
                sh.nordvpn(allowlist_alias, "remove", "ports", port_range[0], port_range[1], "protocol", port.protocol)

        expected_message = allowlist.MSG_ALLOWLIST_PORT_RANGE_REMOVE_ERROR % (port.value.replace(":", " - "), port.protocol)
        assert expected_message in str(ex)


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS_RANGE])
def test_allowlist_port_range_twice_disconnected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        for x in range(2):
            allowlist.add_ports_to_allowlist([port], allowlist_alias)

        assert not firewall.is_active([port])


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS_RANGE])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_port_range_twice_connected(allowlist_alias, tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()

            for x in range(2):
                allowlist.add_ports_to_allowlist([port], allowlist_alias)

            assert firewall.is_active([port])


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS_RANGE])
def test_allowlist_port_range_when_port_from_range_already_allowlisted_disconnected(allowlist_alias, tech, proto, obfuscated, port):
    port_range = port.value.split(":")
    random_port_from_port_range = str(random.randint(int(port_range[0]), int(port_range[1])))

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        already_allowlisted_port = lib.Port(random_port_from_port_range, port.protocol)
        allowlist.add_ports_to_allowlist([already_allowlisted_port], allowlist_alias)
        assert not firewall.is_active([already_allowlisted_port])

        allowlist.add_ports_to_allowlist([port], allowlist_alias)
        assert not firewall.is_active([port]) and not firewall.is_active([already_allowlisted_port])


@pytest.mark.parametrize("allowlist_alias", lib.ALLOWLIST_ALIAS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS_RANGE, ids=[f"{port.protocol}-{port.value}" for port in lib.PORTS_RANGE])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_port_range_when_port_from_range_already_allowlisted_connected(allowlist_alias, tech, proto, obfuscated, port):
    port_range = port.value.split(":")
    random_port_from_port_range = str(random.randint(int(port_range[0]), int(port_range[1])))

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()

            already_allowlisted_port = lib.Port(random_port_from_port_range, port.protocol)
            allowlist.add_ports_to_allowlist([already_allowlisted_port], allowlist_alias)
            assert firewall.is_active([already_allowlisted_port])

            allowlist.add_ports_to_allowlist([port], allowlist_alias)
            assert firewall.is_active([port]) and not firewall.is_active([already_allowlisted_port])
