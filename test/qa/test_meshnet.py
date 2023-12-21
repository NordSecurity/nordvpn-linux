import os

import pytest
import requests
import sh
import timeout_decorator

import lib
from lib import daemon, info, logging, login, meshnet, network, ssh

ssh_client = ssh.Ssh("qa-peer", "root", "root")


# noinspection PyUnusedLocal
def setup_module(module):
    ssh_client.connect()
    daemon.install_peer(ssh_client)
    daemon.start()
    daemon.start_peer(ssh_client)
    login.login_as("default")
    login.login_as("qa-peer", ssh_client)  # TODO: same account is used for everybody, tests can't be run in parallel
    lib.set_technology_and_protocol("nordlynx", "", "")
    sh.nordvpn.set.meshnet.on()
    ssh_client.exec_command("nordvpn set mesh on")
    # Ensure clean starting state
    meshnet.remove_all_peers()
    meshnet.remove_all_peers_in_peer(ssh_client)
    meshnet.revoke_all_invites()
    meshnet.revoke_all_invites_in_peer(ssh_client)
    meshnet.add_peer(ssh_client)


# noinspection PyUnusedLocal
def teardown_module(module):
    meshnet.revoke_all_invites()
    meshnet.remove_all_peers()
    ssh_client.exec_command("nordvpn set mesh off")
    sh.nordvpn.set.meshnet.off()
    ssh_client.exec_command("nordvpn logout --persist-token")
    sh.nordvpn.logout("--persist-token")
    daemon.stop_peer(ssh_client)
    daemon.stop()
    daemon.uninstall_peer(ssh_client)
    ssh_client.disconnect()


# noinspection PyUnusedLocal
def setup_function(function):
    logging.log()


# noinspection PyUnusedLocal
def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


def test_meshnet_connect():
    # Ideally peer update should happen through Notification Center, but that doesn't work often
    sh.nordvpn.meshnet.peer.refresh()
    assert meshnet.is_peer_reachable(ssh_client)


def test_mesh_removed_machine_by_other():
    # find my token from cli
    mytoken = ""
    output = sh.nordvpn.token()
    for ln in output.splitlines():
        if "Token:" in ln:
            _, mytoken = ln.split(None, 2)

    myname = meshnet.get_this_device(sh.nordvpn.mesh.peer.list())

    # find my machineid from api
    mymachineid = ""
    headers = {
        'Accept': 'application/json',
        'Authorization': 'Bearer token:' + mytoken,
    }
    response = requests.get('https://api.nordvpn.com/v1/meshnet/machines', headers=headers)
    for itm in response.json():
        if str(itm['hostname']) in myname:
            mymachineid = itm['identifier']

    # remove myself using api call
    headers = {
        'Accept': 'application/json',
        'Authorization': 'Bearer token:' + mytoken,
    }
    requests.delete('https://api.nordvpn.com/v1/meshnet/machines/' + mymachineid, headers=headers)

    # machine not found error should be handled by disabling meshnet
    try:
        sh.nordvpn.mesh.peer.list()
    except Exception as e:
        assert "Meshnet is not enabled." in str(e)

    sh.nordvpn.set.meshnet.on()  # enable back on for other tests
    meshnet.add_peer(ssh_client)


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
# This doesn't directly test meshnet, but it uses it
def test_allowlist_incoming_connection():
    my_ip = ssh_client.exec_command("echo $SSH_CLIENT").split()[0]

    peer_hostname = meshnet.get_this_device(ssh_client.exec_command("nordvpn mesh peer list"))
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


