import json
import re
import time
from datetime import datetime

import pytest
import requests
import sh
import timeout_decorator

import lib
from lib import daemon, logging, login, meshnet, settings, ssh

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
    peer = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer()
    this_device = meshnet.PeerList.from_str(ssh_client.exec_command("nordvpn mesh peer list")).get_external_peer()

    nickname = "remote-machine"
    sh.nordvpn.mesh.peer.nick.set(peer.hostname, nickname)

    peer = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer() # Refresh nickname

    assert meshnet.is_peer_reachable(ssh_client, peer, meshnet.PeerName.Hostname)
    assert meshnet.is_peer_reachable(ssh_client, peer, meshnet.PeerName.Ip)
    assert meshnet.is_peer_reachable(ssh_client, peer, meshnet.PeerName.Nickname)
    assert nickname == peer.nickname

    nickname = "local-machine"
    ssh_client.exec_command(f"nordvpn mesh peer nick set {this_device.hostname} {nickname}")

    this_device = meshnet.PeerList.from_str(ssh_client.exec_command("nordvpn mesh peer list")).get_external_peer() # Refresh nickname

    assert ssh_client.network.ping(this_device.hostname)
    assert ssh_client.network.ping(this_device.ip)
    assert ssh_client.network.ping(this_device.nickname)
    assert nickname == this_device.nickname


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


@pytest.mark.parametrize(("permission", "permission_state", "expected_message"), meshnet.PERMISSION_SUCCESS_MESSAGE_PARAMETER_SET, \
                         ids=[f"{line[0]}-{line[1]}" for line in meshnet.PERMISSION_SUCCESS_MESSAGE_PARAMETER_SET])
@timeout_decorator.timeout(25)
def test_permission_messages_success(permission, permission_state, expected_message):
    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname

    reverse_permission_value = "allow" if permission_state == "deny" else "deny"
    meshnet.set_permission(peer_hostname, permission, reverse_permission_value)

    got_message = sh.nordvpn.mesh.peer(permission, permission_state, peer_hostname)

    expected_message = expected_message % peer_hostname

    assert expected_message in got_message


@pytest.mark.parametrize(("permission", "permission_state", "expected_message"), meshnet.PERMISSION_ERROR_MESSAGE_PARAMETER_SET, \
                         ids=[f"{line[0]}-{line[1]}" for line in meshnet.PERMISSION_ERROR_MESSAGE_PARAMETER_SET])
@timeout_decorator.timeout(25)
def test_permission_messages_error(permission, permission_state, expected_message):
    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname

    sh.nordvpn.mesh.peer(permission, permission_state, peer_hostname, _ok_code=(0, 1))

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        print(sh.nordvpn.mesh.peer(permission, permission_state, peer_hostname))

    expected_message = expected_message % peer_hostname

    assert expected_message in str(ex)


def test_derp_server_selection_logic():
    def has_duplicates(list):
        return len(list) != len(set(list))

    ssh_client.exec_command("sudo iptables -I OUTPUT 1 -p tcp -m tcp --sport 8765 -j DROP")
    ssh_client.exec_command("sudo iptables -I OUTPUT 1 -p tcp -m tcp --dport 8765 -j DROP")

    daemon.stop_peer(ssh_client)
    ssh_client.exec_command("echo '' > /var/log/nordvpn/daemon.log")
    daemon.start_peer(ssh_client)

    derp_lines_from_logs = []

    while len(derp_lines_from_logs) < 2:
        daemonlog = ssh_client.exec_command("cat /var/log/nordvpn/daemon.log").split("\n")
        derp_lines_from_logs = meshnet.get_lines_with_keywords(daemonlog, ["region_code", "connecting"])
        time.sleep(5)

    server_list = []
    for line in derp_lines_from_logs:
        json_match = re.search(r'\{.*\}', line)

        if json_match:
            json_data = json.loads(json_match.group())
            hostname = json_data.get("body", {}).get("hostname", "")
            server_list.append(hostname)

    # Same server should not be contacted twice in a row
    assert len(server_list) != 0
    assert not has_duplicates(server_list)

    ssh_client.exec_command("sudo iptables -D OUTPUT -p tcp -m tcp --sport 8765 -j DROP")
    ssh_client.exec_command("sudo iptables -D OUTPUT -p tcp -m tcp --dport 8765 -j DROP")


@pytest.mark.skip("LVPN-3428, need a discussion here")
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(80)
def test_direct_connection_rtt_and_loss():
    def get_loss(ping_output: str) -> float:
        """ pass `ping_output`, and get loss returned. """
        return float(ping_output.split("\n")[-3].split(", ")[2].split("%")[0])

    def get_average_rtt(ping_output: str) -> float:
        """ pass `ping_output`, and get average rtt returned. """
        return float(ping_output.split("\n")[-2].split("/")[4])

    def base_test(log: str, peer_hostname: str):
        log = log.split("\n")
        log_relay_events = meshnet.get_lines_with_keywords(log, [peer_hostname, "body", "relay", "connected"])
        log_direct_events = meshnet.get_lines_with_keywords(log, [peer_hostname, "body", "direct", "connected"])

        # Sometimes no relay lines show up, but direct ones do instead. If that happens, direct formed, so we can continue
        if len(log_relay_events) != 0:
            log_relay_event_time = datetime.strptime(log_relay_events[0].split(" ")[1], "%H:%M:%S")
            log_direct_event_time = datetime.strptime(log_direct_events[0].split(" ")[1], "%H:%M:%S")

            assert (log_direct_event_time - log_relay_event_time).total_seconds() < meshnet.TELIO_EXPECTED_RELAY_TO_DIRECT_TIME
        elif len(log_relay_events) == 0 and len(log_direct_events) > 0:
            pass

        # RTT & loss
        ping_output = sh.ping("-c", "20", peer_hostname)
        assert get_average_rtt(ping_output) <= meshnet.TELIO_EXPECTED_RTT
        assert get_loss(ping_output) <= meshnet.TELIO_EXPECTED_PACKET_LOSS

    peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())

    with open(logging.FILE) as tester_log:
        log_content = tester_log.read()
        qapeer_hostname = peer_list.get_external_peer().hostname
        base_test(log_content, qapeer_hostname)


def test_incoming_connections():
    peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())
    local_hostname = peer_list.get_this_device().hostname
    peer_hostname = peer_list.get_external_peer().hostname

    sh.nordvpn.mesh.peer.incoming.deny(peer_hostname)
    assert not ssh_client.network.ping(local_hostname, retry=1)

    ssh_client.exec_command(f"nordvpn mesh peer incoming deny {local_hostname}")
    assert not meshnet.is_peer_reachable(ssh_client, peer_list.get_external_peer(), retry=1)
