import random

import pytest
import sh

import lib
from lib import (
    allowlist,
    daemon,
    firewall,
    info,
    logging,
    login,
    network,
)


def setup_module(module):  # noqa: ARG001
    daemon.start()
    login.login_as("default")


def teardown_module(module):  # noqa: ARG001
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()


@pytest.mark.xfail(reason="LVPN-8096")
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS)
def test_firewall_ipv6_02_allowlist_port(tech, proto, obfuscated, port):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.set_ipv6("on")
            allowlist.add_ports_to_allowlist([port])
            assert not firewall.is_active([port])

            sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
            assert network.is_ipv4_and_ipv6_connected(20)
            assert firewall.is_active([port])

            lib.set_firewall("off")
            assert not firewall.is_active([port])
            assert network.is_ipv4_and_ipv6_connected(20)
        assert network.is_disconnected()
        assert not firewall.is_active([port])


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize("ports", lib.PORTS_RANGE)
def test_firewall_ipv6_03_allowlist_ports_range(tech, proto, obfuscated, ports):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.set_ipv6("on")
            allowlist.add_ports_to_allowlist([ports])
            assert not firewall.is_active([ports])

            sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
            assert network.is_ipv4_and_ipv6_connected(20)
            assert firewall.is_active([ports])

            lib.set_firewall("off")
            assert not firewall.is_active([ports])
            assert network.is_ipv4_and_ipv6_connected(20)
        assert network.is_disconnected()
        assert not firewall.is_active([ports])


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS)
def test_firewall_ipv6_04_allowlist_port_and_protocol(tech, proto, obfuscated, port):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.set_ipv6("on")
            allowlist.add_ports_to_allowlist([port])
            assert not firewall.is_active([port])

            sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
            assert network.is_ipv4_and_ipv6_connected(20)
            assert firewall.is_active([port])

            lib.set_firewall("off")
            assert not firewall.is_active([port])
            assert network.is_ipv4_and_ipv6_connected(20)
        assert network.is_disconnected()
        assert not firewall.is_active([port])


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.parametrize("subnet", lib.SUBNETS)
def test_firewall_ipv6_05_allowlist_subnet(tech, proto, obfuscated, subnet):
    with lib.Defer(lib.flush_allowlist):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            lib.set_firewall("on")
            lib.set_ipv6("on")
            allowlist.add_subnet_to_allowlist([subnet])
            assert not firewall.is_active(None, [subnet])

            sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
            assert network.is_ipv4_and_ipv6_connected(20)
            assert firewall.is_active(None, [subnet])

            lib.set_firewall("off")
            assert not firewall.is_active(None, [subnet])
            assert network.is_ipv4_and_ipv6_connected(20)
        assert network.is_disconnected()
        assert not firewall.is_active(None, [subnet])


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
def test_firewall_ipv6_06_with_killswitch(tech, proto, obfuscated):
    with lib.Defer(lambda: lib.set_killswitch("off")):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        lib.set_ipv6("on")
        assert not firewall.is_active()

        lib.set_killswitch("on")
        assert firewall.is_active()
    assert not firewall.is_active()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
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
