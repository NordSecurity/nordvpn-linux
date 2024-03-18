import json
import re
import socket
import time
from datetime import datetime

import pytest
import sh
import timeout_decorator

import lib
from lib import daemon, info, logging, login, meshnet, network, ssh

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

    ssh_client.download_file("/var/log/nordvpn/daemon.log", f"{logging.PATH}/other-peer-daemon.log")

    ssh_client.exec_command("nordvpn set defaults")
    sh.nordvpn.set.defaults()
    daemon.stop_peer(ssh_client)
    daemon.stop()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(60)
def test_route_to_peer_that_is_connected_to_vpn():
    peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())
    local_hostname = peer_list.get_this_device().hostname
    peer_hostname = peer_list.get_external_peer().hostname

    ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
    ssh_client_mesh.connect()

    ssh_client_mesh.exec_command("nordvpn connect")

    my_ip = network.get_external_device_ip()
    sh.nordvpn.mesh.peer.connect(peer_hostname)
    assert my_ip != network.get_external_device_ip()

    sh.nordvpn.disconnect()
    ssh_client_mesh.exec_command("nordvpn disconnect")

    time.sleep(3) # Other way around

    sh.nordvpn.connect()

    peer_ip = ssh_client_mesh.network.get_external_device_ip()
    ssh_client_mesh.exec_command(f"nordvpn mesh peer connect {local_hostname}")
    assert peer_ip != ssh_client_mesh.network.get_external_device_ip()

    ssh_client_mesh.exec_command("nordvpn disconnect")
    sh.nordvpn.disconnect()

    ssh_client_mesh.disconnect()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(60)
def test_route_to_peer_that_disconnects_from_vpn():
    peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())
    local_hostname = peer_list.get_this_device().hostname
    peer_hostname = peer_list.get_external_peer().hostname

    time.sleep(2) # Takes a second or two for hostname to be recognized by system
    ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
    ssh_client_mesh.connect()

    ssh_client_mesh.exec_command("nordvpn connect")

    my_ip = network.get_external_device_ip()
    sh.nordvpn.mesh.peer.connect(peer_hostname)
    assert my_ip != network.get_external_device_ip()

    ssh_client_mesh.exec_command("nordvpn disconnect")
    assert my_ip == network.get_external_device_ip()

    sh.nordvpn.disconnect()


    time.sleep(3) # Other way around

    sh.nordvpn.connect()

    peer_ip = ssh_client_mesh.network.get_external_device_ip()
    ssh_client_mesh.exec_command(f"nordvpn mesh peer connect {local_hostname}")
    assert peer_ip != ssh_client_mesh.network.get_external_device_ip()

    sh.nordvpn.disconnect()
    assert peer_ip == ssh_client_mesh.network.get_external_device_ip()

    ssh_client_mesh.exec_command("nordvpn disconnect")

    ssh_client_mesh.disconnect()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(60)
def test_route_to_peer_when_already_routing():
    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname

    ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
    ssh_client_mesh.connect()

    ssh_client_mesh.exec_command("nordvpn connect")

    my_ip = network.get_external_device_ip()
    sh.nordvpn.mesh.peer.connect(peer_hostname)
    assert my_ip != network.get_external_device_ip()

    sh.nordvpn.mesh.peer.connect(peer_hostname)
    assert my_ip != network.get_external_device_ip()

    sh.nordvpn.disconnect()
    ssh_client_mesh.exec_command("nordvpn disconnect")


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(130)
def test_route_traffic_to_each_other():
    peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())
    peer_hostname = peer_list.get_external_peer().hostname
    peer_pubkey = peer_list.get_external_peer().public_key

    time.sleep(2) # Takes a second or two for hostname to be recognized by system
    ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
    ssh_client_mesh.connect()

    sh.nordvpn.mesh.peer.connect(peer_pubkey)
    local_hostname = peer_list.get_this_device().hostname
    ssh_client_mesh.exec_command(f"nordvpn mesh peer connect {local_hostname}")

    assert network.is_not_available()
    assert ssh_client_mesh.network.is_not_available()

    sh.nordvpn.disconnect()
    ssh_client_mesh.exec_command("nordvpn disconnect")
    ssh_client_mesh.disconnect()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(70)
def test_routing_deny_for_peer_is_peer_no_netting():
    peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())
    peer_hostname = peer_list.get_external_peer().hostname
    time.sleep(2) # Takes a second or two for hostname to be recognized by system
    ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
    ssh_client_mesh.connect()

    local_ip = peer_list.get_this_device().ip
    ssh_client_mesh.exec_command("nordvpn mesh peer connect " + local_ip)

    sh.nordvpn.mesh.peer.routing.deny(peer_hostname)
    time.sleep(2)

    assert ssh_client_mesh.network.is_not_available()

    ssh_client_mesh.exec_command("nordvpn disconnect")
    ssh_client_mesh.disconnect()


