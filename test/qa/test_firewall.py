import os

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
    logging.log("IP: " + str(network.get_external_device_ip()))
    pytest.skip()


def teardown_module(module):  # noqa: ARG001
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("port", lib.PORTS)
def test_firewall_02_allowlist_port(tech, proto, obfuscated, port):
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("ports", lib.PORTS_RANGE)
def test_firewall_03_allowlist_ports_range(tech, proto, obfuscated, ports):
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
    with lib.Defer(sh.nordvpn.set.killswitch.off):
        lib.set_technology_and_protocol(tech, proto, obfuscated)

        lib.set_firewall("on")
        assert not firewall.is_active()

        lib.set_killswitch("on")
        assert firewall.is_active()
    assert not firewall.is_active()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.parametrize("before_connect", [True, False])
def test_firewall_lan_discovery(tech, proto, obfuscated, before_connect):
    with lib.Defer(lambda: sh.nordvpn.set("lan-discovery", "off", _ok_code=(0, 1))):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            if before_connect:
                sh.nordvpn.set("lan-discovery", "on")

            sh.nordvpn.connect()

            if not before_connect:
                sh.nordvpn.set("lan-discovery", "on")

            rules = os.popen("sudo iptables -S INPUT").read()
            for rule in firewall.INPUT_LAN_DISCOVERY_RULES:
                assert rule in rules, f"{rule} input rule not found in iptables."

            rules = os.popen("sudo iptables -S FORWARD").read()
            for rule in firewall.FORWARD_LAN_DISCOVERY_RULES:
                assert rule in rules, f"{rule} input rule not found in iptables."

            rules = os.popen("sudo iptables -S OUTPUT").read()
            for rule in firewall.OUTPUT_LAN_DISCOVERY_RULES:
                assert rule in rules, f"{rule} output rule not found in iptables"

            sh.nordvpn.set("lan-discovery", "off")

            rules = os.popen("sudo iptables -S INPUT").read()
            for rule in firewall.INPUT_LAN_DISCOVERY_RULES:
                assert rule not in rules, f"{rule} input rule not found in iptables."

            rules = os.popen("sudo iptables -S FORWARD").read()
            for rule in firewall.FORWARD_LAN_DISCOVERY_RULES:
                assert rule not in rules, f"{rule} input rule not found in iptables."

            rules = os.popen("sudo iptables -S OUTPUT").read()
            for rule in firewall.OUTPUT_LAN_DISCOVERY_RULES:
                assert rule not in rules, f"{rule} output rule not found in iptables"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_firewall_lan_allowlist_interaction(tech, proto, obfuscated):
    with lib.Defer(lambda: sh.nordvpn.set("lan-discovery", "off", _ok_code=(0, 1))):
        with lib.Defer(sh.nordvpn.disconnect):
            lib.set_technology_and_protocol(tech, proto, obfuscated)

            sh.nordvpn.connect()

            subnet = "192.168.0.0/18"

            sh.nordvpn.allowlist.add.subnet(subnet)
            sh.nordvpn.set("lan-discovery", "on")

            rules = os.popen("sudo iptables -S INPUT").read()
            assert f"-A INPUT -s {subnet} -i eth0 -m comment --comment nordvpn -j ACCEPT" not in rules, "Whitelist rule was not removed from the INPUT chain when LAN discovery was enabled."

            rules = os.popen("sudo iptables -S FORWARD").read()
            assert f"-A FORWARD -d {subnet} -o eth0 -m comment --comment nordvpn -j ACCEPT" not in rules, "Whitelist rule was not removed from the FORWARD chain when LAN discovery was enabled."

            rules = os.popen("sudo iptables -S OUTPUT").read()
            assert f"-A OUTPUT -s {subnet} -o eth0 -m comment --comment nordvpn -j ACCEPT" not in rules, "Whitelist rule was not removed from the OUTPUT chain when LAN discovery was enabled."

            sh.nordvpn.set("lan-discovery", "off")

            rules = os.popen("sudo iptables -S INPUT").read()
            for rule in firewall.INPUT_LAN_DISCOVERY_RULES:
                assert rule not in rules, f"{rule} input rule not found in iptables."

            rules = os.popen("sudo iptables -S FORWARD").read()
            for rule in firewall.FORWARD_LAN_DISCOVERY_RULES:
                assert rule not in rules, f"{rule} input rule not found in iptables."

            rules = os.popen("sudo iptables -S OUTPUT").read()
            for rule in firewall.OUTPUT_LAN_DISCOVERY_RULES:
                assert rule not in rules, f"{rule} output rule not found in iptables"
