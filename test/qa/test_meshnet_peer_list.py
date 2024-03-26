import os

import pytest
import sh

from lib import daemon, info, logging, login, meshnet, ssh

ssh_client = ssh.Ssh("qa-peer", "root", "root")


def setup_module(module):  # noqa: ARG001
    os.makedirs("/home/qa/.config/nordvpn", exist_ok=True)
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


def base_test_peer_list(filter_list: list[str] = None) -> None:
    ssh_client.exec_command("nordvpn mesh peer list")

    peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())
    local_hostname = peer_list.get_this_device().hostname
    remote_hostname = peer_list.get_external_peer().hostname

    if filter_list is not None:
        routing_allowed = "routing-allowed" in filter_list
        incoming_traffic_allowed = "incoming-traffic-allowed" in filter_list
        allows_sending_files = "allows-sending-files" in filter_list
        allows_routing = "allows-routing" in filter_list
        allows_incoming_traffic = "allows-incoming-traffic" in filter_list

        meshnet.set_permissions(remote_hostname, routing=routing_allowed, incoming=incoming_traffic_allowed, fileshare=allows_sending_files)
        ssh_client.meshnet.set_permissions(local_hostname, routing=allows_routing, incoming=allows_incoming_traffic, fileshare=allows_sending_files)

    local_peer_list = meshnet.get_clean_peer_list(sh.nordvpn.mesh.peer.list())
    local_formed_list = meshnet.PeerList.from_str(local_peer_list).parse_peer_list(filter_list)

    if len(filter_list) != 0:
        local_peer_list_filtered = meshnet.get_clean_peer_list(str(sh.nordvpn.mesh.peer.list("-f", ",".join(filter_list)))).split("\n")

        assert local_formed_list == local_peer_list_filtered
    else:
        assert local_formed_list == local_peer_list.split("\n")

    if "External Peers:" in "\n".join(local_formed_list) and "External Peers:\n[no peers]" not in "\n".join(local_formed_list):
        assert "Status: connected" in local_formed_list


@pytest.mark.parametrize("external", [True, False], ids=lambda value: "external" if value else "")
@pytest.mark.parametrize("internal", [True, False], ids=lambda value: "internal" if value else "")
@pytest.mark.parametrize("offline", [True, False], ids=lambda value: "offline" if value else "")
@pytest.mark.parametrize("online", [True, False], ids=lambda value: "online" if value else "")
@pytest.mark.flaky(reruns=3, reruns_delay=20)
def test_meshnet_peer_list_state_filters(external, internal, offline, online):
    filter_list = ["external"] * external \
                + ["internal"] * internal \
                + ["offline"] * offline \
                + ["online"] * online

    base_test_peer_list(filter_list)


@pytest.mark.parametrize("allows_incoming_traffic", [True, False], ids=lambda value: "allows_incoming_traffic" if value else "")
@pytest.mark.parametrize("allows_routing", [True, False], ids=lambda value: "allows_routing" if value else "")
@pytest.mark.parametrize("allows_sending_files", [True, False], ids=lambda value: "allows_sending_files" if value else "")
@pytest.mark.parametrize("incoming_traffic_allowed", [True, False], ids=lambda value: "incoming_traffic_allowed" if value else "")
@pytest.mark.parametrize("routing_allowed", [True, False], ids=lambda value: "routing_allowed" if value else "")
@pytest.mark.flaky(reruns=3, reruns_delay=20)
def test_meshnet_peer_list_permission_filters(allows_incoming_traffic, allows_routing, allows_sending_files, incoming_traffic_allowed, routing_allowed):
    filter_list = ["allows-incoming-traffic"] * allows_incoming_traffic \
                + ["allows-routing"] * allows_routing \
                + ["allows-sending-files"] * allows_sending_files \
                + ["incoming-traffic-allowed"] * incoming_traffic_allowed \
                + ["routing-allowed"] * routing_allowed

    base_test_peer_list(filter_list)