def validate_input_chain(peer_ip: str, routing: bool, local: bool, incoming: bool, fileshare: bool) -> (bool, str):
    rules = sh.sudo.iptables("-S", "INPUT")

    fileshare_rule = f"-A INPUT -s {peer_ip}/32 -p tcp -m tcp --dport 49111 -m comment --comment nordvpn -j ACCEPT"
    if (fileshare_rule in rules) != fileshare:
        return False, f"Fileshare permissions configured incorrectly, rule expected: {fileshare}\nrules:{rules}"

    incoming_rule = f"-A INPUT -s {peer_ip}/32 -m comment --comment nordvpn -j ACCEPT"
    if (incoming_rule in rules) != incoming:
        return False, f"Incoming permissions configured incorrectly, rule expected: {incoming}\nrules:{rules}"

    # If incoming is not enabled, no rules other than fileshare(if enabled) for that peer should be added
    if not incoming:
        if fileshare:
            rules = rules.replace(fileshare_rule, "")
        if peer_ip not in rules:
            return True, ""
        else:
            return False, f"Rules for peer({peer_ip}) found in the INCOMING chain but peer does not have the incoming permissions\nrules:\n{rules}"

    incoming_rule_idx = rules.find(incoming_rule)

    for lan in meshnet.LANS:
        lan_rule = f"-A INPUT -s {peer_ip}/32 -d {lan} -m comment --comment nordvpn -j DROP"
        lan_rule_idx = rules.find(lan_rule)
        if (routing and local) and lan_rule_idx != -1:
            return False, f"LAN/Routing permissions configured incorrectly\nlocal enabled: {local}\nrouting enabled: {routing}\nrules:\n{rules}"
        # verify that lan_rule is located above the local rule
        if lan_rule_idx > incoming_rule_idx:
            return False, f"LAN/Routing rules ineffective(added after incoming traffic rule)\nlocal enabled: {local}\nrouting enabled: {routing}\nrules:\n{rules}"

    return True, ""


def validate_forward_chain(peer_ip: str, routing: bool, local: bool, incoming: bool, fileshare: bool) -> (bool, str):
    _, _ = incoming, fileshare
    rules = sh.sudo.iptables("-S", "FORWARD")

    # This rule is added above the LAN denial rules if both local and routing is allowed to peer, or bellow LAN denial
    # if only routing is allowed.
    routing_enabled_rule = f"-A FORWARD -s {peer_ip}/32 -m comment --comment nordvpn-exitnode-transient -j ACCEPT"
    routing_enabled_rule_index = rules.find(routing_enabled_rule)

    if routing and (routing_enabled_rule_index == -1):
        return False, f"Routing permission not found\nrules:{rules}"
    if not routing and (routing_enabled_rule_index != -1):
        return False, f"Routing permission found\nrules:{rules}"

    for lan in meshnet.LANS:
        lan_drop_rule = f"-A FORWARD -s 100.64.0.0/10 -d {lan} -m comment --comment nordvpn-exitnode-transient -j DROP"
        lan_drop_rule_index = rules.find(lan_drop_rule)

        # If any peer has routing or local permission, lan block rules should be added, otherwise no rules should be added.
        if (routing or local) and lan_drop_rule_index == -1:
            return False, f"LAN drop rule not added for subnet {lan}\nrules:\n{rules}"
        elif (not routing) and (not lan) and lan_drop_rule_index != -1:
            return False, f"LAN drop rule added for subnet {lan}\nrules:\n{rules}"

        if routing:
            # Local is allowed, routing rule should be above LAN block rules to allow peer to access any subnet.
            if local and (lan_drop_rule_index < routing_enabled_rule_index):
                return False, f"LAN drop rule for subnet {lan} added before routing\nrules: {rules}"
            # Local is not allowed, routing rule should be below LAN block rules to deny peer access to local subnets.
            if (not local) and (lan_drop_rule_index > routing_enabled_rule_index):
                return False, f"LAN drop rule for subnet {lan} added after routing\nrules: {rules}"
            continue

        # If routing is not enabled, but lan is enabled, there should be one rule for each local network for the peer.
        # They should be located above the LAN block rules.
        if not local:
            continue

        lan_allow_rule = f"-A FORWARD -s {peer_ip}/32 -d {lan} -m comment --comment nordvpn-exitnode-transient -j ACCEPT"
        lan_allow_rule_index = rules.find(lan_allow_rule)

        if lan_allow_rule not in rules:
            return False, f"LAN allow rule for subnet {lan} not found\nrules:\n{rules}"

        if lan_allow_rule_index > lan_drop_rule_index:
            return False, f"LAN allow rule is added after LAN drop rule\nrules:\n{rules}"

    return True, ""


