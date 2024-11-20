import glob
import os
import shutil

import sh

from lib import daemon, login, meshnet, ssh

ssh_client = ssh.Ssh("qa-peer", "root", "root")


def setup_module(module):  # noqa: ARG001
    os.system("sudo mkdir -p -m 0777 /home/qa/Downloads")

    ssh_client.connect()
    ssh_client.exec_command("mkdir -p /root/Downloads")


def setup_function(function):  # noqa: ARG001
    sh.sudo.apt.purge("-y", "nordvpn")

    sh.sh(_in=sh.curl("-sSf", "https://downloads.nordcdn.com/apps/linux/install.sh"))

    os.makedirs("/home/qa/.config/nordvpn", exist_ok=True)
    os.makedirs("/home/qa/.cache/nordvpn", exist_ok=True)
    daemon.start()

    login.login_as("default")

    project_root = os.environ["WORKDIR"]
    deb_path = glob.glob(f'{project_root}/dist/app/deb/*amd64.deb')[0]
    sh.sudo.apt.install(deb_path, "-y")

    sh.nordvpn.set.notify.off()
    sh.nordvpn.set.meshnet.on()

    meshnet.remove_all_peers()

    daemon.install_peer(ssh_client)
    daemon.start_peer(ssh_client)
    login.login_as("default", ssh_client)
    ssh_client.exec_command("nordvpn set notify off")
    ssh_client.exec_command("nordvpn set mesh on")

    sh.nordvpn.mesh.peer.list()
    ssh_client.exec_command("nordvpn mesh peer list")


def teardown_function(function):  # noqa: ARG001
    ssh_client.exec_command("rm -rf /root/Downloads/*")

    shutil.rmtree("/home/qa/.config/nordvpn")
    shutil.rmtree("/home/qa/.cache/nordvpn")

    sh.nordvpn.set.mesh.off()
    ssh_client.exec_command("nordvpn set mesh off")

    daemon.stop_peer(ssh_client)
    daemon.uninstall_peer(ssh_client)


def test_meshnet_available_after_update():
    meshnet_help_page = sh.nordvpn.meshnet("--help", _tty_out=False)
    assert "Learn more: https://meshnet.nordvpn.com/" in meshnet_help_page
