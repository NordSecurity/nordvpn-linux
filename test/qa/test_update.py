import glob
import os

import sh

from lib import daemon, login, meshnet, ssh

ssh_client = ssh.Ssh("qa-peer", "root", "root")


def setup_module(module):  # noqa: ARG001
    sh.sudo.apt.purge("-y", "nordvpn")

    sh.sh(_in=sh.curl("-sSf", "https://downloads.nordcdn.com/apps/linux/install.sh"))

    os.makedirs("/home/qa/.config/nordvpn", exist_ok=True)
    os.makedirs("/home/qa/.cache/nordvpn", exist_ok=True)
    daemon.start()

    os.system("sudo mkdir -p -m 0777 /home/qa/Downloads")

    login.login_as("default")

    project_root = os.environ["WORKDIR"]
    deb_path = glob.glob(f'{project_root}/dist/app/deb/*amd64.deb')[0]
    sh.sudo.apt.install(deb_path, "-y")

    sh.nordvpn.set.notify.off()
    sh.nordvpn.set.meshnet.on()

    meshnet.remove_all_peers()

    ssh_client.connect()
    ssh_client.exec_command("mkdir -p /root/Downloads")
    daemon.install_peer(ssh_client)
    daemon.start_peer(ssh_client)
    login.login_as("default", ssh_client)
    ssh_client.exec_command("nordvpn set notify off")
    ssh_client.exec_command("nordvpn set mesh on")

    sh.nordvpn.mesh.peer.list()
    ssh_client.exec_command("nordvpn mesh peer list")
