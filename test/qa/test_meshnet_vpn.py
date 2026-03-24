import time

import pytest

import lib
from lib import daemon, meshnet, network, settings, ssh
from lib.shell import sh_no_tty

ssh_client = ssh.Ssh("qa-peer", "root", "root")


def setup_module(module):  # noqa: ARG001
    meshnet.TestUtils.setup_module(ssh_client)


def teardown_module(module):  # noqa: ARG001
    meshnet.TestUtils.teardown_module(ssh_client)


def setup_function(function):  # noqa: ARG001
    meshnet.TestUtils.setup_function(ssh_client)


def teardown_function(function):  # noqa: ARG001
    meshnet.TestUtils.teardown_function(ssh_client)


@pytest.mark.xfail(condition=meshnet.is_meshnet_test_disabled_from_run(), reason="Run only in nightly")
@pytest.mark.parametrize("lan_discovery", [True, False])
@pytest.mark.parametrize("local", [True, False])
def test_lan_discovery_exitnode(lan_discovery: bool, local: bool):
    """Manual TCs: LVPN-1261, LVPN-1262"""

    peer_ip = meshnet.PeerList.from_str(sh_no_tty.nordvpn.mesh.peer.list()).get_external_peer().ip
    meshnet.set_permissions(peer_ip, True, local, True, True)

    lan_discovery_value = "on" if lan_discovery else "off"
    sh_no_tty.nordvpn.set("lan-discovery", lan_discovery_value, _ok_code=(0, 1))

    # If either LAN discovery or local(or both) is disabled, routing rule should bellow LAN blocking rules.
    def check_rules_routing() -> (bool, str):
        rules = sh_no_tty.sudo.iptables("-S", "FORWARD")

        routing_rule = f"-A FORWARD -s {peer_ip}/32 -m comment --comment nordvpn-exitnode-transient -j ACCEPT"
        routing_rule_idx = rules.find(routing_rule)
        if routing_rule_idx == -1:
            return False, f"Routing rule not found\nrules:\n{rules}"

        for lan in meshnet.LANS:
            lan_drop_rule = f"-A FORWARD -s 100.64.0.0/10 -d {lan} -m comment --comment nordvpn-exitnode-transient -j DROP"
            lan_drop_rule_idx = rules.find(lan_drop_rule)
            if lan_drop_rule_idx == -1:
                return False, f"LAN drop rule not found for subnet {lan}\nrules:\n{rules}"

            if local and lan_discovery:
                if lan_drop_rule_idx < routing_rule_idx:
                    return False, f"Routing rule was added after LAN block rule for subnet {lan}\nrules:\n{rules}"
            elif lan_drop_rule_idx > routing_rule_idx:
                return False, f"Routing rule was added before LAN block rule for subnet {lan}\nrules:\n{rules}"

        return True, ""

    sh_no_tty.nordvpn.connect()
    with lib.Defer(sh_no_tty.nordvpn.disconnect):
        for (result, message) in lib.poll(check_rules_routing):  # noqa: B007
            if result:
                break
        assert result, message


