import time

from lib import (
    daemon,
    firewall,
    info,
    login,
    logging,
    network,
    settings
)
import sh
import lib
import pytest
import timeout_decorator


def setup_function(function):
    daemon.start()
    login.login_as("default")

    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()

    sh.nordvpn.logout("--persist-token")
    sh.nordvpn.set.defaults()
    daemon.stop()


SUBNET_1 = "2.2.2.2"
SUBNET_2 = "3.3.3.3"
SUBNET_3 = "4.4.4.4"

MSG_ROUTING_OFF =  "Routing is set to 'disabled' successfully."
MSG_ROUTING_ON = "Routing is set to 'enabled' successfully."
MSG_ROUTING_OFF_ALREADY =  "Routing is already set to 'disabled'."
MSG_ROUTING_ON_ALREADY = "Routing is already set to 'enabled'."
MSG_ROUTING_USED_BY_MESH = "Routing is currently used by Meshnet. Disable it first."


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_routing_enabled_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    lib.add_subnet_to_allowlist([f"{SUBNET_1}/32", f"{SUBNET_2}/32", f"{SUBNET_3}/32"])

    print(sh.nordvpn.connect())
    assert network.is_available()

    assert "fwmark" in sh.ip.rule.show.table(firewall.IP_ROUTE_TABLE)
    policy_routes = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    assert SUBNET_1 in policy_routes
    assert SUBNET_2 in policy_routes
    assert SUBNET_3 in policy_routes

    network_interface = "nordtun" if tech == "openvpn" else "nordlynx"
    assert network_interface in policy_routes

    assert settings.get_is_routing_enabled()


@pytest.mark.skip("LVPN-3273; LVPN-1574")
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_routing_disabled_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    lib.add_subnet_to_allowlist([f"{SUBNET_1}/32"])

    assert MSG_ROUTING_OFF in sh.nordvpn.set.routing.off()
    assert not settings.get_is_routing_enabled()

    print(sh.nordvpn.connect())

    assert network.is_not_available()

    assert not "fwmark" in sh.ip.rule.show.table(firewall.IP_ROUTE_TABLE)
    assert not SUBNET_1 in sh.ip.route()

    network_interface = "nordtun" if tech == "openvpn" else "nordlynx"
    assert not network_interface in sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)


    assert MSG_ROUTING_ON_ALREADY in sh.nordvpn.set.routing.on()
    assert settings.get_is_routing_enabled()


@pytest.mark.skip("LVPN-3273")
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connected_routing_disable_enable(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    network_interface = "nordtun" if tech == "openvpn" else "nordlynx"

    print(sh.nordvpn.connect())
    assert network.is_available()

    assert MSG_ROUTING_OFF in sh.nordvpn.set.routing.off()
    assert not settings.get_is_routing_enabled()
    assert not network_interface in sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    assert not "mark" in sh.ip.rule()
    assert network.is_not_available()

    assert MSG_ROUTING_ON in sh.nordvpn.set.routing.on()
    assert settings.get_is_routing_enabled()
    assert network_interface in sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    assert "mark" in sh.ip.rule()
    assert network.is_available()


@pytest.mark.skip("LVPN-3273; LVPN-1574")
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connected_routing_enable_disable(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    network_interface = "nordtun" if tech == "openvpn" else "nordlynx"

    assert MSG_ROUTING_OFF in sh.nordvpn.set.routing.off()
    assert not settings.get_is_routing_enabled()

    print(sh.nordvpn.connect())
    assert network.is_not_available()

    assert MSG_ROUTING_ON in sh.nordvpn.set.routing.on()
    assert settings.get_is_routing_enabled()
    assert network_interface in sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    assert "mark" in sh.ip.rule()
    assert network.is_available()

    assert MSG_ROUTING_OFF in sh.nordvpn.set.routing.off()
    assert not settings.get_is_routing_enabled()
    assert not network_interface in sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    assert not "mark" in sh.ip.rule()
    assert network.is_not_available()


@pytest.mark.skip("LVPN-4360")
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES_BASIC1)
def test_meshnet_on_routing_disable(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.mesh.on()
    assert MSG_ROUTING_USED_BY_MESH in sh.nordvpn.set.routing.off()
    assert settings.get_is_routing_enabled()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_routing_already_enabled(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    lib.set_routing("on")

    assert MSG_ROUTING_ON_ALREADY in sh.nordvpn.set.routing.on()
    assert settings.get_is_routing_enabled()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_routing_already_disabled(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    lib.set_routing("off")

    assert MSG_ROUTING_OFF_ALREADY in sh.nordvpn.set.routing.off()
    assert not settings.get_is_routing_enabled()


@pytest.mark.skip("LVPN-3273")
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_toggle_routing_in_the_middle_of_the_connection(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    network_interface = "nordtun" if tech == "openvpn" else "nordlynx"

    print(sh.nordvpn.connect())

    routes = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    rules = sh.ip.rule()
    assert network_interface in routes
    assert "mark" in rules
    assert network.is_available()

    lib.set_routing("off")
    routes = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    rules = sh.ip.rule()
    assert not network_interface in routes
    assert not "mark" in rules
    assert network.is_not_available()

    lib.set_routing("on")
    routes = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    rules = sh.ip.rule()
    assert network_interface in routes
    assert "mark" in rules
    assert network.is_available()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_routing_when_iprule_already_exists(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    network_interface = "nordtun" if tech == "openvpn" else "nordlynx"

    print(sh.nordvpn.connect())

    routes = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
    rules = sh.ip.rule()
    assert f"default dev {network_interface}" in routes
    assert "mark" in rules
    assert network.is_available()

    for line in rules:
        if "fwmark" in line:
            rule = line.split()[1:]

    print(sh.nordvpn.disconnect())
    daemon.stop()

    with lib.Defer(lambda: sh.sudo.ip.rule('del', *rule, _ok_code=(0, 2))):
        sh.sudo.ip.rule.add(*rule)
        daemon.start()
        time.sleep(5)

        print(sh.nordvpn.connect())

        routes = sh.ip.route.show.table(firewall.IP_ROUTE_TABLE)
        rules = sh.ip.rule()
        assert f"default dev {network_interface}" in routes
        assert "mark" in rules
        assert network.is_available()

        routes = sh.ip.route.show.table("main")
        assert f"default dev {network_interface}" not in routes

