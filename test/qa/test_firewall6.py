from lib import (
    daemon,
    info,
    logging,
    login,
    network,
    firewall,
)
import lib
import random
import pytest
import sh
import timeout_decorator


def setup_module(module):
    daemon.start()
    login.login_as("default")
    lib.set_technology_and_protocol("openvpn", "udp", "off")


def teardown_module(module):
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_01():
    lib.set_firewall("on")
    lib.set_ipv6("on")
    assert not firewall.is_active()

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        print(output)
        assert lib.is_connect_successful(output)
        assert network.is_ipv4_and_ipv6_connected(20)

    assert firewall.is_active()

    output = sh.nordvpn.disconnect()
    print(output)
    lib.set_ipv6("off")
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    assert not firewall.is_active()


@pytest.mark.parametrize("port", lib.PORTS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_02_allowlist_port(port):
    lib.set_firewall("on")
    lib.set_ipv6("on")
    lib.add_port_to_allowlist(port)
    assert not firewall.is_active(port)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        print(output)
        assert lib.is_connect_successful(output)
        assert network.is_ipv4_and_ipv6_connected(20)

    assert firewall.is_active(port)

    lib.set_firewall("off")

    assert not firewall.is_active(port)

    with lib.ErrorDefer(lib.flush_allowlist):
        with lib.ErrorDefer(sh.nordvpn.disconnect):
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    lib.set_ipv6("off")
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    assert not firewall.is_active(port)
    lib.flush_allowlist()


@pytest.mark.parametrize("ports", lib.PORTS_RANGE)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_03_allowlist_ports_range(ports):
    lib.set_firewall("on")
    lib.set_ipv6("on")
    lib.add_ports_range_to_allowlist(ports)
    assert not firewall.is_active(ports)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        print(output)
        assert lib.is_connect_successful(output)
        assert network.is_ipv4_and_ipv6_connected(20)

    assert firewall.is_active(ports)

    lib.set_firewall("off")

    assert not firewall.is_active(ports)

    with lib.ErrorDefer(lib.flush_allowlist):
        with lib.ErrorDefer(sh.nordvpn.disconnect):
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    lib.set_ipv6("off")
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    assert not firewall.is_active(ports)
    lib.flush_allowlist()


@pytest.mark.parametrize("port", lib.PORTS)
@pytest.mark.parametrize("protocol", lib.PROTOCOLS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_04_allowlist_port_and_protocol(port, protocol):
    protocol = str(protocol)
    lib.set_firewall("on")
    lib.set_ipv6("on")
    lib.add_port_and_protocol_to_allowlist(port, protocol)
    assert not firewall.is_active(port, protocol)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        print(output)
        assert lib.is_connect_successful(output)
        assert network.is_ipv4_and_ipv6_connected(20)

    assert firewall.is_active(port, protocol)

    lib.set_firewall("off")

    assert not firewall.is_active(port, protocol)

    with lib.ErrorDefer(lib.flush_allowlist):
        with lib.ErrorDefer(sh.nordvpn.disconnect):
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    lib.set_ipv6("off")
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    assert not firewall.is_active(port, protocol)
    lib.flush_allowlist()


@pytest.mark.parametrize("subnet_addr", lib.SUBNETS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_05_allowlist_subnet(subnet_addr):
    lib.set_firewall("on")
    lib.set_ipv6("on")
    lib.add_subnet_to_allowlist(subnet_addr)
    assert not firewall.is_active("", "", subnet_addr)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        print(output)
        assert lib.is_connect_successful(output)
        assert network.is_ipv4_and_ipv6_connected(20)

    assert firewall.is_active("", "", subnet_addr)

    lib.set_firewall("off")

    assert not firewall.is_active("", "", subnet_addr)

    with lib.ErrorDefer(lib.flush_allowlist):
        with lib.ErrorDefer(sh.nordvpn.disconnect):
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    lib.set_ipv6("off")
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    assert not firewall.is_active("", "", subnet_addr)
    lib.flush_allowlist()


def test_firewall_ipv6_06_with_killswitch():
    lib.set_firewall("on")
    lib.set_ipv6("on")
    assert not firewall.is_active()

    lib.set_killswitch("on")
    assert firewall.is_active()

    lib.set_killswitch("off")
    assert not firewall.is_active()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_07_with_killswitch_while_connected():
    lib.set_firewall("on")
    lib.set_ipv6("on")
    assert not firewall.is_active()

    lib.set_killswitch("on")
    assert firewall.is_active()

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        print(output)
        assert lib.is_connect_successful(output)
        assert network.is_ipv4_and_ipv6_connected(20)

    assert firewall.is_active()

    lib.set_killswitch("off")
    assert firewall.is_active()

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert network.is_ipv4_and_ipv6_connected(20)

    output = sh.nordvpn.disconnect()
    print(output)
    lib.set_ipv6("off")
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    assert not firewall.is_active()