@pytest.mark.xfail(condition=meshnet.is_meshnet_test_disabled_from_run(), reason="Run only in nightly")
@pytest.mark.parametrize("lan_discovery", [True, False])
@pytest.mark.parametrize("local", [True, False])
def test_killswitch_exitnode_vpn(lan_discovery: bool, local: bool):
    my_ip = meshnet.PeerList.from_str(sh_no_tty.nordvpn.mesh.peer.list()).get_this_device().ip
    peer_ip = meshnet.PeerList.from_str(sh_no_tty.nordvpn.mesh.peer.list()).get_external_peer().ip

    try:
        ssh_client.exec_command(f"nordvpn mesh peer incoming allow {my_ip}")
    except RuntimeError as err:
        if "already allowed" not in err.args[0]:
            raise
    try:
        ssh_client.exec_command(f"nordvpn mesh peer routing allow {my_ip}")
    except RuntimeError as err:
        if "already allowed" not in err.args[0]:
            raise
    try:
        ssh_client.exec_command(f"nordvpn mesh peer local {'allow' if local else 'deny'} {my_ip}")
    except RuntimeError as err:
        if "already allowed" not in err.args[0]:
            raise
    try:
        ssh_client.exec_command(f"nordvpn set lan-discovery {'on' if lan_discovery else 'off'}")
    except RuntimeError as err:
        if "already set" not in err.args[0]:
            raise

    # Start disconnected from exitnode
    assert network.is_available(), "Network should be available before connecting to exitnode"
    my_external_ip = network.get_external_device_ip()

    # Connect to exitnode
    sh_no_tty.nordvpn.mesh.peer.connect(peer_ip)
    assert daemon.is_connected(), "Daemon should be connected to exitnode"
    assert network.is_available(), "Network should be available when connected to exitnode"
    peer_external_ip = network.get_external_device_ip()

    # Enable killswitch on exitnode
    ssh_client.exec_command("nordvpn set killswitch enabled")
    assert daemon.is_connected(), "Daemon should remain connected with killswitch enabled on exitnode"
    assert network.is_not_available(), "Network should not be available when killswitch is enabled on exitnode"

    # Exitnode connects to VPN
    ssh_client.exec_command("nordvpn connect")
    assert daemon.is_connected(), "Daemon should remain connected when exitnode connects to VPN"
    assert network.is_available(), "Network should be available when exitnode connects to VPN"
    peer_vpn_ip = network.get_external_device_ip()
    assert peer_vpn_ip not in [my_ip, my_external_ip, peer_ip, peer_external_ip], "Exitnode VPN IP should be different from all previous IPs"

    # Exitnode disconnects from VPN
    ssh_client.exec_command("nordvpn disconnect")
    assert daemon.is_connected(), "Daemon should remain connected when exitnode disconnects from VPN"
    assert network.is_not_available(), "Network should not be available when exitnode disconnects with killswitch"

    # Disable killswitch on exitnode
    ssh_client.exec_command("nordvpn set killswitch disabled")
    assert daemon.is_connected(), "Daemon should remain connected after disabling killswitch on exitnode"
    assert network.is_available(), "Network should be available after disabling killswitch on exitnode"

    # Disconnect from exitnode
    sh_no_tty.nordvpn.disconnect()
    assert not daemon.is_connected(), "Daemon should be disconnected after disconnect"
    assert network.is_available(), "Network should be available after disconnecting from exitnode"


@pytest.mark.xfail(condition=meshnet.is_meshnet_test_disabled_from_run(), reason="Run only in nightly")
def test_connect_set_mesh_off():
    peer = meshnet.PeerList.from_str(sh_no_tty.nordvpn.mesh.peer.list()).get_external_peer().hostname
    assert network.is_available(), "Network should be available before connecting"
    sh_no_tty.nordvpn.mesh.peer.connect(peer)
    assert daemon.is_connected(), "Daemon should be connected to peer"
    assert network.is_available(), "Network should be available when connected to peer"
    sh_no_tty.nordvpn.disconnect()
    assert not daemon.is_connected(), "Daemon should be disconnected after disconnect"
    assert network.is_available(), "Network should be available after disconnect"
    sh_no_tty.nordvpn.connect()
    assert daemon.is_connected(), "Daemon should be connected to VPN"
    assert network.is_available(), "Network should be available when connected to VPN"
    sh_no_tty.nordvpn.set.mesh.off()
    assert daemon.is_connected(), "Daemon should remain connected after turning off mesh"
    assert network.is_available(), "Network should be available after turning off mesh"
    sh_no_tty.nordvpn.disconnect()
    assert not daemon.is_connected(), "Daemon should be disconnected after disconnect"
    assert network.is_available(), "Network should be available after disconnect"
    sh_no_tty.nordvpn.set.mesh.on()
    assert not daemon.is_connected(), "Daemon should not be connected after turning on mesh"
    assert network.is_available(), "Network should be available after turning on mesh"


@pytest.mark.xfail(condition=meshnet.is_meshnet_test_disabled_from_run(), reason="Run only in nightly")
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
# This doesn't directly test meshnet, but it uses it
def test_set_defaults_when_connected_2nd_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    daemon.restart() # Temporary solution to avoid Firewall staying enabled in settings - LVPN-4121

    sh_no_tty.nordvpn.set.firewall("off")
    sh_no_tty.nordvpn.set.tpl("on")

    sh_no_tty.nordvpn.connect()
    assert "Status: Connected" in sh_no_tty.nordvpn.status(), "Status should show connected"

    assert not settings.is_firewall_enabled(), "Firewall should be disabled"
    assert settings.is_meshnet_enabled(), "Meshnet should be enabled"
    assert settings.is_tpl_enabled(), "TPL should be enabled"

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled(), "Obfuscation should be enabled when set to on"
    else:
        assert not settings.is_obfuscated_enabled(), "Obfuscation should be disabled when set to off"

    assert "Settings were successfully restored to defaults." in sh_no_tty.nordvpn.set.defaults("--logout"), "Settings restore should show success message"

    assert "Status: Disconnected" in sh_no_tty.nordvpn.status(), "Status should show disconnected after restore"

    assert settings.app_has_defaults_settings(), "App should have default settings after restore"


