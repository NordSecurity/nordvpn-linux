import os

import pytest
import requests
import sh
import timeout_decorator

import lib
from lib import daemon, info, logging, login, meshnet, network, settings, ssh

ssh_client = ssh.Ssh("qa-peer", "root", "root")


def setup_module(module):  # noqa: ARG001
    ssh_client.connect()
    daemon.install_peer(ssh_client)

def teardown_module(module):  # noqa: ARG001
    daemon.uninstall_peer(ssh_client)
    ssh_client.disconnect()


def setup_function(function):  # noqa: ARG001
    logging.log()
    daemon.start()
    daemon.start_peer(ssh_client)
    login.login_as("default")
    login.login_as("qa-peer", ssh_client)  # TODO: same account is used for everybody, tests can't be run in parallel
    sh.nordvpn.set.meshnet.on()
    ssh_client.exec_command("nordvpn set mesh on")
    # Ensure clean starting state
    meshnet.remove_all_peers()
    meshnet.remove_all_peers_in_peer(ssh_client)
    meshnet.revoke_all_invites()
    meshnet.revoke_all_invites_in_peer(ssh_client)
    meshnet.add_peer(ssh_client)


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()
    ssh_client.exec_command("nordvpn set defaults")
    sh.nordvpn.set.defaults()
    daemon.stop_peer(ssh_client)
    daemon.stop()


def test_meshnet_connect():
    # Ideally peer update should happen through Notification Center, but that doesn't work often
    sh.nordvpn.meshnet.peer.refresh()
    peer = meshnet.parse_peer_list(sh.nordvpn.mesh.peer.list()).get_external_peer()
    assert meshnet.is_peer_reachable(ssh_client, peer)


