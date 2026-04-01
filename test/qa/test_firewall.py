import os
import random

import pytest
import sh

import lib
from lib import (
    allowlist,
    firewall,
    network,
    IS_NIGHTLY
)
from lib.dynamic_parametrize import dynamic_parametrize

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_module", "collect_logs")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connected_firewall_disable(tech, proto, obfuscated):
    """Manual TC: LVPN-688"""

    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        assert not firewall.is_active()

        sh.nordvpn.connect()
        assert network.is_connected()
        assert firewall.is_active()

        lib.set_firewall("off")
        assert not firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connected_firewall_enable(tech, proto, obfuscated):
    """Manual TC: LVPN-693"""

    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("off")
        assert not firewall.is_active()

        sh.nordvpn.connect()
        assert network.is_connected()
        assert not firewall.is_active()

        lib.set_firewall("on")
        assert firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_disable_connect(tech, proto, obfuscated):
    """Manual TC: LVPN-598"""

    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("off")
        assert not firewall.is_active()

        sh.nordvpn.connect()
        assert network.is_connected()
        assert not firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_enable_connect(tech, proto, obfuscated):
    """Manual TC: LVPN-593"""

    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        assert not firewall.is_active()

        sh.nordvpn.connect()
        assert network.is_connected()
        assert firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "port",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[lib.PORTS],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{port.protocol}-{port.value}",
)
def test_firewall_02_allowlist_port(tech, proto, obfuscated, port):
    """Manual TC: LVPN-8722"""

    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            allowlist.add_ports_to_allowlist([port])
            assert not firewall.is_active([port])

            sh.nordvpn.connect()
            assert network.is_connected()
            assert firewall.is_active([port])

            lib.set_firewall("off")
            assert not firewall.is_active([port])
        assert network.is_disconnected()
    assert not firewall.is_active([port])


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "ports",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[lib.PORTS_RANGE],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{ports.protocol}-{ports.value}",
)
def test_firewall_03_allowlist_ports_range(tech, proto, obfuscated, ports):
    """Manual TC: LVPN-8725"""

    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            allowlist.add_ports_to_allowlist([ports])
            assert not firewall.is_active([ports])

            sh.nordvpn.connect()
            assert network.is_connected()
            assert firewall.is_active([ports])

            lib.set_firewall("off")
            assert not firewall.is_active([ports])
        assert network.is_disconnected()
    assert not firewall.is_active([ports])


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
def test_firewall_05_allowlist_subnet(tech, proto, obfuscated, subnet):
    """Manual TC: LVPN-8724"""

    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            allowlist.add_subnet_to_allowlist([subnet])
            assert not firewall.is_active(None, [subnet])

            sh.nordvpn.connect()
            assert network.is_connected()
            assert firewall.is_active(None, [subnet])

            lib.set_firewall("off")
            assert not firewall.is_active(None, [subnet])
        assert network.is_disconnected()
    assert not firewall.is_active(None, [subnet])


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_06_with_killswitch(tech, proto, obfuscated):
    """Manual TC: LVPN-8726"""

    with lib.Defer(sh.nordvpn.set.killswitch.off):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        assert not firewall.is_active()

        lib.set_killswitch("on")
        assert firewall.is_active()
    assert not firewall.is_active()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_07_with_killswitch_while_connected(tech, proto, obfuscated):
    """Manual TC: LVPN-8727"""

    with lib.Defer(sh.nordvpn.set.killswitch.off):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            assert not firewall.is_active()

            lib.set_killswitch("on")
            assert firewall.is_active()

            sh.nordvpn.connect()
            assert network.is_connected()
            assert firewall.is_active()

            lib.set_killswitch("off")
            assert firewall.is_active()
        assert network.is_disconnected()
    assert not firewall.is_active()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("before_connect", [True, False])
def test_firewall_lan_discovery(tech, proto, obfuscated, before_connect):
    """Manual TC: LVPN-8947"""

    with lib.Defer(lambda: sh.nordvpn.set("lan-discovery", "off", _ok_code=(0, 1))):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)
            rand_lan_subnet = random.choice(firewall.LAN_DISCOVERY_SUBNETS)
            pre_allow_out = sh.ip.route.get(rand_lan_subnet)
            if before_connect:
                sh.nordvpn.set("lan-discovery", "on")

            sh.nordvpn.connect()

            if not before_connect:
                sh.nordvpn.set("lan-discovery", "on")

            assert pre_allow_out == sh.ip.route.get(rand_lan_subnet), "add meainingful assert message later"

            sh.nordvpn.set("lan-discovery", "off")

            assert pre_allow_out != sh.ip.route.get(rand_lan_subnet)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_lan_allowlist_interaction(tech, proto, obfuscated):
    """Manual TC: LVPN-8941"""

    with lib.Defer(lambda: sh.nordvpn.set("lan-discovery", "off", _ok_code=(0, 1))):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            sh.nordvpn.connect()

            subnet = "192.168.0.0/18"
            # 192.168.200.255 is routed through tunnel iface since it it not in the 192.168.0.0/18 subnet
            ip_not_in_subnet = "192.168.200.255"
            sh.nordvpn.allowlist.add.subnet(subnet)
            routed_through_tunnel_out = sh.ip.route.get(ip_not_in_subnet)
            sh.nordvpn.set("lan-discovery", "on")
            routed_through_eth_out = sh.ip.route.get(ip_not_in_subnet)
            # with lan discovery on 192.168.200.255 is in the lan discovery 192.168.0.0/16 subnet
            assert routed_through_tunnel_out != routed_through_eth_out, "LAN discovery did not replace existing smaller subnet"

            sh.nordvpn.set("lan-discovery", "off")

            assert routed_through_tunnel_out == sh.ip.route.get(ip_not_in_subnet)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_lan_allowlist_work_together(tech, proto, obfuscated):
    """Manual TC: LVPN-10010"""

    with lib.Defer(lambda: sh.nordvpn.set("lan-discovery", "off", _ok_code=(0, 1))):
        with lib.Defer(sh.nordvpn.disconnect):
            subnet = "1.1.1.1/32"
            with lib.Defer(lambda: sh.nordvpn.allowlist.remove.subnet(subnet, _ok_code=(0, 1))):

                lib.set_technology_and_protocol(tech, proto, obfuscated)
                pre_allow_out = sh.ip.route.get("1.1.1.1")

                sh.nordvpn.allowlist.add.subnet(subnet)
                sh.nordvpn.set("lan-discovery", "on")
                sh.nordvpn.connect()
                assert pre_allow_out == sh.ip.route.get("1.1.1.1"), "Allowlisted subnet is not going through default interface"
                assert pre_allow_out != sh.ip.route.get("1.0.0.1")