def set_permissions(peer: str, routing: bool, local: bool, incoming: bool, fileshare: bool):
    def bool_to_permission(permission: bool) -> str:
        if permission:
            return "allow"
        return "deny"

    # ignore any failures that might occur when permissions are already configured to the desired value
    sh.nordvpn.mesh.peer.routing(bool_to_permission(routing), peer, _ok_code=(0, 1))
    sh.nordvpn.mesh.peer.local(bool_to_permission(local), peer, _ok_code=(0, 1))
    sh.nordvpn.mesh.peer.incoming(bool_to_permission(incoming), peer, _ok_code=(0, 1))
    sh.nordvpn.mesh.peer.fileshare(bool_to_permission(fileshare), peer, _ok_code=(0, 1))


@pytest.mark.parametrize("routing", [True, False])
@pytest.mark.parametrize("local", [True, False])
@pytest.mark.parametrize("incoming", [True, False])
@pytest.mark.parametrize("fileshare", [True, False])
def test_exitnode_permissions(routing: bool,
                              local: bool,
                              incoming: bool,
                              fileshare: bool):
    peer_ip = meshnet.get_this_device_ipv4(ssh_client.exec_command("nordvpn mesh peer list"))
    set_permissions(peer_ip, routing, local, incoming, fileshare)

    (result, message) = validate_input_chain(peer_ip, routing, local, incoming, fileshare)
    assert result, message

    (result, message) = validate_forward_chain(peer_ip, routing, local, incoming, fileshare)
    assert result, message

    rules = sh.sudo.iptables("-S", "POSTROUTING", "-t", "nat")

    if routing:
        assert f"-A POSTROUTING -s {peer_ip}/32 -o eth0 -m comment --comment nordvpn -j MASQUERADE" in rules
    else:
        assert f"-A POSTROUTING -s {peer_ip}/32 -o eth0 -m comment --comment nordvpn -j MASQUERADE" not in rules


@pytest.mark.parametrize("lan_discovery", [True, False])
@pytest.mark.parametrize("local", [True, False])
def test_lan_discovery_exitnode(lan_discovery: bool, local: bool):
    peer_ip = meshnet.get_this_device_ipv4(ssh_client.exec_command("nordvpn mesh peer list"))
    set_permissions(peer_ip, True, local, True, True)

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
            else:
                if lan_drop_rule_idx > routing_rule_idx:
                    return False, f"Routing rule was added before LAN block rule for subnet {lan}\nrules:\n{rules}"

        return True, ""

    sh.nordvpn.connect()
    with lib.Defer(sh.nordvpn.disconnect):
        for (result, message) in lib.poll(check_rules_routing):
            if result:
                break
        assert result, message


def test_connect_set_mesh_off():
    output = f"{sh.nordvpn.mesh.peer.list(_tty_out=False)}"
    peer = meshnet.get_peers(output)[0]
    sh.nordvpn.mesh.peer.connect(peer)
    sh.nordvpn.disconnect()
    sh.nordvpn.connect()
    sh.nordvpn.set.mesh.off()

    with lib.Defer(sh.nordvpn.set.mesh.on):
        with lib.Defer(sh.nordvpn.disconnect):
            assert network.is_connected()


def test_remove_peer_firewall_update():
    peer_ip = meshnet.get_this_device_ipv4(ssh_client.exec_command("nordvpn mesh peer list"))
    set_permissions(peer_ip, True, True, True, True)

    sh.nordvpn.mesh.peer.remove(peer_ip)
    sh.nordvpn.mesh.peer.refresh()

    def all_peer_permissions_removed() -> (bool, str):
        rules = sh.sudo.iptables("-S")
        if peer_ip not in rules:
            return True, ""
        return False, f"Rules for peer were not removed from firewall\nPeer IP: {peer_ip}\nrules:\n{rules}"

    result, message = None, None
    for (result, message) in lib.poll(all_peer_permissions_removed):
        if result:
            break

    assert result, message


def test_account_switch():
    sh.nordvpn.logout("--persist-token")
    login.login_as("qa-peer")
    sh.nordvpn.set.mesh.on()  # expecting failure here

    # Recover starting state (this is the simplest way)
    teardown_module(None)
    setup_module(None)


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
