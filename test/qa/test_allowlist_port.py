import random

import pytest
import sh

import lib
from lib import allowlist, daemon, firewall, info, logging, login
from lib.dynamic_parametrize import dynamic_parametrize


def setup_module(module):  # noqa: ARG001
    firewall.add_and_delete_random_route()


def setup_function(function):  # noqa: ARG001
    daemon.start()
    login.login_as("default")

    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()

    sh.nordvpn.logout("--persist-token")
    sh.nordvpn.set.defaults()
    daemon.stop()


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS + lib.PORTS_RANGE,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_does_not_create_new_routes_when_adding_deleting_port_disconnected(tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output_before_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    allowlist.add_ports_to_allowlist([port])
    assert not firewall.is_active([port])
    output_after_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    allowlist.remove_ports_from_allowlist([port])
    assert not firewall.is_active([port])
    output_after_delete = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)

    assert output_before_add == output_after_add
    assert output_after_add == output_after_delete


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS + lib.PORTS_RANGE,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_does_not_create_new_routes_when_adding_deleting_port_connected(tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    output_before_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    allowlist.add_ports_to_allowlist([port])
    assert firewall.is_active([port])
    output_after_add = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    allowlist.remove_ports_from_allowlist([port])
    assert not firewall.is_active([port])
    output_after_delete = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)

    assert output_before_add == output_after_add
    assert output_after_add == output_after_delete


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_twice_disconnected(tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    allowlist.add_ports_to_allowlist([port])

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if port.protocol == lib.Protocol.ALL:
            sh.nordvpn(allowlist.get_alias(), "add", "port", port.value)
        else:
            sh.nordvpn(allowlist.get_alias(), "add", "port", port.value, "protocol", port.protocol)

    expected_message = allowlist.MSG_ALLOWLIST_PORT_ADD_ERROR % (port.value, port.protocol)
    assert expected_message in str(ex)
    assert str(sh.nordvpn.settings()).count(port.value) == 1
    assert not firewall.is_active([port])


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_twice_connected(tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    allowlist.add_ports_to_allowlist([port])

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if port.protocol == lib.Protocol.ALL:
            sh.nordvpn(allowlist.get_alias(), "add", "port", port.value)
        else:
            sh.nordvpn(allowlist.get_alias(), "add", "port", port.value, "protocol", port.protocol)

    expected_message = allowlist.MSG_ALLOWLIST_PORT_ADD_ERROR % (port.value, port.protocol)
    assert expected_message in str(ex)
    assert str(sh.nordvpn.settings()).count(port.value) == 1
    assert firewall.is_active([port])

    sh.nordvpn.disconnect()
    assert not firewall.is_active([port])


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS + lib.PORTS_RANGE,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_and_remove_disconnected(tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    allowlist.add_ports_to_allowlist([port])
    assert not firewall.is_active([port])

    allowlist.remove_ports_from_allowlist([port])
    assert not firewall.is_active([port])


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS + lib.PORTS_RANGE,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_and_remove_connected(tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    allowlist.add_ports_to_allowlist([port])
    assert firewall.is_active([port])

    allowlist.remove_ports_from_allowlist([port])
    assert firewall.is_active() and not firewall.is_active([port])


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_remove_nonexistent_disconnected(tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if port.protocol == lib.Protocol.ALL:
            sh.nordvpn(allowlist.get_alias(), "remove", "port", port.value)
        else:
            sh.nordvpn(allowlist.get_alias(), "remove", "port", port.value, "protocol", port.protocol)

    expected_message = allowlist.MSG_ALLOWLIST_PORT_REMOVE_ERROR % (port.value, port.protocol)
    assert expected_message in str(ex)

@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_remove_nonexistent_connected(tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if port.protocol == lib.Protocol.ALL:
            sh.nordvpn(allowlist.get_alias(), "remove", "port", port.value)
        else:
            sh.nordvpn(allowlist.get_alias(), "remove", "port", port.value, "protocol", port.protocol)

    expected_message = allowlist.MSG_ALLOWLIST_PORT_REMOVE_ERROR % (port.value, port.protocol)
    assert expected_message in str(ex)


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS_RANGE,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_range_remove_nonexistent_disconnected(tech, proto, obfuscated, port):
    port_range = port.value.split(":")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if port.protocol == lib.Protocol.ALL:
            sh.nordvpn(allowlist.get_alias(), "remove", "ports", port_range[0], port_range[1])
        else:
            sh.nordvpn(allowlist.get_alias(), "remove", "ports", port_range[0], port_range[1], "protocol", port.protocol)

    expected_message = allowlist.MSG_ALLOWLIST_PORT_RANGE_REMOVE_ERROR % (port.value.replace(":", " - "), port.protocol)
    assert expected_message in str(ex)


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS_RANGE,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_range_remove_nonexistent_connected(tech, proto, obfuscated, port):
    port_range = port.value.split(":")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if port.protocol == lib.Protocol.ALL:
            sh.nordvpn(allowlist.get_alias(), "remove", "ports", port_range[0], port_range[1])
        else:
            sh.nordvpn(allowlist.get_alias(), "remove", "ports", port_range[0], port_range[1], "protocol", port.protocol)

    expected_message = allowlist.MSG_ALLOWLIST_PORT_RANGE_REMOVE_ERROR % (port.value.replace(":", " - "), port.protocol)
    assert expected_message in str(ex)

@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS_RANGE,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_range_twice_disconnected(tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    for _ in range(2):
        allowlist.add_ports_to_allowlist([port])

    assert not firewall.is_active([port])


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS_RANGE,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_range_twice_connected(tech, proto, obfuscated, port):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    for _ in range(2):
        allowlist.add_ports_to_allowlist([port])

    assert firewall.is_active([port])


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS_RANGE,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_range_when_port_from_range_already_allowlisted_disconnected(tech, proto, obfuscated, port):
    port_range = port.value.split(":")
    random_port_from_port_range = str(random.randint(int(port_range[0]), int(port_range[1])))

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    already_allowlisted_port = lib.Port(random_port_from_port_range, port.protocol)
    allowlist.add_ports_to_allowlist([already_allowlisted_port])
    assert not firewall.is_active([already_allowlisted_port])

    allowlist.add_ports_to_allowlist([port])
    assert not firewall.is_active([port]) and not firewall.is_active([already_allowlisted_port])


@dynamic_parametrize(["tech", "proto", "obfuscated", "port"], randomized_source=lib.TECHNOLOGIES, ordered_source=lib.PORTS_RANGE,
                     id_pattern="{ordered.protocol}-{ordered.value}-{randomized[0]}-{randomized[1]}-{randomized[2]}", always_pair=lib.TECHNOLOGIES_BASIC1[0])
def test_allowlist_port_range_when_port_from_range_already_allowlisted_connected(tech, proto, obfuscated, port):
    port_range = port.value.split(":")
    random_port_from_port_range = str(random.randint(int(port_range[0]), int(port_range[1])))

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    already_allowlisted_port = lib.Port(random_port_from_port_range, port.protocol)
    allowlist.add_ports_to_allowlist([already_allowlisted_port])
    assert firewall.is_active([already_allowlisted_port])

    allowlist.add_ports_to_allowlist([port])
    assert firewall.is_active([port]) and not firewall.is_active([already_allowlisted_port])
