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

def validate_input_chain(peer_ip: str, routing: bool, local: bool, incoming: bool, fileshare: bool) -> (bool, str):
    rules = sh.sudo.iptables("-S", "INPUT")

    fileshare_rule = f"-A INPUT -s {peer_ip}/32 -p tcp -m tcp --dport 49111 -m comment --comment nordvpn -j ACCEPT"
    if (fileshare_rule in rules) != fileshare:
        return False, f"Fileshare permissions configured incorrectly, rule expected: {fileshare}\nrules:{rules}"

    incoming_rule = f"-A INPUT -s {peer_ip}/32 -m comment --comment nordvpn -j ACCEPT"
    if (incoming_rule in rules) != incoming:
        return False, f"Incoming permissions configured incorrectly, rule expected: {incoming}\nrules:{rules}"

    # If incoming is not enabled, no rules other than fileshare(if enabled) for that peer should be added
    if not incoming:
        if fileshare:
            rules = rules.replace(fileshare_rule, "")
        if peer_ip not in rules:
            return True, ""
        else:
            return False, f"Rules for peer({peer_ip}) found in the INCOMING chain but peer does not have the incoming permissions\nrules:\n{rules}"

    incoming_rule_idx = rules.find(incoming_rule)

    for lan in LANS:
        lan_rule = f"-A INPUT -s {peer_ip}/32 -d {lan} -m comment --comment nordvpn -j DROP"
        lan_rule_idx = rules.find(lan_rule)
        if (routing and local) and lan_rule_idx != -1:
            return False, f"LAN/Routing permissions configured incorrectly\nlocal enabled: {local}\nrouting enabled: {routing}\nrules:\n{rules}"
        # verify that lan_rule is located above the local rule
        if lan_rule_idx > incoming_rule_idx:
            return False, f"LAN/Routing rules ineffective(added after incoming traffic rule)\nlocal enabled: {local}\nrouting enabled: {routing}\nrules:\n{rules}"

    return True, ""


def validate_forward_chain(peer_ip: str, routing: bool, local: bool, incoming: bool, fileshare: bool) -> (bool, str):
    _, _ = incoming, fileshare
    rules = sh.sudo.iptables("-S", "FORWARD")

    # This rule is added above the LAN denial rules if both local and routing is allowed to peer, or bellow LAN denial
    # if only routing is allowed.
    routing_enabled_rule = f"-A FORWARD -s {peer_ip}/32 -m comment --comment nordvpn-exitnode-transient -j ACCEPT"
    routing_enabled_rule_index = rules.find(routing_enabled_rule)

    if routing and (routing_enabled_rule_index == -1):
        return False, f"Routing permission not found\nrules:{rules}"
    if not routing and (routing_enabled_rule_index != -1):
        return False, f"Routing permission found\nrules:{rules}"

    for lan in LANS:
        lan_drop_rule = f"-A FORWARD -s 100.64.0.0/10 -d {lan} -m comment --comment nordvpn-exitnode-transient -j DROP"
        lan_drop_rule_index = rules.find(lan_drop_rule)

        # If any peer has routing or local permission, lan block rules should be added, otherwise no rules should be added.
        if (routing or local) and lan_drop_rule_index == -1:
            return False, f"LAN drop rule not added for subnet {lan}\nrules:\n{rules}"
        elif (not routing) and (not lan) and lan_drop_rule_index != -1:
            return False, f"LAN drop rule added for subnet {lan}\nrules:\n{rules}"

        if routing:
            # Local is allowed, routing rule should be above LAN block rules to allow peer to access any subnet.
            if local and (lan_drop_rule_index < routing_enabled_rule_index):
                return False, f"LAN drop rule for subnet {lan} added before routing\nrules: {rules}"
            # Local is not allowed, routing rule should be below LAN block rules to deny peer access to local subnets.
            if (not local) and (lan_drop_rule_index > routing_enabled_rule_index):
                return False, f"LAN drop rule for subnet {lan} added after routing\nrules: {rules}"
            continue

        # If routing is not enabled, but lan is enabled, there should be one rule for each local network for the peer.
        # They should be located above the LAN block rules.
        if not local:
            continue

        lan_allow_rule = f"-A FORWARD -s {peer_ip}/32 -d {lan} -m comment --comment nordvpn-exitnode-transient -j ACCEPT"
        lan_allow_rule_index = rules.find(lan_allow_rule)

        if lan_allow_rule not in rules:
            return False, f"LAN allow rule for subnet {lan} not found\nrules:\n{rules}"

        if lan_allow_rule_index > lan_drop_rule_index:
            return False, f"LAN allow rule is added after LAN drop rule\nrules:\n{rules}"

    return True, ""


def set_permissions(peer: str, routing: bool, local: bool, incoming: bool, fileshare: bool):
    def bool_to_permission(permission: bool) -> str:
        if permission:
            return "allow"
        return "deny"

    # ignore any failures that might occur when permissions are already configured to the desired value
    sh.nordvpn.mesh.peer.routing(bool_to_permission(routing), peer, _ok_code=(0, 1))
    sh.nordvpn.mesh.peer.local(bool_to_permission(local), peer, _ok_code=(0, 1))
    sh.nordvpn.mesh.peer.incoming(bool_to_permission(incoming), peer, _ok_code=(0, 1))
    sh.nordvpn.mesh.peer.fileshare(bool_to_permission(fileshare), peer, _ok_code=(0, 1))
