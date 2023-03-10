from lib import login, ssh
from enum import Enum
import sh
import os
import time
import logging as logger
import re

PEER_USERNAME = os.environ.get("QA_PEER_USERNAME")

class PeerName(Enum):
    Hostname = 0
    Ip = 1
    Pubkey = 2

def get_peer_name(output: str, name_type: PeerName) -> str:
    match name_type:
        case PeerName.Hostname:
            return get_this_device(output)
        case PeerName.Ip:
            return get_this_device_ipv4(output)
        case PeerName.Pubkey:
            return get_this_device_pubkey(output)

def add_peer(ssh_client: ssh.Ssh, tester_allow_fileshare: bool = True, peer_allow_fileshare: bool = True):
    """
    adds QA peer to meshnet
    try to minimize usage of this, because there's a weekly invite limit
    """
    tester_allow_fileshare_arg = f"-allow-peer-send-files={str(tester_allow_fileshare).lower()}"
    peer_allow_fileshare_arg = f"-allow-peer-send-files={str(peer_allow_fileshare).lower()}"
    sh.nordvpn.mesh.inv.send("--allow-incoming-traffic=true", "--allow-traffic-routing=true", tester_allow_fileshare_arg, PEER_USERNAME)
    local_user, _ = login.get_default_credentials()
    ssh_client.exec_command(f"yes | nordvpn mesh inv accept --allow-incoming-traffic=true --allow-traffic-routing=true {peer_allow_fileshare_arg} {local_user}")

def get_peers(output: str) -> list:
    """parses list of peer names from 'nordvpn meshnet peer list' output"""
    output = output[output.find("Local Peers:"):] # skip this device
    peers = []
    for line in output.split("\n"):
        if line.find("Hostname:") != -1:
            peers.append(line.split(" ")[1])
    return peers

def get_this_device(output: str):
    """parses current device hostname from 'nordvpn meshnet peer list' output"""
    output_lines = output.split("\n")
    for i, line in enumerate(output_lines):
        if line.find("This device:") != -1:
            return output_lines[i+1].split(" ")[1]

def get_this_device_ipv4(output: str):
    """parses current device ip from 'nordvpn meshnet peer list' output"""
    output_lines = output.split("\n")
    for i, line in enumerate(output_lines):
        if line.find("This device:") != -1:
            return output_lines[i+2].split(" ")[1]

def get_this_device_pubkey(output: str):
    """parses current device pubkey from 'nordvpn meshnet peer list' output"""
    output_lines = output.split("\n")
    for i, line in enumerate(output_lines):
        if line.find("This device:") != -1:
            # example: Public Key: uAexQo2yuiVBZocvuiFPQjAujkDmQVemKaircpxDaUc=
            return output_lines[i+3].split(" ")[2]

def remove_all_peers():
    """removes all meshnet peers from local device"""
    output = f"{sh.nordvpn.mesh.peer.list(_tty_out=False)}" # convert to string, _tty_out false disables colors
    for p in get_peers(output):
        sh.nordvpn.mesh.peer.remove(p)

def remove_all_peers_in_peer(ssh_client: ssh.Ssh):
    """removes all meshnet peers from peer device"""
    output = ssh_client.exec_command("nordvpn mesh peer list")
    for p in get_peers(output):
        ssh_client.exec_command(f"nordvpn mesh peer remove {p}")

def is_peer_reachable(ssh_client: ssh.Ssh, retry: int = 5) -> bool:
    """returns True when ping to peer succeeds."""
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_hostname = get_this_device(output)
    i = 0
    while i < retry:
        try:
            return "icmp_seq=" in sh.ping("-c", "1", peer_hostname)
        except sh.ErrorReturnCode as e:
            print(e.stdout)
            print(e.stderr)
            time.sleep(1)
            i += 1
    print(sh.nordvpn.mesh.peer.list())
    output = ssh_client.exec_command("nordvpn mesh peer list")
    print(output)
    return False

def get_sent_invites(output: str) -> list:
    """parses list of sent invites from 'nordvpn meshnet inv list' output"""
    emails = []
    for line in output.split("\n"):
        if line.find("Received Invites:") != -1:
            break # End of sent invites
        if line.find("Email:") != -1:
            emails.append(line.split(" ")[1])
    return emails

def revoke_all_invites():
    """revokes all sent meshnet invites in local device"""
    output = f"{sh.nordvpn.mesh.inv.list(_tty_out=False)}" # convert to string, _tty_out false disables colors
    for i in get_sent_invites(output):
        sh.nordvpn.mesh.inv.revoke(i)

def revoke_all_invites_in_peer(ssh_client: ssh.Ssh):
    """revokes all sent meshnet invites in peer device"""
    output = ssh_client.exec_command("nordvpn mesh inv list")
    for i in get_sent_invites(output):
        ssh_client.exec_command(f"nordvpn mesh inv revoke {i}")