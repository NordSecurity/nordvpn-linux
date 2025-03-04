import pytest
import sh
import grpc

import lib
import sys
import os
from lib import daemon, meshnet, settings, ssh
from lib.shell import sh_no_tty

sys.path.append(os.path.abspath(os.path.join(
    os.path.dirname(__file__), 'lib/protobuf/meshnet')))

from lib.protobuf.meshnet import (service_pb2_grpc, empty_pb2)
import lib.protobuf.meshnet

ssh_client = ssh.Ssh("qa-peer", "root", "root")


def setup_module(module):  # noqa: ARG001
    meshnet.TestUtils.setup_module(ssh_client)


def teardown_module(module):  # noqa: ARG001
    meshnet.TestUtils.teardown_module(ssh_client)


def setup_function(function):  # noqa: ARG001
    meshnet.TestUtils.setup_function(ssh_client)


def teardown_function(function):  # noqa: ARG001
    meshnet.TestUtils.teardown_function(ssh_client)


# This doesn't directly test meshnet, but it uses it
def test_allowlist_incoming_connection():
    my_ip = ssh_client.exec_command("echo $SSH_CLIENT").split()[0]

    peer_hostname = meshnet.PeerList.from_str(sh_no_tty.nordvpn.mesh.peer.list()).get_external_peer().hostname
    # Initiate ssh connection via mesh because we are going to lose the main connection
    ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
    ssh_client_mesh.connect()

    ssh_client_mesh.exec_command("nordvpn set killswitch on")
    # We should not have direct connection anymore after connecting to VPN
    with pytest.raises(sh.ErrorReturnCode_1):
        assert "icmp_seq=" not in sh.ping("-c", "1", "qa-peer")

        ssh_client_mesh.exec_command(f"nordvpn allowlist add subnet {my_ip}/32")
        # Direct connection should work again after allowlisting
        assert "icmp_seq=" in sh.ping("-c", "1", "qa-peer")
    ssh_client_mesh.exec_command("nordvpn set killswitch off")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
# This doesn't directly test meshnet, but it uses it
def test_set_defaults_when_logged_in_2nd_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh_no_tty.nordvpn.set.fwmark("0xe2f2")
    sh_no_tty.nordvpn.set.killswitch("on")
    sh_no_tty.nordvpn.set.tpl("on")
    sh_no_tty.nordvpn.set.autoconnect("on")
    sh_no_tty.nordvpn.set("lan-discovery", "on")

    assert settings.is_meshnet_enabled()
    assert "0xe1f1" not in  sh_no_tty.nordvpn.settings()
    assert daemon.is_killswitch_on()
    assert settings.is_tpl_enabled()
    assert settings.is_autoconnect_enabled()
    assert settings.is_lan_discovery_enabled()

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled()
    else:
        assert not settings.is_obfuscated_enabled()

    assert "Settings were successfully restored to defaults." in  sh_no_tty.nordvpn.set.defaults()

    assert settings.app_has_defaults_settings()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
# This doesn't directly test meshnet, but it uses it
def test_set_defaults_when_logged_out_1st_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh_no_tty.nordvpn.set.fwmark("0xe2f2")
    sh_no_tty.nordvpn.set.killswitch("on")
    sh_no_tty.nordvpn.set("lan-discovery", "on")
    sh_no_tty.nordvpn.set.analytics("off")
    sh_no_tty.nordvpn.set.tpl("on")

    assert settings.is_meshnet_enabled()
    assert "0xe1f1" not in  sh_no_tty.nordvpn.settings()
    assert daemon.is_killswitch_on()
    assert settings.is_lan_discovery_enabled()
    assert not settings.are_analytics_enabled()
    assert settings.is_tpl_enabled()

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled()
    else:
        assert not settings.is_obfuscated_enabled()

    sh_no_tty.nordvpn.logout("--persist-token")

    assert "Settings were successfully restored to defaults." in  sh_no_tty.nordvpn.set.defaults()

    assert settings.app_has_defaults_settings()


# This doesn't directly test meshnet, but it uses it
def test_set_post_quantum_on_meshnet_enabled():

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh_no_tty.nordvpn.set(settings.get_pq_alias(), "on")

    assert "Post-quantum encryption and Meshnet are not compatible. Please disable one feature to use the other." in str(ex.value)


# This doesn't directly test meshnet, but it uses it
def test_set_meshnet_on_post_quantum_enabled():

    sh_no_tty.nordvpn.set.meshnet("off")

    sh_no_tty.nordvpn.set(settings.get_pq_alias(), "on")

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh_no_tty.nordvpn.set.meshnet("on")

    assert "Post-quantum encryption and Meshnet are not compatible. Please disable one feature to use the other." in str(ex.value)

def test_mesh_private_key_is_revoked_on_mesh_off():
    sh.nordvpn.set.meshnet("off")

    with grpc.insecure_channel(lib.NORDVPND_SOCKET) as channel:
        stub = service_pb2_grpc.MeshnetStub(channel)
        response = stub.GetPrivateKey(empty_pb2.Empty())
        assert response.private_key == "", "Meshnet private key should be removed when meshnet is disabled."

def test_mesh_private_key_is_revoked_on_mesh_off_vpn_disconnect():
    sh.nordvpn.connect()
    sh.nordvpn.set.meshnet("off")

    with grpc.insecure_channel(lib.NORDVPND_SOCKET) as channel:
        stub = service_pb2_grpc.MeshnetStub(channel)
        response = stub.GetPrivateKey(empty_pb2.Empty())
        assert response.private_key != "", "Meshnet private key should not be removed mid-VPN connection."

        sh.nordvpn.disconnect()
        response = stub.GetPrivateKey(empty_pb2.Empty())
        assert response.private_key == "", "Meshnet private key should be removed when meshnet is disabled."

def test_mesh_private_key_is_revoked_on_mesh_off_daemon_shutdown():
    sh.nordvpn.connect()
    sh.nordvpn.set.meshnet("off")

    with grpc.insecure_channel(lib.NORDVPND_SOCKET) as channel:
        stub = service_pb2_grpc.MeshnetStub(channel)
        response = stub.GetPrivateKey(empty_pb2.Empty())
        assert response.private_key != "", "Meshnet private key should not be removed mid-VPN connection."

    daemon.restart()
    def is_meshnet_pk_removed() -> bool:
        try:
            with grpc.insecure_channel(lib.NORDVPND_SOCKET) as channel:
                stub = service_pb2_grpc.MeshnetStub(channel)
                response = stub.GetPrivateKey(empty_pb2.Empty())
                return response.private_key == "", "Meshnet private key should be removed when meshnet is disabled."
        except Exception:# noqa: BLE001
            return False

    pk_removed = False
    for pk_removed in lib.poll(is_meshnet_pk_removed):
        if pk_removed:
            break
    assert pk_removed, "Meshnet private key was not removed after daemon shutdown."

