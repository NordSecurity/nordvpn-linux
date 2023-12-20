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


def teardown_module(module):
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connected_firewall_disable(tech, proto, obfuscated):
    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_ipv6("on")
        lib.set_firewall("on")
        assert not firewall.is_active()

        sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        assert network.is_ipv4_and_ipv6_connected(20)
        assert firewall.is_active()

        lib.set_firewall("off")
        assert not firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connected_firewall_enable(tech, proto, obfuscated):
    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_ipv6("on")
        lib.set_firewall("off")
        assert not firewall.is_active()

        sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        assert network.is_ipv4_and_ipv6_connected(20)
        assert not firewall.is_active()

        lib.set_firewall("on")
        assert firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_disable_connect(tech, proto, obfuscated):
    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_ipv6("on")
        lib.set_firewall("off")
        assert not firewall.is_active()

        sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        assert network.is_ipv4_and_ipv6_connected(20)
        assert not firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_enable_connect(tech, proto, obfuscated):
    with lib.Defer(sh.nordvpn.disconnect):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_ipv6("on")
        lib.set_firewall("on")
        assert not firewall.is_active()

        sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        assert network.is_ipv4_and_ipv6_connected(20)
        assert firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_02_allowlist_port(tech, proto, obfuscated, port):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.set_ipv6("on")
            lib.add_port_to_allowlist([port])
            assert not firewall.is_active([port])

            sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
            assert network.is_ipv4_and_ipv6_connected(20)
            assert firewall.is_active([port])

            lib.set_firewall("off")
            assert not firewall.is_active([port])
            assert network.is_ipv4_and_ipv6_connected(20)
        assert network.is_disconnected()
        assert not firewall.is_active([port])


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize("ports", lib.PORTS_RANGE)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_03_allowlist_ports_range(tech, proto, obfuscated, ports):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.set_ipv6("on")
            lib.add_ports_range_to_allowlist([ports])
            assert not firewall.is_active([ports])

            sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
            assert network.is_ipv4_and_ipv6_connected(20)
            assert firewall.is_active([ports])

            lib.set_firewall("off")
            assert not firewall.is_active([ports])
            assert network.is_ipv4_and_ipv6_connected(20)
        assert network.is_disconnected()
        assert not firewall.is_active([ports])


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_04_allowlist_port_and_protocol(tech, proto, obfuscated, port):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.set_ipv6("on")
            lib.add_port_to_allowlist([port])
            assert not firewall.is_active([port])

            sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
            assert network.is_ipv4_and_ipv6_connected(20)
            assert firewall.is_active([port])

            lib.set_firewall("off")
            assert not firewall.is_active([port])
            assert network.is_ipv4_and_ipv6_connected(20)
        assert network.is_disconnected()
        assert not firewall.is_active([port])


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_05_allowlist_subnet(tech, proto, obfuscated, subnet):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.set_ipv6("on")
            lib.add_subnet_to_allowlist([subnet])
            assert not firewall.is_active("", [subnet])

            sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
            assert network.is_ipv4_and_ipv6_connected(20)
            assert firewall.is_active("", [subnet])

            lib.set_firewall("off")
            assert not firewall.is_active("", [subnet])
            assert network.is_ipv4_and_ipv6_connected(20)
        assert network.is_disconnected()
        assert not firewall.is_active("", [subnet])


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_06_with_killswitch(tech, proto, obfuscated):
    with lib.Defer(lambda: lib.set_killswitch("off")):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        lib.set_ipv6("on")
        assert not firewall.is_active()

        lib.set_killswitch("on")
        assert firewall.is_active()
    assert not firewall.is_active()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_firewall_ipv6_07_with_killswitch_while_connected(tech, proto, obfuscated):
    with lib.Defer(sh.nordvpn.disconnect):
        with lib.Defer(lambda: lib.set_killswitch("off")):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.set_ipv6("on")
            assert not firewall.is_active()

            lib.set_killswitch("on")
            assert firewall.is_active()

            sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
            assert network.is_ipv4_and_ipv6_connected(20)
            assert firewall.is_active()
        assert firewall.is_active()
    assert network.is_disconnected()
    assert not firewall.is_active()