def test_mesh_removed_machine_by_other():
    # find my token from cli
    mytoken = ""
    output = sh.nordvpn.token()
    for ln in output.splitlines():
        if "Token:" in ln:
            _, mytoken = ln.split(None, 2)

    myname = meshnet.parse_peer_list(sh.nordvpn.mesh.peer.list()).get_this_device().hostname
    # find my machineid from api
    mymachineid = ""
    headers = {
        'Accept': 'application/json',
        'Authorization': 'Bearer token:' + mytoken,
    }
    response = requests.get('https://api.nordvpn.com/v1/meshnet/machines', headers=headers, timeout=5)
    for itm in response.json():
        if str(itm['hostname']) in myname:
            mymachineid = itm['identifier']

    # remove myself using api call
    headers = {
        'Accept': 'application/json',
        'Authorization': 'Bearer token:' + mytoken,
    }
    requests.delete('https://api.nordvpn.com/v1/meshnet/machines/' + mymachineid, headers=headers, timeout=5)

    # machine not found error should be handled by disabling meshnet
    try:
        sh.nordvpn.mesh.peer.list()
    except Exception as e:  # noqa: BLE001
        assert "Meshnet is not enabled." in str(e)

    sh.nordvpn.set.meshnet.on()  # enable back on for other tests
    meshnet.add_peer(ssh_client)


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
# This doesn't directly test meshnet, but it uses it
def test_allowlist_incoming_connection():
    my_ip = ssh_client.exec_command("echo $SSH_CLIENT").split()[0]

    peer_hostname = meshnet.parse_peer_list(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname
    # Initiate ssh connection via mesh because we are going to lose the main connection
    ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
    ssh_client_mesh.connect()
    with lib.Defer(ssh_client_mesh.disconnect):
        ssh_client_mesh.exec_command("nordvpn set killswitch on")
        with lib.Defer(lambda: ssh_client_mesh.exec_command("nordvpn set killswitch off")):
            # We should not have direct connection anymore after connecting to VPN
            with pytest.raises(sh.ErrorReturnCode_1):
                assert "icmp_seq=" not in sh.ping("-c", "1", "qa-peer")

                ssh_client_mesh.exec_command(f"nordvpn allowlist add subnet {my_ip}/32")
                with lib.Defer(lambda: ssh_client_mesh.exec_command("nordvpn allowlist remove all")):
                    # Direct connection should work again after allowlisting
                    assert "icmp_seq=" in sh.ping("-c", "1", "qa-peer")


@pytest.mark.parametrize("routing", [True, False])
@pytest.mark.parametrize("local", [True, False])
@pytest.mark.parametrize("incoming", [True, False])
@pytest.mark.parametrize("fileshare", [True, False])
def test_exitnode_permissions(routing: bool, local: bool, incoming: bool, fileshare: bool):
    peer_ip = meshnet.parse_peer_list(sh.nordvpn.mesh.peer.list()).get_external_peer().ip
    meshnet.set_permissions(peer_ip, routing, local, incoming, fileshare)

    (result, message) = meshnet.validate_input_chain(peer_ip, routing, local, incoming, fileshare)
    assert result, message

    (result, message) = meshnet.validate_forward_chain(peer_ip, routing, local, incoming, fileshare)
    assert result, message

    rules = sh.sudo.iptables("-S", "POSTROUTING", "-t", "nat")

    if routing:
        assert f"-A POSTROUTING -s {peer_ip}/32 ! -d 100.64.0.0/10 -m comment --comment nordvpn -j MASQUERADE" in rules
    else:
        assert f"-A POSTROUTING -s {peer_ip}/32 ! -d 100.64.0.0/10 -m comment --comment nordvpn -j MASQUERADE" not in rules


@pytest.mark.parametrize("lan_discovery", [True, False])
@pytest.mark.parametrize("local", [True, False])
def test_lan_discovery_exitnode(lan_discovery: bool, local: bool):
    peer_ip = meshnet.parse_peer_list(sh.nordvpn.mesh.peer.list()).get_external_peer().ip
    meshnet.set_permissions(peer_ip, True, local, True, True)

    lan_discovery_value = "on" if lan_discovery else "off"
    sh.nordvpn.set("lan-discovery", lan_discovery_value, _ok_code=(0, 1))

    # If either LAN discovery or local(or both) is disabled, routing rule should bellow LAN blocking rules.
    def check_rules_routing() -> (bool, str):
        rules = sh.sudo.iptables("-S", "FORWARD")

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

    sh.nordvpn.connect()
    with lib.Defer(sh.nordvpn.disconnect):
        for (result, message) in lib.poll(check_rules_routing):  # noqa: B007
            if result:
                break
        assert result, message


@pytest.mark.parametrize("lan_discovery", [True, False])
@pytest.mark.parametrize("local", [True, False])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
def test_killswitch_exitnode(lan_discovery: bool, local: bool):
    output = f"{sh.nordvpn.mesh.peer.list(_tty_out=False)}"
    my_ip = meshnet.get_this_device_ipv4(output)
    peer_ip = meshnet.get_peers(output)[0]

    # Initiate ssh connection via mesh because we are going to lose the main connection
    ssh_client_mesh = ssh.Ssh(peer_ip, "root", "root")
    ssh_client_mesh.connect()

    try:
        ssh_client_mesh.exec_command(f"nordvpn mesh peer incoming allow {my_ip}")
    except RuntimeError as err:
        if "already allowed" not in err.args[0]:
            raise
    try:
        ssh_client_mesh.exec_command(f"nordvpn mesh peer routing allow {my_ip}")
    except RuntimeError as err:
        if "already allowed" not in err.args[0]:
            raise
    try:
        ssh_client_mesh.exec_command(f"nordvpn mesh peer local {'allow' if local else 'deny'} {my_ip}")
    except RuntimeError as err:
        if "already allowed" not in err.args[0]:
            raise
    try:
        ssh_client_mesh.exec_command(f"nordvpn set lan-discovery {'on' if lan_discovery else 'off'}")
    except RuntimeError as err:
        if "already set" not in err.args[0]:
            raise

    # Start disconnected from exitnode
    assert network.is_available()

    # Connect to exitnode
    sh.nordvpn.mesh.peer.connect(peer_ip)
    assert daemon.is_connected()
    assert network.is_available()

    # Enable killswitch on exitnode
    ssh_client_mesh.exec_command("nordvpn set killswitch enabled")
    assert daemon.is_connected()
    assert network.is_not_available()

    # Disconnect from exitnode
    sh.nordvpn.disconnect()
    assert not daemon.is_connected()
    assert network.is_available()

    # Connect to exitnode
    sh.nordvpn.mesh.peer.connect(peer_ip)
    assert daemon.is_connected()
    assert network.is_not_available()

    # Disable killswitch on exitnode
    ssh_client_mesh.exec_command("nordvpn set killswitch disabled")
    assert daemon.is_connected()
    assert network.is_available()

    # Disconnect from exitnode
    sh.nordvpn.disconnect()
    assert not daemon.is_connected()
    assert network.is_available()


@pytest.mark.skip(reason="Connecting qa-peer to VPN does not work - LVPN-4665")
@pytest.mark.parametrize("lan_discovery", [True, False])
@pytest.mark.parametrize("local", [True, False])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
def test_killswitch_exitnode_vpn(lan_discovery: bool, local: bool):
    output = f"{sh.nordvpn.mesh.peer.list(_tty_out=False)}"
    my_ip = meshnet.get_this_device_ipv4(output)
    peer_ip = meshnet.get_peers(output)[0]

    # Initiate ssh connection via mesh because we are going to lose the main connection
    ssh_client_mesh = ssh.Ssh(peer_ip, "root", "root")
    ssh_client_mesh.connect()

    try:
        ssh_client_mesh.exec_command(f"nordvpn mesh peer incoming allow {my_ip}")
    except RuntimeError as err:
        if "already allowed" not in err.args[0]:
            raise
    try:
        ssh_client_mesh.exec_command(f"nordvpn mesh peer routing allow {my_ip}")
    except RuntimeError as err:
        if "already allowed" not in err.args[0]:
            raise
    try:
        ssh_client_mesh.exec_command(f"nordvpn mesh peer local {'allow' if local else 'deny'} {my_ip}")
    except RuntimeError as err:
        if "already allowed" not in err.args[0]:
            raise
    try:
        ssh_client_mesh.exec_command(f"nordvpn set lan-discovery {'on' if lan_discovery else 'off'}")
    except RuntimeError as err:
        if "already set" not in err.args[0]:
            raise

    # Start disconnected from exitnode
    assert network.is_available()
    my_external_ip = network.get_external_device_ip()

    # Connect to exitnode
    sh.nordvpn.mesh.peer.connect(peer_ip)
    assert daemon.is_connected()
    assert network.is_available()
    peer_external_ip = network.get_external_device_ip()

    # Enable killswitch on exitnode
    ssh_client_mesh.exec_command("nordvpn set killswitch enabled")
    assert daemon.is_connected()
    assert network.is_not_available()

    # Exitnode connects to VPN
    ssh_client_mesh.exec_command("nordvpn connect")  # TODO: Connecting qa-peer to VPN does not work
    # The VPN connection has failed. Please check your internet connection and try connecting to the VPN again. If the issue persists, contact our customer support.
    assert daemon.is_connected()
    assert network.is_available()
    peer_vpn_ip = network.get_external_device_ip()
    assert peer_vpn_ip not in [my_ip, my_external_ip, peer_ip, peer_external_ip]

    # Exitnode disconnects from VPN
    ssh_client_mesh.exec_command("nordvpn disconnect")
    assert daemon.is_connected()
    assert network.is_not_available()

    # Disable killswitch on exitnode
    ssh_client_mesh.exec_command("nordvpn set killswitch disabled")
    assert daemon.is_connected()
    assert network.is_available()

    # Disconnect from exitnode
    sh.nordvpn.disconnect()
    assert not daemon.is_connected()
    assert network.is_available()


def test_connect_set_mesh_off():
    peer = meshnet.parse_peer_list(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname
    assert network.is_available()
    sh.nordvpn.mesh.peer.connect(peer)
    assert daemon.is_connected()
    assert network.is_available()
    sh.nordvpn.disconnect()
    assert not daemon.is_connected()
    assert network.is_available()
    sh.nordvpn.connect()
    assert daemon.is_connected()
    assert network.is_available()
    sh.nordvpn.set.mesh.off()
    assert daemon.is_connected()
    assert network.is_available()
    sh.nordvpn.disconnect()
    assert not daemon.is_connected()
    assert network.is_available()
    sh.nordvpn.set.mesh.on()
    assert not daemon.is_connected()
    assert network.is_available()


def test_remove_peer_firewall_update():
    peer_ip = meshnet.parse_peer_list(sh.nordvpn.mesh.peer.list()).get_external_peer().ip
    meshnet.set_permissions(peer_ip, True, True, True, True)

    sh.nordvpn.mesh.peer.remove(peer_ip)
    sh.nordvpn.mesh.peer.refresh()

    def all_peer_permissions_removed() -> (bool, str):
        rules = sh.sudo.iptables("-S")
        if peer_ip not in rules:
            return True, ""
        return False, f"Rules for peer were not removed from firewall\nPeer IP: {peer_ip}\nrules:\n{rules}"

    result, message = None, None
    for (result, message) in lib.poll(all_peer_permissions_removed):  # noqa: B007
        if result:
            break

    assert result, message


def test_account_switch():
    sh.nordvpn.logout("--persist-token")
    login.login_as("qa-peer")
    sh.nordvpn.set.mesh.on()  # expecting failure here


def test_invite_send():

    assert "Meshnet invitation to 'test@test.com' was sent." in meshnet.send_meshnet_invite("test@test.com")

    assert "test@test.com" in sh.nordvpn.meshnet.invite.list()


def test_invite_send_repeated():
    with lib.Defer(lambda: sh.nordvpn.meshnet.invite.revoke("test@test.com")):
        meshnet.send_meshnet_invite("test@test.com")

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            meshnet.send_meshnet_invite("test@test.com")

        assert "Meshnet invitation for 'test@test.com' already exists." in str(ex.value)


def test_invite_send_own_email():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        meshnet.send_meshnet_invite(os.environ.get("DEFAULT_LOGIN_USERNAME"))

    assert "Email should belong to a different user." in str(ex.value)


def test_invite_send_not_an_email():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        meshnet.send_meshnet_invite("test")

    assert "Invalid email 'test'." in str(ex.value)


@pytest.mark.skip(reason="A different error message is expected - LVPN-262")
def test_invite_send_long_email():
    # A long email address containing more than 256 characters is created
    email = "test" * 65 + "@test.com"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        meshnet.send_meshnet_invite(email)

    assert "It's not you, it's us. We're having trouble with our servers. If the issue persists, please contact our customer support." not in str(ex.value)


@pytest.mark.skip(reason="A different error message is expected - LVPN-262")
def test_invite_send_email_special_character():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        meshnet.send_meshnet_invite("\u2222@test.com")

    assert "It's not you, it's us. We're having trouble with our servers. If the issue persists, please contact our customer support." not in str(ex.value)


def test_invite_revoke():

    meshnet.send_meshnet_invite("test@test.com")

    assert "Meshnet invitation to 'test@test.com' was revoked." in sh.nordvpn.meshnet.invite.revoke("test@test.com")

    assert "test@test.com" not in sh.nordvpn.meshnet.invite.list()


def test_invite_revoke_repeated():
    with lib.Defer(lambda: sh.nordvpn.meshnet.invite.revoke("test@test.com")):
        meshnet.send_meshnet_invite("test@test.com")

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.revoke("test@test.com")

    assert "No invitation from 'test@test.com' was found." in str(ex.value)


def test_invite_revoke_non_existent():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.revoke("test@test.com")

    assert "No invitation from 'test@test.com' was found." in str(ex.value)


def test_invite_revoke_non_existent_long_email():
    email = "test" * 65 + "@test.com"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.revoke(email)

    assert f"No invitation from '{email}' was found." in str(ex.value)


def test_invite_revoke_non_existent_special_character():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.revoke("\u2222@test.com")

    assert "No invitation from '\u2222@test.com' was found." in str(ex.value)


def test_invite_deny():

    meshnet.remove_all_peers()
    meshnet.remove_all_peers_in_peer(ssh_client)
    meshnet.revoke_all_invites()
    meshnet.revoke_all_invites_in_peer(ssh_client)

    meshnet.send_meshnet_invite(os.environ.get("QA_PEER_USERNAME"))

    assert os.environ.get("DEFAULT_LOGIN_USERNAME") in ssh_client.exec_command("nordvpn meshnet invite list")
    assert f"Meshnet invitation from '{os.environ.get('DEFAULT_LOGIN_USERNAME')}' was denied." in meshnet.deny_meshnet_invite(ssh_client)


def test_invite_deny_non_existent():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.deny("test@test.com")

    assert "No invitation from 'test@test.com' was found." in str(ex.value)


def test_invite_deny_non_existent_long_email():
    email = "test" * 65 + "@test.com"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.deny(email)

    assert f"No invitation from '{email}' was found." in str(ex.value)


def test_invite_deny_non_existent_special_character():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.deny("\u2222@test.com")

    assert "No invitation from '\u2222@test.com' was found." in str(ex.value)


def test_invite_accept():

    meshnet.remove_all_peers()
    meshnet.remove_all_peers_in_peer(ssh_client)
    meshnet.revoke_all_invites()
    meshnet.revoke_all_invites_in_peer(ssh_client)

    meshnet.send_meshnet_invite(os.environ.get("QA_PEER_USERNAME"))

    assert os.environ.get("DEFAULT_LOGIN_USERNAME") in ssh_client.exec_command("nordvpn meshnet invite list")
    assert f"Meshnet invitation from '{os.environ.get('DEFAULT_LOGIN_USERNAME')}' was accepted." in meshnet.accept_meshnet_invite(ssh_client)


def test_invite_accept_non_existent():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.accept("test@test.com")

    assert "No invitation from 'test@test.com' was found." in str(ex.value)


def test_invite_accept_non_existent_long_email():
    email = "test" * 65 + "@test.com"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.accept(email)

    assert f"No invitation from '{email}' was found." in str(ex.value)


def test_invite_accept_non_existent_special_character():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.accept("\u2222@test.com")

    assert "No invitation from '\u2222@test.com' was found." in str(ex.value)


@pytest.mark.parametrize("meshnet_allias", meshnet.MESHNET_ALIAS)
def test_set_meshnet_on_when_logged_out(meshnet_allias):
    
    sh.nordvpn.logout("--persist-token")
    assert not settings.is_meshnet_enabled()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set(meshnet_allias, "on")

    assert "You are not logged in." in str(ex.value)


@pytest.mark.skip(reason="LVPN-4590")
@pytest.mark.parametrize("meshnet_allias", meshnet.MESHNET_ALIAS)
def test_set_meshnet_off_when_logged_out(meshnet_allias):
    
    sh.nordvpn.logout("--persist-token")
    assert not settings.is_meshnet_enabled()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set(meshnet_allias, "off")

    assert "You are not logged in." in str(ex.value)


@pytest.mark.parametrize("meshnet_allias", meshnet.MESHNET_ALIAS)
def test_set_meshnet_off_on(meshnet_allias):

    assert "Meshnet is set to 'disabled' successfully." in sh.nordvpn.set(meshnet_allias, "off")
    assert not settings.is_meshnet_enabled()

    assert "Meshnet is set to 'enabled' successfully." in sh.nordvpn.set(meshnet_allias, "on")
    assert settings.is_meshnet_enabled()


@pytest.mark.parametrize("meshnet_allias", meshnet.MESHNET_ALIAS)
def test_set_meshnet_on_repeated(meshnet_allias):

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set(meshnet_allias, "on")

    assert "Meshnet is already enabled." in str(ex.value)


@pytest.mark.parametrize("meshnet_allias", meshnet.MESHNET_ALIAS)
def test_set_meshnet_off_repeated(meshnet_allias):

    sh.nordvpn.set(meshnet_allias, "off")

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set(meshnet_allias, "off")

    assert "Meshnet is already disabled." in str(ex.value)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES) # Only using standard technologies here because of "LVPN-4601 - Enabling Auto-connect disables Obfuscation"
# This doesn't directly test meshnet, but it uses it
def test_set_defaults_when_logged_in_2nd_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    
    sh.nordvpn.set.fwmark("0xe2f2")
    sh.nordvpn.set.killswitch("on")
    sh.nordvpn.set.tpl("on")
    sh.nordvpn.set.autoconnect("on")
    sh.nordvpn.set("lan-discovery", "on")

    assert settings.is_meshnet_enabled()
    assert "0xe1f1" not in sh.nordvpn.settings()
    assert daemon.is_killswitch_on()
    assert settings.is_tpl_enabled()
    assert settings.is_autoconnect_enabled()
    assert settings.is_lan_discovery_enabled()
    
    if tech == "openvpn":
        assert not settings.is_obfuscated_enabled()

    assert "Settings were successfully restored to defaults." in sh.nordvpn.set.defaults()

    assert settings.app_has_defaults_settings()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
# This doesn't directly test meshnet, but it uses it
def test_set_defaults_when_logged_out_1st_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.fwmark("0xe2f2")
    sh.nordvpn.set.killswitch("on")
    sh.nordvpn.set("lan-discovery", "on")
    sh.nordvpn.set.analytics("off")
    sh.nordvpn.set.tpl("on")

    assert settings.is_meshnet_enabled()
    assert "0xe1f1" not in sh.nordvpn.settings()
    assert daemon.is_killswitch_on()
    assert settings.is_lan_discovery_enabled()
    assert not settings.are_analytics_enabled()
    assert settings.is_tpl_enabled()
    
    if obfuscated == "on":
        assert settings.is_obfuscated_enabled()
    else:
        assert not settings.is_obfuscated_enabled()

    sh.nordvpn.logout("--persist-token")

    assert "Settings were successfully restored to defaults." in sh.nordvpn.set.defaults()

    assert settings.app_has_defaults_settings()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
# This doesn't directly test meshnet, but it uses it
def test_set_defaults_when_connected_2nd_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    daemon.restart() # Temporary solution to avoid Firewall staying enabled in settings - LVPN-4121

    sh.nordvpn.set.firewall("off")
    sh.nordvpn.set.tpl("on")
    sh.nordvpn.set.ipv6("on")

    sh.nordvpn.connect()
    assert "Status: Connected" in sh.nordvpn.status()

    assert not settings.is_firewall_enabled()
    assert settings.is_meshnet_enabled()
    assert settings.is_tpl_enabled()
    assert settings.is_ipv6_enabled()
    
    if obfuscated == "on":
        assert settings.is_obfuscated_enabled()
    else:
        assert not settings.is_obfuscated_enabled()

    assert "Settings were successfully restored to defaults." in sh.nordvpn.set.defaults()

    assert "Status: Disconnected" in sh.nordvpn.status()

    assert settings.app_has_defaults_settings()
