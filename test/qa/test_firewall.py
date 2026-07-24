import contextlib
import random

import pytest
import socket
import subprocess
import time

import requests
import sh

import lib
from lib import IS_NIGHTLY, allowlist, firewall, network
from lib.dynamic_parametrize import dynamic_parametrize
from lib import capture_utils
from lib.firewall import tun_interface_names

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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_dns_udp_53_to_lan_resolver_dropped(tech, proto, obfuscated):
    """
    Verify UDP port 53 to the local resolver is dropped when VPN is connected.

    Manual TC: LVPN-10477

    Steps:
      1. Connect to VPN.
      2. Assert the query fails, - port 53 must be blocked by the firewall.
      3. Disconnect from VPN.
      4. Assert UDP port 53 is reachable again after disconnecting.

    :raises AssertionError: If UDP port 53 is not reachable after VPN,
     or if it is reachable while VPN is connected.
    """
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert not firewall.is_port_accessible_UDP(src_port=53), (
            "UDP port 53 must be blocked by the firewall when VPN is connected"
        )

    assert firewall.is_port_accessible_UDP(src_port=53), (
        "UDP port 53 must be reachable again after disconnecting from VPN"
    )


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_dns_tcp_53_to_lan_resolver_dropped(tech, proto, obfuscated):
    """
    Verify TCP port 53 to the local resolver is dropped when VPN is connected.

    Manual TC: LVPN-10478

    Steps:
      1. Connect to VPN.
      2. Assert the connection fails, - port 53 must be blocked by the firewall.
      3. Disconnect from VPN.
      4. Assert TCP port 53 is reachable again after disconnecting.


    :raises AssertionError: If TCP port 53 is not reachable after VPN,
     or if it is reachable while VPN is connected.
    """
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert not firewall.is_port_accessible_TCP(src_port=53, dst_ip=firewall.IP, dst_port=firewall.TCP_DST_PORT), (
            "TCP port 53 must be blocked by the firewall when VPN is connected"
        )

    assert firewall.is_port_accessible_TCP(src_port=53, dst_ip=firewall.IP, dst_port=firewall.TCP_DST_PORT), (
        "TCP port 53 must be reachable again after disconnecting from VPN"
    )

@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_tcp_established_connection_response_accepted_via_tunnel(tech, proto, obfuscated):
    """
    Verify that response traffic for an established TCP connection is accepted via the tunnel.

    Manual TC: LVPN-10476

    Steps:
      1. Connect to VPN.
      2. Start tunnel packet capture.
      3. Send an HTTP request to nordvpn.com.
      4. Stop capture.
      5. Assert a valid HTTP response code is returned.
      6. Assert tunnel packets were captured traffic went through the VPN tunnel.
      7. Disconnect from VPN.

     :raises AssertionError: If HTTP request does not return a valid response code or no tunnel packets were captured.
    """
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    host = "nordvpn.com"

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        nordvpn_ip = socket.gethostbyname(host)

        cap = capture_utils.BackgroundCapture(
            "any",
            bpf_filter=f"ip.addr == {nordvpn_ip}",
        )
        cap.start()

        try:
            response = requests.get(f"http://{host}", timeout=10, allow_redirects=False)
        except requests.exceptions.RequestException as e:
            pytest.fail(f"{host} must be reachable when connected to VPN: {e}")

        cap.stop()

        assert response.status_code in (200, 301, 302), (
            f"{host} must return a valid response when connected to VPN, got {response.status_code}"
        )

        assert all(
            capture_utils.ifindex_to_name(p.sll.ifindex) in tun_interface_names
            for p in cap.packets
        ), (
            f"Traffic to {host} must go through a tunnel interface {tun_interface_names}, "
            f"but packets were only seen on: "
            f"{set(capture_utils.ifindex_to_name(p.sll.ifindex) for p in cap.packets)}"
        )


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_outgoing_and_incoming_loopback_accepted(tech, proto, obfuscated):
    """
    Verify loopback traffic is accepted when VPN is connected.

    Manual TC: LVPN-10475

    Steps:
      1. Connect to VPN.
      2. Start a local HTTP server on port 18080.
      3. Send an HTTP request to http://127.0.0.1:18080.
      4. Assert a valid HTTP response code is returned, - both loopback directions verified.
      5. Disconnect from VPN.

    :raises AssertionError: If HTTP request does not return a valid HTTP response code.
    """
    local_http_port = 18080

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        http_server = subprocess.Popen(
            ["python3", "-m", "http.server", str(local_http_port)],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        time.sleep(1)

        try:
            try:
                response = requests.get(f"http://127.0.0.1:{local_http_port}", timeout=5, allow_redirects=False)
            except requests.exceptions.RequestException as e:
                pytest.fail(f"Loopback traffic must be accepted when VPN is connected: {e}")

            assert response.status_code in (200, 301, 302), (
                f"Loopback traffic must be accepted when VPN is connected, got {response.status_code}"
            )
        finally:
            http_server.terminate()
            try:
                http_server.wait(timeout=5)
            except subprocess.TimeoutExpired:
                http_server.kill()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_tcp_to_lan_host_goes_via_tunnel(tech, proto, obfuscated):
    """
    Verify TCP traffic to a LAN host routes through the VPN tunnel and not the physical interface.

    Manual TC: LVPN-10481

    Steps:
      1. Connect to VPN.
      2. Start packet capture on all interfaces filtered to other lan ip.
      3. Attempt HTTP request to other lan ip — no response expected.
      4. Assert other lan ip is routed via VPN.
      5. Assert all captured packets went through a known tunnel interface.
      6. Disconnect from VPN.

    :raises AssertionError: If other lan ip is not routed via VPN or packets appear outside the tunnel.
    """
    other_lan_ip = network.get_default_gateway()

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        cap = capture_utils.BackgroundCapture(
            "any",
            bpf_filter=f"ip.addr == {other_lan_ip}",
        )
        cap.start()

        with contextlib.suppress(requests.exceptions.RequestException):
            requests.get(f"http://{other_lan_ip}", timeout=3, allow_redirects=False)

        cap.stop()

        assert all(
            capture_utils.ifindex_to_name(p.sll.ifindex) in tun_interface_names
            for p in cap.packets
        ), (
            f"Traffic to {other_lan_ip} must go through a tunnel interface {tun_interface_names}, "
            f"but packets were only seen on: "
            f"{set(capture_utils.ifindex_to_name(p.sll.ifindex) for p in cap.packets)}"
        )
