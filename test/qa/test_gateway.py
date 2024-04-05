import pytest
import sh
import timeout_decorator

import lib
from lib import (
    daemon,
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


# Test for 3.8.10 hotfix. Default gateway is not detected when there is not a physical interface
# Issue: 491
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_default_gateway_is_detected():
    # Create bridge interface
    sh.sudo.ip.link.add.br0.type.bridge()

    # Get IP address and interface name of your current default gateway
    output = sh.ip.route.show("default")
    _, _, ip_addr, _, iface = output.split(None, 5)
    logging.log(ip_addr)
    logging.log(iface)

    # Add IP address to a bridge interface, but make sure that the IP is in the same network as default gateway's IP
    sh.sudo.ip.addr.add.dev.br0(ip_addr)

    # Up the bridge interface
    sh.sudo.ip.link.set.br0.up()

    # Add current default gateway to bridge
    sh.sudo.ip.link.set(iface, "master", "br0")

    # Set bridge as default gateway
    sh.sudo.ip.route.change.default.dev.br0()

    output = sh.bridge.link()
    logging.log(output)

    # Commands to undo the previous work and return the routing tables to their original state
    # Remove the interface from the bridge
    remove_iface = sh.sudo.bake("ip", "link", "set", iface, "nomaster")
    # Down the interface
    down_iface = sh.sudo.bake("ip", "link", "set", iface, "down")
    # Delete the bridge
    remove_br = sh.sudo.bake("ip", "link", "delete", "br0", "type", "bridge")
    # Up the interface
    up_iface = sh.sudo.bake("ip", "link", "set", iface, "up")
    # Add the original default gateway
    add_dg = sh.sudo.bake("ip", "route", "add", "default", "via", ip_addr, "dev", iface)

    with lib.ErrorDefer(add_dg):
        with lib.ErrorDefer(up_iface):
            with lib.ErrorDefer(remove_br):
                with lib.ErrorDefer(down_iface):
                    with lib.ErrorDefer(remove_iface):
                        print(sh.ip.route())
                        # Connect to VPN
                        output = sh.nordvpn.connect(_tty_out=False)
                        print(output)
                        assert lib.is_connect_successful(output)
                        assert network.is_connected()

                        output = sh.nordvpn.disconnect()
                        print(output)
                        assert lib.is_disconnect_successful(output)
                        assert daemon.is_disconnected()

    remove_iface()
    down_iface()
    remove_br()
    up_iface()
    add_dg()
