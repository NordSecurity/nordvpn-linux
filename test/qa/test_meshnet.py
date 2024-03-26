import pytest
import requests
import sh
import timeout_decorator

import lib
from lib import daemon, login, meshnet, settings, ssh

ssh_client = ssh.Ssh("qa-peer", "root", "root")


def setup_module(module):  # noqa: ARG001
    meshnet.TestUtils.setup_module(ssh_client)


def teardown_module(module):  # noqa: ARG001
    meshnet.TestUtils.teardown_module(ssh_client)


def setup_function(function):  # noqa: ARG001
    meshnet.TestUtils.setup_function(ssh_client)


def teardown_function(function):  # noqa: ARG001
    meshnet.TestUtils.teardown_function(ssh_client)


def test_meshnet_connect():
    # Ideally peer update should happen through Notification Center, but that doesn't work often
    sh.nordvpn.meshnet.peer.refresh()
    peer = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer()
    assert meshnet.is_peer_reachable(ssh_client, peer)


def test_mesh_removed_machine_by_other():
    # find my token from cli
    mytoken = ""
    output = sh.nordvpn.token()
    for ln in output.splitlines():
        if "Token:" in ln:
            _, mytoken = ln.split(None, 2)

    myname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_this_device().hostname
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

    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname
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
    peer_ip = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().ip
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


def test_remove_peer_firewall_update():
    peer_ip = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().ip
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
