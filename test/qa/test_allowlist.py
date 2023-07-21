import ipaddress
from lib import (
    daemon,
    info,
    logging,
    login,
    network,
)
import lib
import pytest
import sh
import socket
import timeout_decorator
import socket


def setup_module(module):
    daemon.start()
    login.login_as("default")
    # Add a random route and delete it to create routing table
    # Otherwise exceptions happen in tests
    cmd = sh.sudo.ip.route.add.default.via.bake("127.0.0.1")
    cmd.table("205")
    sh.sudo.ip.route.delete.default.table("205")


def teardown_module(module):
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


# Tests for 3.8.2 hotfix. Allowlist should not create routes.
# Issue 400
@pytest.mark.parametrize("subnet_addr", lib.SUBNETS)
def test_allowlist_does_not_create_new_routes_when_adding_deleting_subnets(subnet_addr):
    output_before_add = sh.ip.route.show.table(205)
    sh.nordvpn.allowlist.add.subnet(subnet_addr)
    output_after_add = sh.ip.route.show.table(205)
    sh.nordvpn.allowlist.remove.subnet(subnet_addr)
    output_after_delete = sh.ip.route.show.table(205)

    assert output_before_add == output_after_add
    assert output_after_add == output_after_delete


@pytest.mark.parametrize("port", lib.PORTS)
def test_allowlist_does_not_create_new_routes_when_adding_deleting_ports(port):
    output_before_add = sh.ip.route.show.table(205)
    sh.nordvpn.allowlist.add.port(port)
    output_after_add = sh.ip.route.show.table(205)
    sh.nordvpn.allowlist.remove.port(port)
    output_after_delete = sh.ip.route.show.table(205)

    assert output_before_add == output_after_add
    assert output_after_add == output_after_delete


def test_allowlist_is_not_set_when_disconnected():
    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        subnet = "1.1.1.0/24"
        assert subnet not in sh.ip.route.show.table(205)
        lib.add_subnet_to_allowlist(subnet)
        assert subnet not in sh.ip.route.show.table(205)

        port = 22
        assert f"port {port}" not in sh.sudo.iptables("-S")
        lib.add_port_to_allowlist(port)
        assert f"port {port}" not in sh.sudo.iptables("-S")


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_requires_connection():
    with lib.Defer(sh.nordvpn.allowlist.remove.all):
        subnet = "1.1.1.0/24"
        port = 22

        with lib.Defer(sh.nordvpn.disconnect):
            sh.nordvpn.connect()

            assert subnet not in sh.ip.route.show.table(205)
            lib.add_subnet_to_allowlist(subnet)
            assert subnet in sh.ip.route.show.table(205)

            assert f"port {port}" not in sh.sudo.iptables("-S")
            lib.add_port_to_allowlist(port)
            assert f"port {port}" in sh.sudo.iptables("-S")

        assert subnet not in sh.ip.route.show.table(205)
        assert f"port {port}" not in sh.sudo.iptables("-S")


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_allowlist_subnet():
    ip_provider = "icanhazip.com"
    my_ip = sh.curl(ip_provider)

    assert lib.is_connect_successful(sh.nordvpn.connect())
    with lib.Defer(sh.nordvpn.disconnect):
        my_vpn_ip = sh.curl(ip_provider)
        assert my_vpn_ip != my_ip

        _, _, ip_provider_addresses = socket.gethostbyname_ex(ip_provider)
        for ip in ip_provider_addresses:
            sh.nordvpn.allowlist.add.subnet(f"{ip}/32")
        with lib.Defer(sh.nordvpn.allowlist.remove.all):
            my_vpn_ip_after_allowlist = sh.curl(ip_provider)
            assert my_vpn_ip_after_allowlist == my_ip
            

