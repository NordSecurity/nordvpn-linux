from lib import logging, daemon, ssh, meshnet, network, login

from lib.shell import sh_no_tty

import os, time, sh

ssh_client = ssh.Ssh("qa-peer", "root", "root")

def teardown_module(module):  # noqa: ARG001
    meshnet.TestUtils.teardown_module(ssh_client)

def teardown_function(function):  # noqa: ARG001
    meshnet.remove_all_peers()
    # meshnet.remove_all_peers_in_peer(ssh_client)
    meshnet.revoke_all_invites()
    # meshnet.revoke_all_invites_in_peer(ssh_client)
    meshnet.TestUtils.teardown_function(ssh_client)


def test_failure():
    os.makedirs("/home/qa/.config/nordvpn", exist_ok=True)
    os.makedirs("/home/qa/.cache/nordvpn", exist_ok=True)
    ssh_client.connect()
    daemon.install_peer(ssh_client)
    meshnet.TestUtils.allowlist_ssh(ssh_client, network.FWMARK)

    logging.log()

    # if setup_function fails, teardown won't be executed, so daemon is not stopped
    if daemon.is_running():
        daemon.stop()

    daemon.start()
    daemon.start_peer(ssh_client)

    # time.sleep(1)
    login.login_as("default")
    # login.login_as("qa-peer", ssh_client)
    # sh_no_tty.nordvpn.set.meshnet.on()
    sh.nordvpn.set.mesh.on()

    # time.sleep(10)

    sh.nordvpn.connect()
