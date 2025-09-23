import glob
import os
import shutil

import pytest
import sh

import lib
from lib import daemon, fileshare, login, meshnet, network, poll, ssh, logging
from lib.shell import sh_no_tty
from test_connect import connect_base_test, disconnect_base_test

PROJECT_ROOT = os.environ['WORKDIR']

class TestData:
    INVOLVES_MESHNET = None

class ProductionApplicationData:
    APP_VERSION = None

ssh_client = ssh.Ssh("qa-peer", "root", "root")

def setup_module(module):  # noqa: ARG001
    os.system("sudo mkdir -p -m 0777 /home/qa/Downloads")

    ssh_client.connect()
    ssh_client.exec_command("mkdir -p /root/Downloads")


def setup_function(function):  # noqa: ARG001
    TestData.INVOLVES_MESHNET = any(keyword in os.environ["PYTEST_CURRENT_TEST"] for keyword in ["meshnet", "fileshare"])

    sh.sudo.apt.purge("-y", "nordvpn")

    sh.sh(_in=sh.curl("-sSf", "https://downloads.nordcdn.com/apps/linux/install.sh"))
    ProductionApplicationData.APP_VERSION = sh.nordvpn("-v").split()[2]

    daemon.start()

    daemon.stop() # TODO: LVPN-6403
    deb_path = glob.glob(f'{PROJECT_ROOT}/dist/app/deb/*amd64.deb')[0]
    sh.sudo.apt.install(deb_path, "-y")
    daemon.start() # TODO: LVPN-6403

    # login into the app after update, because if user didn't agree with the consent it will be logged out at update
    login.login_as("default")

    if TestData.INVOLVES_MESHNET:
        sh.nordvpn.set.notify.off()
        assert "Meshnet is set to 'enabled' successfully." in sh.nordvpn.set.meshnet.on()

        meshnet.remove_all_peers()

        daemon.install_peer(ssh_client)
        daemon.start_peer(ssh_client)
        login.login_as("default", ssh_client)
        ssh_client.exec_command("nordvpn set notify off")
        ssh_client.exec_command("nordvpn set mesh on")
        sh.nordvpn.mesh.peer.refresh()

        meshnet.are_peers_connected(ssh_client)


def teardown_function(function):  # noqa: ARG001
    if TestData.INVOLVES_MESHNET:
        logging.log(data="Disable meshnet")
        ssh_client.exec_command("rm -rf /root/Downloads/*")

        sh.nordvpn.set.mesh.off()
        ssh_client.exec_command("nordvpn set mesh off")

        daemon.stop_peer(ssh_client)

        dest_logs_path = f"{PROJECT_ROOT}/dist/logs"
        meshnet.download_remote_peer_logs(ssh_client=ssh_client, dest_logs_path=dest_logs_path)
        shutil.copy("/home/qa/.cache/nordvpn/norduserd.log", dest_logs_path)
        shutil.copy("/home/qa/.cache/nordvpn/nordfileshare.log", dest_logs_path)

        daemon.uninstall_peer(ssh_client)
    daemon.stop() # TODO: LVPN-6403
    assert network.is_disconnected()


@pytest.mark.xfail
def test_meshnet_available_after_update():
    """Manual TC: LVPN-3204"""

    meshnet_help_page = sh_no_tty.nordvpn.meshnet("--help")
    assert "Learn more: https://meshnet.nordvpn.com/" in meshnet_help_page

    parsed_peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())

    local_hostname = parsed_peer_list.get_this_device().hostname
    ssh_client.exec_command(f"nordvpn mesh peer routing allow {local_hostname}")
    sh.nordvpn.mesh.peer.refresh()

    peer_hostname = parsed_peer_list.get_internal_peer().hostname
    output = sh.nordvpn.mesh.peer.connect(peer_hostname)
    assert meshnet.is_connect_successful(output, peer_hostname)
    assert daemon.is_connected()
    assert network.is_available()

    output = sh.nordvpn.disconnect()
    assert lib.is_disconnect_successful(output)
    assert not daemon.is_connected()
    assert network.is_available()


@pytest.mark.xfail
def test_fileshare_available_after_update():
    """Manual TC: LVPN-3205"""

    fileshare_help_page = sh.nordvpn.fileshare("--help", _tty_out=False)
    assert "Learn more: https://meshnet.nordvpn.com/features/sharing-files-in-meshnet" in fileshare_help_page

    wdir = fileshare.create_directory(5)

    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().hostname

    fileshare.start_transfer(peer_hostname, wdir.dir_path)

    for remote_transfer_id, error_message in poll(lambda: fileshare.get_new_incoming_transfer(ssh_client), attempts=5):  # noqa: B007
        if remote_transfer_id is not None:
            break

    assert remote_transfer_id is not None, error_message

    ssh_client.exec_command(f"nordvpn fileshare accept {remote_transfer_id}")

    fileshare.files_from_transfer_exist_in_filesystem(remote_transfer_id, [wdir], ssh_client)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_quick_connect_after_update(tech, proto, obfuscated):
    """Manual TC: LVPN-8506"""

    if tech == "openvpn" and proto == "udp" and obfuscated == "on":
        tech_name = lib.technology_to_upper_camel_case(tech)
        assert f"Technology is set to '{tech_name}' successfully." in sh.nordvpn.set.technology(tech)

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated))
    disconnect_base_test()


def test_changelog_after_update():
    """Manual TC: LVPN-4152"""

    if ProductionApplicationData.APP_VERSION in sh.nordvpn("-v"):
        pytest.skip("Changelog not implemented yet.")

    changelog_path = "/usr/share/doc/nordvpn/changelog.Debian.gz"
    changelog = sh.dpkg_parsechangelog("-l", changelog_path)

    # take version without checksum as it will never be present in the changelog
    nordvpn_version = sh.nordvpn("-v").split()[2].split('+')[0]
    assert nordvpn_version in changelog
    assert "*" in changelog
