import glob
import os
import shutil

import sh

import lib
from lib import daemon, fileshare, login, meshnet, network, poll, ssh


class TestData:
    INVOLVES_MESHNET = None

ssh_client = ssh.Ssh("qa-peer", "root", "root")

def setup_module(module):  # noqa: ARG001
    os.system("sudo mkdir -p -m 0777 /home/qa/Downloads")

    ssh_client.connect()
    ssh_client.exec_command("mkdir -p /root/Downloads")


def setup_function(function):  # noqa: ARG001
    TestData.INVOLVES_MESHNET = any(keyword in os.environ["PYTEST_CURRENT_TEST"] for keyword in ["meshnet", "fileshare"])

    sh.sudo.apt.purge("-y", "nordvpn")

    sh.sh(_in=sh.curl("-sSf", "https://downloads.nordcdn.com/apps/linux/install.sh"))

    daemon.start()

    login.login_as("default")

    project_root = os.environ["WORKDIR"]
    deb_path = glob.glob(f'{project_root}/dist/app/deb/*amd64.deb')[0]
    sh.sudo.apt.install(deb_path, "-y")

    if TestData.INVOLVES_MESHNET:
        sh.nordvpn.set.notify.off()
        assert "Meshnet is set to 'enabled' successfully." in sh.nordvpn.set.meshnet.on()

        meshnet.remove_all_peers()

        daemon.install_peer(ssh_client)
        daemon.start_peer(ssh_client)
        login.login_as("default", ssh_client)
        ssh_client.exec_command("nordvpn set notify off")
        ssh_client.exec_command("nordvpn set mesh on")

        sh.nordvpn.mesh.peer.list()
        ssh_client.exec_command("nordvpn mesh peer list")


def teardown_function(function):  # noqa: ARG001
    if TestData.INVOLVES_MESHNET:
        ssh_client.exec_command("rm -rf /root/Downloads/*")

        sh.nordvpn.set.mesh.off()
        ssh_client.exec_command("nordvpn set mesh off")

        daemon.stop_peer(ssh_client)
        daemon.uninstall_peer(ssh_client)


def test_meshnet_available_after_update():
    meshnet_help_page = sh.nordvpn.meshnet("--help", _tty_out=False)
    assert "Learn more: https://meshnet.nordvpn.com/" in meshnet_help_page

    local_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_this_device().hostname
    ssh_client.exec_command(f"nordvpn mesh peer routing allow {local_hostname}")

    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().hostname
    output = sh.nordvpn.mesh.peer.connect(peer_hostname)
    assert meshnet.is_connect_successful(output, peer_hostname)

    assert network.is_available()

    output = sh.nordvpn.disconnect()
    assert lib.is_disconnect_successful(output)


def test_fileshare_available_after_update():
    fileshare_help_page = sh.nordvpn.fileshare("--help", _tty_out=False)
    assert "Learn more: https://meshnet.nordvpn.com/features/sharing-files-in-meshnet" in fileshare_help_page

    wdir = fileshare.create_directory(5)

    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().hostname

    fileshare.start_transfer(peer_hostname, wdir.dir_path)

    for remote_transfer_id, error_message in poll(lambda: fileshare.get_new_incoming_transfer(ssh_client)):  # noqa: B007
        if remote_transfer_id is not None:
            break

    assert remote_transfer_id is not None, error_message

    ssh_client.exec_command(f"nordvpn fileshare accept {remote_transfer_id}")

    fileshare.files_from_transfer_exist_in_filesystem(remote_transfer_id, [wdir], ssh_client)
