import time

from lib import (
    daemon,
    info,
    login,
    logging,
    network,
)
import sh
import lib
import pytest
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
    lib.set_routing("on")


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_routing_on():
    subnet1 = "1.1.1.1"
    subnet2 = "2.2.2.2"
    subnet3 = "3.3.3.3"
    lib.add_subnet_to_allowlist(f"{subnet1}/32")
    lib.add_subnet_to_allowlist(f"{subnet2}/32")
    lib.add_subnet_to_allowlist(f"{subnet3}/32")
    table = 205

    with lib.ErrorDefer(sh.nordvpn.allowlist.remove.all):
        with lib.ErrorDefer(sh.nordvpn.disconnect):
            output = sh.nordvpn.connect()
            assert lib.is_connect_successful(output)

            rules = sh.ip.rule.show.table(table)
            assert "fwmark" in rules
            policyRoutes = sh.ip.route.show.table(table)
            assert subnet1 in policyRoutes
            assert subnet2 in policyRoutes
            assert subnet3 in policyRoutes
            assert "nordlynx" in policyRoutes


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_routing_off():
    subnet = "1.1.1.1"
    table = 205
    lib.add_subnet_to_allowlist(f"{subnet}/32")
    lib.set_routing("off")

    with lib.ErrorDefer(sh.nordvpn.allowlist.remove.all):
        with lib.ErrorDefer(sh.nordvpn.set.routing.on):
            with lib.ErrorDefer(sh.nordvpn.disconnect):
                print(sh.nordvpn.connect())

                rules = sh.ip.rule.show.table(table)
                assert not "fwmark" in rules
                routes = sh.ip.route()
                assert not subnet in routes
                policyRoutes = sh.ip.route.show.table(table)
                assert not "nordlynx" in policyRoutes

    print(sh.nordvpn.disconnect())
    lib.set_routing("on")
    lib.flush_allowlist()
    

@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)    
def test_toggle_routing_in_the_middle_of_the_connection():
    table = 205

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        print(sh.nordvpn.connect())

        routes = sh.ip.route.show.table(table)
        rules = sh.ip.rule()
        assert "nordlynx" in routes
        assert "mark" in rules
        assert network.is_available()
        
        with lib.ErrorDefer(sh.nordvpn.set.routing.on):
            lib.set_routing("off")
            routes = sh.ip.route.show.table(table)
            rules = sh.ip.rule()
            assert not "nordlynx" in routes
            assert not "mark" in rules
            assert network.is_not_available()
        
        lib.set_routing("on")
        routes = sh.ip.route.show.table(table)
        rules = sh.ip.rule()
        assert "nordlynx" in routes
        assert "mark" in rules
        assert network.is_available()

    print(sh.nordvpn.disconnect())


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_routing_when_iprule_already_exists(tech, proto, obfuscated):
    table = 205
    network_interface = "nordtun" if tech == "openvpn" else "nordlynx"
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        print(sh.nordvpn.connect())

        routes = sh.ip.route.show.table(table)
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

        with lib.ErrorDefer(sh.nordvpn.disconnect):
            print(sh.nordvpn.connect())

            routes = sh.ip.route.show.table(table)
            rules = sh.ip.rule()
            assert f"default dev {network_interface}" in routes
            assert "mark" in rules
            assert network.is_available()

            routes = sh.ip.route.show.table("main")
            assert f"default dev {network_interface}" not in routes

        print(sh.nordvpn.disconnect())