def test_route_to_nonexistant_node():
    nonexistant_node_name = "penguins-are-cool.nord"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.mesh.peer.connect(nonexistant_node_name)

    expected_message = meshnet.MSG_PEER_UNKNOWN % nonexistant_node_name

    assert expected_message in str(ex)


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(90)
def test_route_to_peer_status_valid():
    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname
    time.sleep(2) # Takes a second or two for hostname to be recognized by system
    ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
    ssh_client_mesh.connect()

    peer_nick = "a-A-a"
    sh.nordvpn.mesh.peer.nick.set(peer_hostname, peer_nick)
    sh.nordvpn.mesh.peer.connect(peer_nick)

    connect_time = time.monotonic()

    time.sleep(15)
    sh.ping("-c", "1", "-w", "1", "103.86.96.100")

    status_time = time.monotonic()
    status_output = sh.nordvpn.status().lstrip('\r -')

    # Split the data into lines, filter out lines that don't contain ':',
    # split each line into key-value pairs, strip whitespace, and convert keys to lowercase
    status_info = {
        a.strip().lower(): b.strip()
        for a, b in (
            element.split(':')  # Split each line into key-value pair
            for element in filter(lambda line: len(line.split(':')) == 2, status_output.split('\n'))  # Filter lines containing ':'
        )
    }

    logging.log("status_info: " + str(status_info))
    logging.log("status_info: " + str(sh.nordvpn.status()))

    assert "Connected" in status_info['status']
    assert peer_hostname in status_info['hostname']
    assert socket.gethostbyname(peer_hostname) in status_info['ip']
    assert "NORDLYNX" in status_info['current technology']
    assert "UDP" in status_info['current protocol']

    transfer_data = status_info['transfer'].split(" ")
    transfer_received = float(transfer_data[0])
    transfer_sent = float(transfer_data[3])
    assert transfer_received >= 0
    assert transfer_sent > 0

    time_connected = int(status_info['uptime'].split(" ")[0])
    time_passed = status_time - connect_time
    if "minute" in status_info["uptime"]:
        time_connected_seconds = int(status_info['uptime'].split(" ")[2])
        assert time_connected * 60 + time_connected_seconds >= time_passed - 1 and time_connected * 60 + time_connected_seconds <= time_passed + 1
    else:
        assert time_connected >= time_passed - 1 and time_connected <= time_passed + 1

    sh.nordvpn.disconnect()
    ssh_client_mesh.disconnect()


@pytest.mark.skip(reason="Test suit exits, before test can be completed.")
def test_route_to_peer_that_is_disconnected():
    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname

    ssh_client.exec_command("nordvpn set mesh off")

    time.sleep(140)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.mesh.peer.connect(peer_hostname)

    expected_message = meshnet.MSG_PEER_OFFLINE % peer_hostname

    assert expected_message in str(ex)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES[:-1])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_route_traffic_to_peer_wrong_tech(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.mesh.peer.connect(peer_hostname)

    assert meshnet.MSG_ROUTING_NEED_NORDLYNX in str(ex)


@pytest.mark.skip(reason="LVPN-241")
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_route_to_peer_that_has_killswitch_on():
    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname
    time.sleep(2) # Takes a second or two for hostname to be recognized by system
    ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
    ssh_client_mesh.connect()

    ssh_client_mesh.exec_command("nordvpn set killswitch on")

    sh.nordvpn.mesh.peer.connect(peer_hostname)
    assert network.is_not_available()

    sh.nordvpn.disconnect()
    ssh_client_mesh.exec_command("nordvpn set killswitch off")
    ssh_client_mesh.disconnect()


def test_derp_server_selection_logic():
    def has_duplicates(list):
        return len(list) != len(set(list))

    ssh_client.exec_command("sudo iptables -I OUTPUT 1 -p tcp -m tcp --sport 8765 -j DROP")
    ssh_client.exec_command("sudo iptables -I OUTPUT 1 -p tcp -m tcp --dport 8765 -j DROP")

    daemon.stop_peer(ssh_client)
    ssh_client.exec_command("echo '' > /var/log/nordvpn/daemon.log")
    daemon.start_peer(ssh_client)

    time.sleep(30)

    dlog = ssh_client.exec_command("cat /var/log/nordvpn/daemon.log").split("\n")

    derp_lines_from_logs =  meshnet.get_lines_with_keywords(dlog, ["body", "relay", "connected"])

    server_list = []
    for line in derp_lines_from_logs:
        json_match = re.search(r'\{.*\}', line)

        if json_match:
            json_data = json.loads(json_match.group())
            hostname = json_data.get("body", {}).get("hostname", "")
            server_list.append(hostname)

    assert not has_duplicates(server_list)

    ssh_client.exec_command("sudo iptables -D OUTPUT -p tcp -m tcp --sport 8765 -j DROP")
    ssh_client.exec_command("sudo iptables -D OUTPUT -p tcp -m tcp --dport 8765 -j DROP")


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
