import os

import pytest
import requests
import sh
import timeout_decorator

import lib
from lib import daemon, info, logging, login, meshnet, network, settings, ssh

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


