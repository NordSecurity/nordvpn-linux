import os
import re
import subprocess
import time
from enum import Enum

import sh

from . import login, ssh

PEER_USERNAME = os.environ.get("QA_PEER_USERNAME")

LANS = [
    "169.254.0.0/16",
    "192.168.0.0/16",
    "172.16.0.0/12",
    "10.0.0.0/8",
]

strip_colors = re.compile(r'\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])', flags=re.IGNORECASE)


class PeerName(Enum):
    Hostname = 0
    Ip = 1
    Pubkey = 2


# Used for test parametrization, when the same test has to be run with different Meshnet alias.
MESHNET_ALIAS = [
    "meshnet",
    "mesh"
]
    

def get_peer_name(output: str, name_type: PeerName) -> str:
    match name_type:
        case PeerName.Hostname:
            return get_this_device(output)
        case PeerName.Ip:
            return get_this_device_ipv4(output)
        case PeerName.Pubkey:
            return get_this_device_pubkey(output)


def add_peer(ssh_client: ssh.Ssh,
             tester_allow_fileshare: bool = True,
             tester_allow_routing: bool = True,
             tester_allow_local: bool = True,
             tester_allow_incoming: bool = True,
             peer_allow_fileshare: bool = True,
             peer_allow_routing: bool = True,
             peer_allow_local: bool = True,
             peer_allow_incoming: bool = True):
    """
    Adds QA peer to meshnet.

    Try to minimize usage of this, because there's a weekly invite limit.
    """
    tester_allow_fileshare_arg = f"--allow-peer-send-files={str(tester_allow_fileshare).lower()}"
    tester_allow_routing_arg = f"--allow-traffic-routing={str(tester_allow_routing).lower()}"
    tester_allow_local_arg = f"--allow-local-network-access={str(tester_allow_local).lower()}"
    tester_allow_incoming_arg = f"--allow-incoming-traffic={str(tester_allow_incoming).lower()}"

    peer_allow_fileshare_arg = f"--allow-peer-send-files={str(peer_allow_fileshare).lower()}"
    peer_allow_routing_arg = f"--allow-traffic-routing={str(peer_allow_routing).lower()}"
    peer_allow_local_arg = f"--allow-local-network-access={str(peer_allow_local).lower()}"
    peer_allow_incoming_arg = f"--allow-incoming-traffic={str(peer_allow_incoming).lower()}"

    sh.nordvpn.mesh.inv.send(tester_allow_incoming_arg, tester_allow_local_arg, tester_allow_routing_arg, tester_allow_fileshare_arg, PEER_USERNAME)
    local_user, _ = login.get_default_credentials()
    ssh_client.exec_command(f"yes | nordvpn mesh inv accept {peer_allow_local_arg} {peer_allow_incoming_arg} {peer_allow_routing_arg} {peer_allow_fileshare_arg} {local_user}")

    sh.nordvpn.mesh.peer.refresh()


def get_peers(output: str) -> list:
    """Parses list of peer names from 'nordvpn meshnet peer list' output."""
    output = output[output.find("Local Peers:"):]  # skip this device
    peers = []
    for line in output.split("\n"):
        if "Hostname:" in line:
            peers.append(strip_colors.sub('', line.split(" ")[-1]))
    return peers


def get_this_device(output: str):
    """Parses current device hostname from 'nordvpn meshnet peer list' output."""
    output_lines = output.split("\n")
    for i, line in enumerate(output_lines):
        if "This device:" in line:
            for subline in output_lines[i + 1:]:
                if "Hostname:" in subline:
                    return strip_colors.sub('', subline.split(" ")[-1])
    return None


def get_this_device_ipv4(output: str):
    """Parses current device ip from 'nordvpn meshnet peer list' output."""
    output_lines = output.split("\n")
    for i, line in enumerate(output_lines):
        if "This device:" in line:
            for subline in output_lines[i + 1:]:
                if "IP:" in subline:
                    return strip_colors.sub('', subline.split(" ")[-1])
    return None


