import logging
import time
from urllib.parse import quote

import requests
from . import shell

TECH_OPENVPN_UDP_OBFUSCATION_ON = 15
TECH_OPENVPN_UDP_OBFUSCATION_OFF = 3
TECH_OPENVPN_TCP_OBFUSCATION_ON = 17
TECH_OPENVPN_TCP_OBFUSCATION_OFF = 5
TECH_OPENVPN_LIST = [
    TECH_OPENVPN_UDP_OBFUSCATION_ON,
    TECH_OPENVPN_UDP_OBFUSCATION_OFF,
    TECH_OPENVPN_TCP_OBFUSCATION_ON,
    TECH_OPENVPN_TCP_OBFUSCATION_OFF,
]
TECH_NORDLYNX = 35
TECH_NORDWHISPER = 51

TECH_IDS = {
    "openvpn": {
        "udp": {
            "on": TECH_OPENVPN_UDP_OBFUSCATION_ON,
            "off": TECH_OPENVPN_UDP_OBFUSCATION_OFF,
        },
        "tcp": {
            "on": TECH_OPENVPN_TCP_OBFUSCATION_ON,
            "off": TECH_OPENVPN_TCP_OBFUSCATION_OFF,
        },
    },
    "nordlynx": TECH_NORDLYNX,
    "nordwhisper": TECH_NORDWHISPER,
}

GROUP_DOUBLE_VPN = 1
GROUP_ONION_OVER_VPN = 3
GROUP_DEDICATED_IP = 9
GROUP_STANDARD_VPN_SERVERS = 11
GROUP_P2P = 15
GROUP_OBFUSCATED_SERVERS = 17
GROUP_EUROPE = 19
GROUP_THE_AMERICAS = 21
GROUP_ASIA_PACIFIC = 23
GROUP_AFRICA_THE_MIDDLE_EAST_AND_INDIA = 25

GROUP_IDS = {
    "Double_VPN": GROUP_DOUBLE_VPN,
    "Onion_Over_VPN": GROUP_ONION_OVER_VPN,
    "Dedicated_IP": GROUP_DEDICATED_IP,
    "Standard_VPN_Servers": GROUP_STANDARD_VPN_SERVERS,
    "P2P": GROUP_P2P,
    "Obfuscated_Servers": GROUP_OBFUSCATED_SERVERS,
    "Europe": GROUP_EUROPE,
    "The_Americas": GROUP_THE_AMERICAS,
    "Asia_Pacific": GROUP_ASIA_PACIFIC,
    "Africa_The_Middle_East_And_India": GROUP_AFRICA_THE_MIDDLE_EAST_AND_INDIA,
}

class ServerInfo:
    def __init__(self, server_info):
        self.name = server_info["name"]
        self.hostname = server_info["hostname"]
        self.city = server_info["locations"][0]["country"]["city"]["name"]
        self.country = server_info["locations"][0]["country"]["name"]


def get_hostname_by(technology="", protocol="", obfuscated="", group_name=""):
    """Returns server name and hostname from core API."""

    (tech_id, group_id) = get_request_parameters(technology, protocol, obfuscated, group_name)

    # api limits
    time.sleep(2)

    url = f"https://api.nordvpn.com/v1/servers?limit=10&filters[servers.status]=online&filters[servers_technologies][id]={tech_id}&filters[servers_groups][id]={group_id}"
    logging.debug(url)
    response = requests.get(url, timeout=5).json()
    assert len(response) > 0, "API returned an empty servers list"
    logging.debug(response)
    server = response[0]
    validate_server(server_json=str(server), tech_id=tech_id, group_id=group_id)
    return ServerInfo(server_info=server)


def get_server_info(server_name):
    server_name = quote(server_name)
    url = f"https://api.nordvpn.com/v1/servers?filters[servers.name]={server_name}&fields[servers.locations]"
    server_info = requests.get(url, timeout=5).json()

    city = server_info[0]["locations"][0]["country"]["city"]["name"]
    country = server_info[0]["locations"][0]["country"]["name"]

    return city, country

# TODO: LVPN-7744
def get_dedicated_ip():
    """Returns Dedicated IP server name."""
    token = shell.sh_no_tty.nordvpn.token().split("\n")[1].split(" ")[1]

    headers = {'Accept': 'application/json', 'Authorization': 'Bearer token:' + token}
    response = requests.get('https://api.nordvpn.com/v1/users/services', headers=headers, timeout=5)

    dedicated_ip_part = ""

    for itm in response.json():
        if "Dedicated IP" in str(itm):
            dedicated_ip_part = itm
            break

    if dedicated_ip_part == "":
        return None

    dip_server_id = dedicated_ip_part['details']['servers'][0]['id']

    headers = {'Accept': 'application/json'}
    response = requests.get(f'https://api.nordvpn.com/v1/servers?&filters[servers.id]={dip_server_id}', headers=headers, timeout=5)
    server_info = response.json()[0]
    return ServerInfo(server_info=server_info)


def get_request_parameters(technology="", protocol="", obfuscated="", group_name=""):
    """Returns (endpoint, technology id, group id) for the core API."""
    tech_id = None
    group_id = 0

    if technology != "":
        if protocol != "":
            tech_id = TECH_IDS.get(technology, {}).get(protocol, {}).get(obfuscated, None)
        else:
            tech_id = TECH_IDS.get(technology)

    if group_name != "":
        group_id = GROUP_IDS.get(group_name, GROUP_STANDARD_VPN_SERVERS)
    elif tech_id in [TECH_OPENVPN_UDP_OBFUSCATION_ON, TECH_OPENVPN_TCP_OBFUSCATION_ON]:
        # If the technology requires obfuscated servers we must also specify it in the group id
        group_id = GROUP_OBFUSCATED_SERVERS
    else:
        # If group_id is empty, the API will default it to 11, GROUP_STANDARD_VPN_SERVERS
        # So let's just set it to standard if there is no other group specifications
        group_id = GROUP_STANDARD_VPN_SERVERS

    if tech_id is None:
        tech_id = ""

    return (tech_id, group_id)

def validate_server(server_json="", tech_id=0, group_id=0):
    if tech_id in TECH_OPENVPN_LIST:
        assert "OpenVPN" in server_json, "The server does not support OpenVPN"
    if tech_id == TECH_NORDLYNX:
        assert "wireguard" in server_json, "The server does not support Nordlynx"
    if tech_id == TECH_NORDWHISPER:
        assert "nordwhisper" in server_json, "The server does not support Nordwhisper"

    if group_id == GROUP_DEDICATED_IP:
        assert "Dedicated IP" in server_json, "The server does not support DIP"
    if group_id == GROUP_OBFUSCATED_SERVERS:
        assert "Obfuscated" in server_json, "The server does not support obfuscation"
