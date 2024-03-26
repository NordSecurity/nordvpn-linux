import pytest
import sh

from lib import daemon, meshnet, network, ssh

ssh_client = ssh.Ssh("qa-peer", "root", "root")


def setup_module(module):  # noqa: ARG001
    meshnet.TestUtils.setup_module(ssh_client)


def teardown_module(module):  # noqa: ARG001
    meshnet.TestUtils.teardown_module(ssh_client)


def setup_function(function):  # noqa: ARG001
    meshnet.TestUtils.setup_function(ssh_client)


def teardown_function(function):  # noqa: ARG001
    meshnet.TestUtils.teardown_function(ssh_client)


@pytest.mark.parametrize("lan_discovery", [True, False])
@pytest.mark.parametrize("local", [True, False])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
def test_killswitch_exitnode(lan_discovery: bool, local: bool):
    my_ip = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_this_device().ip
    peer_ip = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().ip

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