def get_this_device_pubkey(output: str):
    """Parses current device pubkey from 'nordvpn meshnet peer list' output."""
    output_lines = output.split("\n")
    for i, line in enumerate(output_lines):
        if "This device:" in line:
            for subline in output_lines[i + 1:]:
                if "Public Key:" in subline:
                    return strip_colors.sub('', subline.split(" ")[-1])
    return None


def remove_all_peers():
    """Removes all meshnet peers from local device."""
    output = f"{sh.nordvpn.mesh.peer.list(_tty_out=False)}"  # convert to string, _tty_out false disables colors
    for p in get_peers(output):
        sh.nordvpn.mesh.peer.remove(p)


def remove_all_peers_in_peer(ssh_client: ssh.Ssh):
    """Removes all meshnet peers from peer device."""
    output = ssh_client.exec_command("nordvpn mesh peer list")
    for p in get_peers(output):
        ssh_client.exec_command(f"nordvpn mesh peer remove {p}")


def is_peer_reachable(ssh_client: ssh.Ssh, retry: int = 5) -> bool:
    """Returns True when ping to peer succeeds."""
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
    """Parses list of sent invites from 'nordvpn meshnet inv list' output."""
    emails = []
    for line in output.split("\n"):
        if line.find("Received Invites:") != -1:
            break  # End of sent invites
        if line.find("Email:") != -1:
            emails.append(line.split(" ")[1])
    return emails


def revoke_all_invites():
    """Revokes all sent meshnet invites in local device."""
    output = f"{sh.nordvpn.mesh.inv.list(_tty_out=False)}"  # convert to string, _tty_out false disables colors
    for i in get_sent_invites(output):
        sh.nordvpn.mesh.inv.revoke(i)


def revoke_all_invites_in_peer(ssh_client: ssh.Ssh):
    """Revokes all sent meshnet invites in peer device."""
    output = ssh_client.exec_command("nordvpn mesh inv list")
    for i in get_sent_invites(output):
        ssh_client.exec_command(f"nordvpn mesh inv revoke {i}")


def send_meshnet_invite(email):
    try:
        command = ["nordvpn", "meshnet", "invite", "send", email]
        process = subprocess.Popen(command, stdout=subprocess.PIPE, stdin=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

        for _ in range(4):
            process.stdin.write('\n')
            process.stdin.flush()

        try:
            stdout, stderr = process.communicate(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            stdout, stderr = process.communicate()

        if process.returncode != 0:
            raise sh.ErrorReturnCode_1(full_cmd="", stdout=b"", stderr=stdout.split('\n')[-2].encode('utf-8'))

        return stdout.strip().split('\n')[-1]
    except subprocess.CalledProcessError as e:
        print(f"Error occurred: {e}")
        raise sh.ErrorReturnCode_1 from None


def accept_meshnet_invite(ssh_client: ssh.Ssh,
             peer_allow_fileshare: bool = True,
             peer_allow_routing: bool = True,
             peer_allow_local: bool = True,
             peer_allow_incoming: bool = True):

    peer_allow_fileshare_arg = f"--allow-peer-send-files={str(peer_allow_fileshare).lower()}"
    peer_allow_routing_arg = f"--allow-traffic-routing={str(peer_allow_routing).lower()}"
    peer_allow_local_arg = f"--allow-local-network-access={str(peer_allow_local).lower()}"
    peer_allow_incoming_arg = f"--allow-incoming-traffic={str(peer_allow_incoming).lower()}"

    local_user, _ = login.get_default_credentials()
    output = ssh_client.exec_command(f"yes | nordvpn mesh inv accept {peer_allow_local_arg} {peer_allow_incoming_arg} {peer_allow_routing_arg} {peer_allow_fileshare_arg} {local_user}")
    sh.nordvpn.mesh.peer.refresh()

    return output


def deny_meshnet_invite(ssh_client: ssh.Ssh):

    local_user, _ = login.get_default_credentials()
    output = ssh_client.exec_command(f"yes | nordvpn mesh inv deny {local_user}")
    
    return output