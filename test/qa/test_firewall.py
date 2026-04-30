import random

import pytest
import sh

import lib
from lib import IS_NIGHTLY, allowlist, firewall, network
from lib.dynamic_parametrize import dynamic_parametrize

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_module", "collect_logs")


def setup_module(module):  # noqa: ARG001
    firewall.setup_port_sock_server(None)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connected_firewall_disable(tech, proto, obfuscated):
    """Manual TC: LVPN-688"""

    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        assert not firewall.is_active(), "Firewall should not be active before connecting"

        sh.nordvpn.connect()
        assert network.is_connected(), "Network should be connected"
        assert firewall.is_active(), "Firewall should be active when connected"

        lib.set_firewall("off")
        assert not firewall.is_active(), "Firewall should be inactive after disabling"
    assert network.is_disconnected(), "Network should be disconnected after context"
    assert not firewall.is_active(), "Firewall should be inactive after disconnecting"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connected_firewall_enable(tech, proto, obfuscated):
    """Manual TC: LVPN-693"""

    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("off")
        assert not firewall.is_active(), "Firewall should not be active when disabled"

        sh.nordvpn.connect()
        assert network.is_connected(), "Network should be connected"
        assert not firewall.is_active(), "Firewall should remain inactive when disabled"

        lib.set_firewall("on")
        assert firewall.is_active(), "Firewall should be active after enabling"
    assert network.is_disconnected(), "Network should be disconnected after context"
    assert not firewall.is_active(), "Firewall should be inactive after disconnecting"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_disable_connect(tech, proto, obfuscated):
    """Manual TC: LVPN-598"""

    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("off")
        assert not firewall.is_active(), "Firewall should not be active when disabled"

        sh.nordvpn.connect()
        assert network.is_connected(), "Network should be connected"
        assert not firewall.is_active(), "Firewall should remain inactive when disabled"
    assert network.is_disconnected(), "Network should be disconnected after context"
    assert not firewall.is_active(), "Firewall should be inactive after disconnecting"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_enable_connect(tech, proto, obfuscated):
    """Manual TC: LVPN-593"""

    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        assert not firewall.is_active(), "Firewall should not be active before connecting"

        sh.nordvpn.connect()
        assert network.is_connected(), "Network should be connected"
        assert firewall.is_active(), "Firewall should be active when connected"
    assert network.is_disconnected(), "Network should be disconnected after context"
    assert not firewall.is_active(), "Firewall should be inactive after disconnecting"


