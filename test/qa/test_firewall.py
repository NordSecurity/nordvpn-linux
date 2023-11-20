from lib import (
    daemon,
    info,
    logging,
    login,
    network,
    firewall,
)
import lib
import pytest
import sh
import timeout_decorator


def setup_module(module):
    daemon.start()
    login.login_as("default")


def teardown_module(module):
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connected_firewall_disable(tech, proto, obfuscated):
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


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connected_firewall_enable(tech, proto, obfuscated):
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


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_disable_connect(tech, proto, obfuscated):
    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("off")
        assert not firewall.is_active()

        sh.nordvpn.connect()
        assert network.is_connected()
        assert not firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_enable_connect(tech, proto, obfuscated):
    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        assert not firewall.is_active()

        sh.nordvpn.connect()
        assert network.is_connected()
        assert firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_02_allowlist_port(tech, proto, obfuscated, port):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.add_port_to_allowlist(port)
            assert not firewall.is_active(port)

            sh.nordvpn.connect()
            assert network.is_connected()
            assert firewall.is_active(port)

            lib.set_firewall("off")
            assert not firewall.is_active(port)
        assert network.is_disconnected()
    assert not firewall.is_active(port)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("ports", lib.PORTS_RANGE)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(100)
def test_firewall_03_allowlist_ports_range(tech, proto, obfuscated, ports):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.add_ports_range_to_allowlist(ports)
            assert not firewall.is_active(ports)

            sh.nordvpn.connect()
            assert network.is_connected()
            assert firewall.is_active(ports)

            lib.set_firewall("off")
            assert not firewall.is_active(ports)
        assert network.is_disconnected()
    assert not firewall.is_active(ports)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS)
@pytest.mark.parametrize("protocol", lib.PROTOCOLS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_04_allowlist_port_and_protocol(tech, proto, obfuscated, port, protocol):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            protocol = str(protocol)
            lib.set_firewall("on")
            lib.add_port_and_protocol_to_allowlist(port, protocol)
            assert not firewall.is_active(port, protocol)

            sh.nordvpn.connect()
            assert network.is_connected()
            assert firewall.is_active(port, protocol)

            lib.set_firewall("off")
            assert not firewall.is_active(port, protocol)
        assert network.is_disconnected()
    assert not firewall.is_active(port, protocol)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("subnet_addr", lib.SUBNETS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_05_allowlist_subnet(tech, proto, obfuscated, subnet_addr):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.add_subnet_to_allowlist(subnet_addr)
            assert not firewall.is_active("", "", subnet_addr)

            sh.nordvpn.connect()
            assert network.is_connected()
            assert firewall.is_active("", "", subnet_addr)

            lib.set_firewall("off")
            assert not firewall.is_active("", "", subnet_addr)
        assert network.is_disconnected()
    assert not firewall.is_active("", "", subnet_addr)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_firewall_06_with_killswitch(tech, proto, obfuscated):
    with lib.Defer(sh.nordvpn.set.killswitch.off):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        assert not firewall.is_active()

        lib.set_killswitch("on")
        assert firewall.is_active()
    assert not firewall.is_active()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_07_with_killswitch_while_connected(tech, proto, obfuscated):
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


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.parametrize("before_connect", [True, False])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_lan_discovery(tech, proto, obfuscated, before_connect):
    with lib.Defer(lambda: sh.nordvpn.set("lan-discovery", "off", _ok_code=(0,1))):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            if before_connect:
                sh.nordvpn.set("lan-discovery", "on")

            sh.nordvpn.connect()

            if not before_connect:
                sh.nordvpn.set("lan-discovery", "on")

            rules = sh.sudo.iptables("-S", "INPUT")
            for rule in firewall.inputLanDiscoveryRules:
                assert rule in rules, f"{rule} input rule not found in iptables."

            rules = sh.sudo.iptables("-S", "OUTPUT")
            for rule in firewall.outputLanDiscoveryRules:
                assert rule in rules, f"{rule} output rule not found in iptables"

            sh.nordvpn.set("lan-discovery", "off")

            rules = sh.sudo.iptables("-S", "INPUT")
            for rule in firewall.inputLanDiscoveryRules:
                assert rule not in rules, f"{rule} input rule not found in iptables."

            rules = sh.sudo.iptables("-S", "OUTPUT")
            for rule in firewall.outputLanDiscoveryRules:
                assert rule not in rules, f"{rule} output rule not found in iptables"


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_lan_allowlist_interaction(tech, proto, obfuscated):
    with lib.Defer(lambda: sh.nordvpn.set("lan-discovery", "off", _ok_code=(0,1))):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            sh.nordvpn.connect()

            subnet = "192.168.0.0/18"

            sh.nordvpn.allowlist.add.subnet(subnet)
            sh.nordvpn.set("lan-discovery", "on")

            rules = sh.sudo.iptables("-S", "INPUT")
            assert f"-A INPUT -s {subnet} -i eth0 -m comment --comment nordvpn -j ACCEPT" not in rules, "Whitelist rule was not removed from the INPUT chain when LAN discovery was enabled."

            rules = sh.sudo.iptables("-S", "OUTPUT")
            assert f"-A OUTPUT -s {subnet} -o eth0 -m comment --comment nordvpn -j ACCEPT" not in rules, "Whitelist rule was not removed from the OUTPUT chain when LAN discovery was enabled."

            sh.nordvpn.set("lan-discovery", "off")

            rules = sh.sudo.iptables("-S", "INPUT")
            for rule in firewall.inputLanDiscoveryRules:
                assert rule not in rules, f"{rule} input rule not found in iptables."

            rules = sh.sudo.iptables("-S", "OUTPUT")
            for rule in firewall.outputLanDiscoveryRules:
                assert rule not in rules, f"{rule} output rule not found in iptables"