def test_route_to_peer_that_is_connected_to_vpn():
    """Manual TC: LVPN-430"""

    peer_list = meshnet.PeerList.from_str(sh_no_tty.nordvpn.mesh.peer.list())
    local_hostname = peer_list.get_this_device().hostname
    peer_hostname = peer_list.get_external_peer().hostname

    ssh_client.exec_command("nordvpn connect")

    my_ip = network.get_external_device_ip()
    output = sh_no_tty.nordvpn.mesh.peer.connect(peer_hostname)
    assert meshnet.is_connect_successful(output, peer_hostname), "Meshnet peer connect should be successful"
    assert my_ip != network.get_external_device_ip(), "IP should change when connected to peer"

    lib.is_disconnect_successful(sh_no_tty.nordvpn.disconnect())
    ssh_client.exec_command("nordvpn disconnect")

    time.sleep(1) # Other way around

    sh_no_tty.nordvpn.connect()

    peer_ip = ssh_client.network.get_external_device_ip()
    output = ssh_client.exec_command(f"nordvpn mesh peer connect {local_hostname}")
    assert meshnet.is_connect_successful(output, local_hostname), "Peer meshnet connect to local device should be successful"
    assert peer_ip != ssh_client.network.get_external_device_ip(), "Peer IP should change when connected to local device"

    ssh_client.exec_command("nordvpn disconnect")
    lib.is_disconnect_successful(sh_no_tty.nordvpn.disconnect())


@pytest.mark.xfail(condition=meshnet.is_meshnet_test_disabled_from_run(), reason="Run only in nightly")
def test_route_to_peer_that_disconnects_from_vpn():
    peer_list = meshnet.PeerList.from_str(sh_no_tty.nordvpn.mesh.peer.list())
    local_hostname = peer_list.get_this_device().hostname
    peer_hostname = peer_list.get_external_peer().hostname

    ssh_client.exec_command("nordvpn connect")

    my_ip = network.get_external_device_ip()
    output = sh_no_tty.nordvpn.mesh.peer.connect(peer_hostname)
    assert meshnet.is_connect_successful(output, peer_hostname), "Meshnet peer connect should be successful"
    assert my_ip != network.get_external_device_ip(), "IP should change when connected to peer"

    ssh_client.exec_command("nordvpn disconnect")
    assert my_ip == network.get_external_device_ip(), "IP should restore to original value after peer disconnect"

    lib.is_disconnect_successful(sh_no_tty.nordvpn.disconnect())


    time.sleep(1) # Other way around

    sh_no_tty.nordvpn.connect()

    peer_ip = ssh_client.network.get_external_device_ip()
    output = ssh_client.exec_command(f"nordvpn mesh peer connect {local_hostname}")
    assert meshnet.is_connect_successful(output, local_hostname), "Peer meshnet connect to local device should be successful"
    assert peer_ip != ssh_client.network.get_external_device_ip(), "Peer IP should change when connected to local device"

    sh_no_tty.nordvpn.disconnect()
    assert peer_ip == ssh_client.network.get_external_device_ip(), "Peer IP should restore to original value after disconnect"

    lib.is_disconnect_successful(ssh_client.exec_command("nordvpn disconnect"))


def test_route_traffic_to_peer_once_again_when_already_routing():
    peer_hostname = meshnet.PeerList.from_str(sh_no_tty.nordvpn.mesh.peer.list()).get_external_peer().hostname

    ssh_client.exec_command("nordvpn connect")

    my_ip = network.get_external_device_ip()
    output = sh_no_tty.nordvpn.mesh.peer.connect(peer_hostname)
    assert meshnet.is_connect_successful(output, peer_hostname), "First meshnet peer connect should be successful"
    assert network.is_connected(), "Network should be connected after peer routing"
    assert my_ip != network.get_external_device_ip(), "IP should change when first routing to peer"

    output = sh_no_tty.nordvpn.mesh.peer.connect(peer_hostname)
    assert meshnet.is_connect_successful(output, peer_hostname), "Second meshnet peer connect should be successful"
    assert network.is_connected(), "Network should still be connected after repeated peer routing"
    assert my_ip != network.get_external_device_ip(), "IP should remain changed when routing again to same peer"

    sh_no_tty.nordvpn.disconnect()
    ssh_client.exec_command("nordvpn disconnect")