@dynamic_parametrize(
    [
        "tech",
        "proto",
        "obfuscated",
        "port",
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
            assert not firewall.is_active(), "Firewall is not configured"
            assert firewall.is_source_port_reachable([port]), "Whitelisted port is not blocked"

            sh.nordvpn.connect()
            assert network.is_connected(), "VPN is connected and there is internet"
            assert firewall.is_active(), "Firewall is configured"
            assert firewall.is_source_port_reachable([port]), "Whitelisted port is not blocked"

            lib.set_firewall("off")
            assert not firewall.is_active(), "Firewall is not configured"
            # Firewall off means that allowlisted packets are not told to not go through vpn
            assert not firewall.is_source_port_reachable([port]), "Routing to the ports is broken if firewall is off"
        assert network.is_disconnected(), "VPN is disconnected and internet is working"
    assert not firewall.is_active() and firewall.is_source_port_reachable([port]), "Firewall is not configured and whitelisted port is working"


@dynamic_parametrize(
    [
        "tech",
        "proto",
        "obfuscated",
        "ports",
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
            assert not firewall.is_active(), "Firewall is not configured"
            assert firewall.is_source_port_reachable([ports]), "Port is reachable"

            sh.nordvpn.connect()
            assert network.is_connected(), "VPN is connected"
            assert firewall.is_active(), "Firewall is configured"
            assert firewall.is_source_port_reachable([ports]), "Port is reachable outside of the tunnel"

            lib.set_firewall("off")
            assert not firewall.is_active(), "Firewall is not configured"
            assert not firewall.is_source_port_reachable([ports]), "Port routing is broken because firewall is disabled"
        assert network.is_disconnected(), "VPN disconnected"
    assert not firewall.is_active(), "Firewall is not configured"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
def test_firewall_05_allowlist_subnet(tech, proto, obfuscated, subnet):
    """Manual TC: LVPN-8724"""

    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            allowlist.add_subnet_to_allowlist([subnet])
            assert not firewall.is_ip_routed_via_VPN([subnet]), "Whitelisted IP not routed thru VPN"

            sh.nordvpn.connect()
            assert network.is_connected(), "VPN is connected"
            assert not firewall.is_ip_routed_via_VPN([subnet]), "Whitelisted port is not routed thru VPN"

            lib.set_firewall("off")
            assert not firewall.is_ip_routed_via_VPN([subnet]), "Whitelisted port is not routed thru VPN"
        assert network.is_disconnected(), "VPN is disconnected"
    assert not firewall.is_ip_routed_via_VPN([subnet]), "Whitelisted port is not routed thru VPN"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_06_with_killswitch(tech, proto, obfuscated):
    """Manual TC: LVPN-8726"""

    with lib.Defer(sh.nordvpn.set.killswitch.off):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        assert not firewall.is_active(), "Firewall should not be active before killswitch is enabled"

        lib.set_killswitch("on")
        assert firewall.is_active(), "Firewall should be active when killswitch is enabled"
    assert not firewall.is_active(), "Firewall should be inactive after killswitch is disabled"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_07_with_killswitch_while_connected(tech, proto, obfuscated):
    """Manual TC: LVPN-8727"""

    with lib.Defer(sh.nordvpn.set.killswitch.off):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            assert not firewall.is_active(), "Firewall should not be active before killswitch is enabled"

            lib.set_killswitch("on")
            assert firewall.is_active(), "Firewall should be active when killswitch is enabled"

            sh.nordvpn.connect()
            assert network.is_connected(), "Network should be connected"
            assert firewall.is_active(), "Firewall should remain active when connected with killswitch"

            lib.set_killswitch("off")
            assert firewall.is_active(), "Firewall should remain active after killswitch is disabled"
        assert network.is_disconnected(), "Network should be disconnected after context"
    assert not firewall.is_active(), "Firewall should be inactive after killswitch is disabled"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("before_connect", [True, False])
def test_firewall_lan_discovery(tech, proto, obfuscated, before_connect):
    """Manual TC: LVPN-8947"""

    with lib.Defer(lambda: sh.nordvpn.set("lan-discovery", "off", _ok_code=(0, 1))):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)
            rand_lan_subnet = random.choice(firewall.LAN_DISCOVERY_SUBNETS)

            if before_connect:
                sh.nordvpn.set("lan-discovery", "on")

            sh.nordvpn.connect()

            if not before_connect:
                sh.nordvpn.set("lan-discovery", "on")

            assert not firewall.is_ip_routed_via_VPN([rand_lan_subnet])

            sh.nordvpn.set("lan-discovery", "off")

            assert firewall.is_ip_routed_via_VPN([rand_lan_subnet])


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
            assert firewall.is_ip_routed_via_VPN([ip_not_in_subnet])
            sh.nordvpn.set("lan-discovery", "on")
            # with lan discovery on 192.168.200.255 is in the lan discovery 192.168.0.0/16 subnet
            assert not firewall.is_ip_routed_via_VPN([ip_not_in_subnet]), "LAN discovery did not replace existing smaller subnet"

            sh.nordvpn.set("lan-discovery", "off")

            assert firewall.is_ip_routed_via_VPN([ip_not_in_subnet])


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_lan_allowlist_work_together(tech, proto, obfuscated):
    """Manual TC: LVPN-10010"""

    with lib.Defer(lambda: sh.nordvpn.set("lan-discovery", "off", _ok_code=(0, 1))):
        with lib.Defer(sh.nordvpn.disconnect):
            subnet = "1.1.1.1/32"
            with lib.Defer(lambda: sh.nordvpn.allowlist.remove.subnet(subnet, _ok_code=(0, 1))):
                lib.set_technology_and_protocol(tech, proto, obfuscated)

                sh.nordvpn.allowlist.add.subnet(subnet)
                sh.nordvpn.set("lan-discovery", "on")
                sh.nordvpn.connect()
                assert not firewall.is_ip_routed_via_VPN(["1.1.1.1"]), "Allowlisted subnet is not going through VPN"
                assert firewall.is_ip_routed_via_VPN(["1.0.0.1"]), "Not whitelisted subnet is going through VPN"
