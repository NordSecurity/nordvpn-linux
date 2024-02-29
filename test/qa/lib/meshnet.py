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


class Peer:
    def __init__(
            self,
            hostname: str,
            nickname: str,
            ip: str,
            public_key: str,
            os: str,
            distribution: str,
            status: str = None,
            allow_incoming_traffic: bool = None,
            allow_routing: bool = None,
            allow_lan_access: bool = None,
            allow_sending_files: bool = None,
            allows_incoming_traffic: bool = None,
            allows_routing: bool = None,
            allows_lan_access: bool = None,
            allows_sending_files: bool = None,
            accept_fileshare_automatically: bool = None
            ):
        self.hostname = hostname
        self.nickname = nickname
        self.status = status
        self.ip = ip
        self.public_key = public_key
        self.os = os
        self.distribution = distribution
        self.allow_incoming_traffic = self._convert_to_bool(allow_incoming_traffic)
        self.allow_routing = self._convert_to_bool(allow_routing)
        self.allow_lan_access = self._convert_to_bool(allow_lan_access)
        self.allow_sending_files = self._convert_to_bool(allow_sending_files)
        self.allows_incoming_traffic = self._convert_to_bool(allows_incoming_traffic)
        self.allows_routing = self._convert_to_bool(allows_routing)
        self.allows_lan_access = self._convert_to_bool(allows_lan_access)
        self.allows_sending_files = self._convert_to_bool(allows_sending_files)
        self.accept_fileshare_automatically = self._convert_to_bool(accept_fileshare_automatically)

    def _convert_to_bool(self, value):
        return value.lower() == "enabled" if value is not None else None


class PeerName(Enum):
    Hostname = 0
    Ip = 1
    Pubkey = 2


class PeerList:
    def __init__(self):
        self.this_device: list[Peer] = []
        self.internal_peers: list[Peer] = []
        self.external_peers: list[Peer] = []

    def _str_to_dictionary(self, data: str):
        new_dictionary = {a.strip().lower(): b.strip()
                for a, b in (element.split(':')
                                for element in
                                filter(lambda line: len(line.split(':')) == 2, data.split('\n')))}

        return new_dictionary

    def _add_peer(self, to_peer_list: list[Peer], peer_data: str):
        peer_data_dictionary = self._str_to_dictionary(peer_data)

        # This device case
        peer = Peer(
                hostname = peer_data_dictionary['hostname'],
                nickname = peer_data_dictionary["nickname"],
                ip = peer_data_dictionary["ip"],
                public_key = peer_data_dictionary["public key"],
                os = peer_data_dictionary["os"],
                distribution = peer_data_dictionary["distribution"],
            )

        # Internal, external peer cases
        if "Status" in peer_data_dictionary:
            peer.status = peer_data_dictionary["status"]
            peer.allow_incoming_traffic = peer_data_dictionary["allow incoming traffic"]
            peer.allow_routing = peer_data_dictionary["allow routing"]
            peer.allow_lan_access = peer_data_dictionary["allow local network access"]
            peer.allow_sending_files = peer_data_dictionary["allow sending files"]
            peer.allows_incoming_traffic = peer_data_dictionary["allows incoming traffic"]
            peer.allows_routing = peer_data_dictionary["allows routing"]
            peer.allows_lan_access = peer_data_dictionary["allows local network access"]
            peer.allows_sending_files = peer_data_dictionary["allows sending files"]
            peer.accept_fileshare_automatically = peer_data_dictionary["accept fileshare automatically"]
        
        to_peer_list.append(peer)

    def set_this_device(self, peer_data: str):
        self.this_device = []
        self._add_peer(self.this_device, peer_data)

    def get_this_device(self) -> Peer:
        return self.this_device[0]


    def add_internal_peer(self, peer_data: str) -> None:
        self._add_peer(self.internal_peers, peer_data)

    def get_internal_peer(self) -> Peer | None:
        if len(self.internal_peers) != 0:
            return self.internal_peers[0]
        return None

    def get_all_internal_peers(self) -> list[Peer]:
        return self.internal_peers


    def add_external_peer(self, peer_data: str):
        self._add_peer(self.external_peers, peer_data)

    def get_external_peer(self) -> Peer | None:
        if len(self.external_peers) != 0:
            return self.external_peers[0]
        return None

    def get_all_external_peers(self) -> list[Peer]:
        return self.external_peers


# Used for test parametrization, when the same test has to be run with different Meshnet alias.
MESHNET_ALIAS = [
    "meshnet",
    "mesh"
]

def get_peer_name(peer: Peer, name_type: PeerName) -> str:
    match name_type:
        case PeerName.Hostname:
            return peer.hostname
        case PeerName.Ip:
            return peer.ip
        case PeerName.Pubkey:
            return peer.public_key


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


def remove_all_peers():
    """Removes all meshnet peers from local device."""
    peer_list = parse_peer_list(sh.nordvpn.mesh.peer.list())

    for peer in peer_list.get_all_internal_peers() + peer_list.get_all_external_peers():
        sh.nordvpn.mesh.peer.remove(peer.hostname)


def remove_all_peers_in_peer(ssh_client: ssh.Ssh):
    """Removes all meshnet peers from peer device."""
    peer_list = parse_peer_list(ssh_client.exec_command("nordvpn mesh peer list"))

    for peer in peer_list.get_all_internal_peers() + peer_list.get_all_external_peers():
        ssh_client.exec_command(f"nordvpn mesh peer remove {peer.hostname}")


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


def get_clean_peer_list(peer_list: str):
    output = strip_colors.sub('', str(peer_list))
    output = "This " + output.split("This", 1)[-1].strip()
    return output


def parse_peer_list(output: str):
    """ Converts output/meshnet peer list string to PeerList object. """

    def remove_text_before_and_keyword(input_string, keyword):
        index = input_string.find(keyword)

        if index != -1:
            return input_string[index + len(keyword):]
        else:
            return input_string

    peer_list = get_clean_peer_list(output)
    peer_list_object = PeerList()

    this_device = peer_list.split("\n\n")[0].replace("This device:\n", "")
    peer_list_object.set_this_device(this_device)

    internal_peers = remove_text_before_and_keyword(peer_list, "Local Peers:\n").split("\n\n\n")[0]
    if "[no peers]" not in internal_peers:
        internal_peer_list = internal_peers.split("\n\n")

        for peer_data in internal_peer_list:
            peer_list_object.add_internal_peer(peer_data)

    external_peers = remove_text_before_and_keyword(peer_list, "External Peers:\n")
    if "[no peers]" not in external_peers:
        external_peer_list = external_peers.split("\n\n")

        for peer_data in external_peer_list:
            peer_list_object.add_external_peer(peer_data)

    return peer_list_object


def is_peer_reachable(ssh_client: ssh.Ssh, peer: Peer, retry: int = 5) -> bool:
    """Returns True when ping to peer succeeds."""
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_hostname = peer.hostname
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
