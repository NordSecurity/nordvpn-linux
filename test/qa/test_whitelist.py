from lib import (
    daemon,
    info,
    logging,
    login,
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


# Tests for 3.8.2 hotfix. Whitelist should not create routes.
# Issue 400
@pytest.mark.parametrize("subnet_addr", lib.SUBNETS)
def test_whitelist_does_not_create_new_routes_when_adding_deleting_subnets(subnet_addr):
    output_before_add = sh.ip.route.show.table(205)
    sh.nordvpn.whitelist.add.subnet(subnet_addr)
    output_after_add = sh.ip.route.show.table(205)
    sh.nordvpn.whitelist.remove.subnet(subnet_addr)
    output_after_delete = sh.ip.route.show.table(205)

    assert output_before_add == output_after_add
    assert output_after_add == output_after_delete


@pytest.mark.parametrize("port", lib.PORTS)
def test_whitelist_does_not_create_new_routes_when_adding_deleting_ports(port):
    output_before_add = sh.ip.route.show.table(205)
    sh.nordvpn.whitelist.add.port(port)
    output_after_add = sh.ip.route.show.table(205)
    sh.nordvpn.whitelist.remove.port(port)
    output_after_delete = sh.ip.route.show.table(205)

    assert output_before_add == output_after_add
    assert output_after_add == output_after_delete


def test_whitelist_is_not_set_when_disconnected():
    with lib.Defer(sh.nordvpn.whitelist.remove.all):
        subnet = "1.1.1.0/24"
        assert subnet not in sh.ip.route.show.table(205)
        lib.add_subnet_to_whitelist(subnet)
        assert subnet not in sh.ip.route.show.table(205)

        port = 22
        assert f"port {port}" not in sh.sudo.iptables("-S")
        lib.add_port_to_whitelist(port)
        assert f"port {port}" not in sh.sudo.iptables("-S")


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(20)
def test_whitelist_requires_connection():
    with lib.Defer(sh.nordvpn.whitelist.remove.all):
        subnet = "1.1.1.0/24"
        port = 22

        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()

            assert subnet not in sh.ip.route.show.table(205)
            lib.add_subnet_to_whitelist(subnet)
            assert subnet in sh.ip.route.show.table(205)

            assert f"port {port}" not in sh.sudo.iptables("-S")
            lib.add_port_to_whitelist(port)
            assert f"port {port}" in sh.sudo.iptables("-S")

        assert subnet not in sh.ip.route.show.table(205)
        assert f"port {port}" not in sh.sudo.iptables("-S")
