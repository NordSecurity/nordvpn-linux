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


# Tests for 3.8.2 hotfix. Whitelist should not create routes.
# Issue 400
@pytest.mark.parametrize("subnet_addr", lib.SUBNETS)
def test_whitelist_does_not_create_new_routes_when_adding_deleting_subnets(subnet_addr):
    try: # ip fails/panics if routing table was not created before
        output_before_add = sh.ip.route.show.table(205)
    except:
        output_before_add = ""
    sh.nordvpn.whitelist.add.subnet(subnet_addr)
    try:
        output_after_add = sh.ip.route.show.table(205)
    except:
        output_after_add = ""
    sh.nordvpn.whitelist.remove.subnet(subnet_addr)
    try:
        output_after_delete = sh.ip.route.show.table(205)
    except:
        output_after_delete = ""
    

    assert output_before_add == output_after_add
    assert output_after_add == output_after_delete


@pytest.mark.parametrize("port", lib.PORTS)
def test_whitelist_does_not_create_new_routes_when_adding_deleting_ports(port):
    try:
        output_before_add = sh.ip.route.show.table(205)
    except:
        output_before_add = ""
    sh.nordvpn.whitelist.add.port(port)
    try:
        output_after_add = sh.ip.route.show.table(205)
    except:
        output_after_add = ""
    sh.nordvpn.whitelist.remove.port(port)
    try:
        output_after_delete = sh.ip.route.show.table(205)
    except:
        output_after_delete = ""

    assert output_before_add == output_after_add
    assert output_after_add == output_after_delete


def test_whitelist_is_not_set_when_disconnected():
    with lib.Defer(sh.nordvpn.whitelist.remove.all):
        subnet = "1.1.1.0/24"
        try:
            output = sh.ip.route.show.table(205)
        except:
            output = ""
        assert subnet not in output
        lib.add_subnet_to_whitelist(subnet)
        try:
            output = sh.ip.route.show.table(205)
        except:
            output = ""
        assert subnet not in output

        port = 22
        assert f"port {port}" not in sh.sudo.iptables("-S")
        lib.add_port_to_whitelist(port)
        assert f"port {port}" not in sh.sudo.iptables("-S")


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
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


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_whitelist_subnet():
    ip_provider = "icanhazip.com"
    my_ip = sh.curl(ip_provider)

    assert lib.is_connect_successful(sh.nordvpn.connect())
    with lib.Defer(sh.nordvpn.disconnect):
        my_vpn_ip = sh.curl(ip_provider)
        assert my_vpn_ip != my_ip

        _, _, ip_provider_addresses = socket.gethostbyname_ex(ip_provider)
        for ip in ip_provider_addresses:
            sh.nordvpn.whitelist.add.subnet(f"{ip}/32")
        with lib.Defer(sh.nordvpn.whitelist.remove.all):
            my_vpn_ip_after_whitelist = sh.curl(ip_provider)
            assert my_vpn_ip_after_whitelist == my_ip
            

