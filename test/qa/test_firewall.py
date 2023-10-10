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
    sh.nordvpn.set("lan-discovery", "off", _ok_code=(0,1))


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_01():
    lib.set_firewall("on")
    assert not firewall.is_active()

    output = sh.nordvpn.connect()

    print(output)
    assert lib.is_connect_successful(output)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert firewall.is_active()
        lib.set_firewall("off")
        assert not firewall.is_active()
        assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    assert not firewall.is_active()


@pytest.mark.parametrize("port", lib.PORTS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_02_allowlist_port(port):
    lib.set_firewall("on")
    lib.add_port_to_allowlist(port)

    with lib.ErrorDefer(lib.flush_allowlist):
        assert not firewall.is_active(port)

    output = sh.nordvpn.connect()

    print(output)
    assert lib.is_connect_successful(output)
    assert firewall.is_active(port)

    with lib.ErrorDefer(lib.flush_allowlist):
        lib.set_firewall("off")
        assert not firewall.is_active(port)
        with lib.ErrorDefer(sh.nordvpn.disconnect):
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    with lib.ErrorDefer(lib.flush_allowlist):
        assert not firewall.is_active(port)

    lib.flush_allowlist()


@pytest.mark.parametrize("ports", lib.PORTS_RANGE)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(100)
def test_firewall_03_allowlist_ports_range(ports):
    lib.set_firewall("on")
    lib.add_ports_range_to_allowlist(ports)

    with lib.ErrorDefer(lib.flush_allowlist):
        assert not firewall.is_active(ports)

    output = sh.nordvpn.connect()

    print(output)
    assert lib.is_connect_successful(output)

    with lib.ErrorDefer(lib.flush_allowlist):
        assert firewall.is_active(ports)
        lib.set_firewall("off")
        assert not firewall.is_active(ports)
        with lib.ErrorDefer(sh.nordvpn.disconnect):
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    with lib.ErrorDefer(lib.flush_allowlist):
        assert not firewall.is_active(ports)

    lib.flush_allowlist()


@pytest.mark.parametrize("port", lib.PORTS)
@pytest.mark.parametrize("protocol", lib.PROTOCOLS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_04_allowlist_port_and_protocol(port, protocol):
    protocol = str(protocol)
    lib.set_firewall("on")
    lib.add_port_and_protocol_to_allowlist(port, protocol)

    with lib.ErrorDefer(lib.flush_allowlist):
        assert not firewall.is_active(port, protocol)

    output = sh.nordvpn.connect()

    print(output)
    assert lib.is_connect_successful(output)

    with lib.ErrorDefer(lib.flush_allowlist):
        assert firewall.is_active(port, protocol)
        lib.set_firewall("off")
        assert not firewall.is_active(port, protocol)
        with lib.ErrorDefer(sh.nordvpn.disconnect):
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    with lib.ErrorDefer(lib.flush_allowlist):
        assert not firewall.is_active(port, protocol)

    lib.flush_allowlist()


@pytest.mark.parametrize("subnet_addr", lib.SUBNETS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_05_allowlist_subnet(subnet_addr):
    lib.set_firewall("on")
    lib.add_subnet_to_allowlist(subnet_addr)

    with lib.ErrorDefer(lib.flush_allowlist):
        assert not firewall.is_active("", "", subnet_addr)

    output = sh.nordvpn.connect()

    print(output)
    assert lib.is_connect_successful(output)

    with lib.ErrorDefer(lib.flush_allowlist):
        assert firewall.is_active("", "", subnet_addr)
        lib.set_firewall("off")
        assert not firewall.is_active("", "", subnet_addr)
        with lib.ErrorDefer(sh.nordvpn.disconnect):
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    with lib.ErrorDefer(lib.flush_allowlist):
        assert not firewall.is_active("", "", subnet_addr)

    lib.flush_allowlist()

def test_firewall_06_with_killswitch():
    lib.set_firewall("on")
    assert not firewall.is_active()

    lib.set_killswitch("on")

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert firewall.is_active()

    lib.set_killswitch("off")
    assert not firewall.is_active()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_07_with_killswitch_while_connected():
    lib.set_firewall("on")
    assert not firewall.is_active()

    lib.set_killswitch("on")

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert firewall.is_active()

    output = sh.nordvpn.connect()

    print(output)
    assert lib.is_connect_successful(output)
    assert firewall.is_active()

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert network.is_connected()

    lib.set_killswitch("off")
    assert firewall.is_active()

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    assert not firewall.is_active()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_exitnode():
    lib.set_firewall("on")
    assert not firewall.is_active()

    lib.set_killswitch("on")

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert firewall.is_active()

    output = sh.nordvpn.connect()

    print(output)
    assert lib.is_connect_successful(output)
    assert firewall.is_active()

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert network.is_connected()

    lib.set_killswitch("off")
    assert firewall.is_active()

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    assert not firewall.is_active()


@pytest.mark.parametrize("before_connect", [True, False])
def test_firewall_lan_discovery(before_connect):
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


def test_firewall_lan_allowlist_interaction():
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
