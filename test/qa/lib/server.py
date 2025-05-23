import logging
import time
from urllib.parse import quote

import requests
from . import shell


class ServerInfo:
    def __init__(self, server_info):
        self.name = server_info["name"]
        self.hostname = server_info["hostname"]
        self.city = server_info["locations"][0]["country"]["city"]["name"]
        self.country = server_info["locations"][0]["country"]["name"]


def get_hostname_by(technology="", protocol="", obfuscated="", group_id=""):
    """Returns server name and hostname from core API."""
    tech_id = ""

    if technology != "":
        tech_ids = {
            "openvpn": {
                "udp": {
                    "on": 15,
                    "off": 3,
                },
                "tcp": {
                    "on": 17,
                    "off": 5,
                },
            },
            "nordlynx": 35,
            "nordwhisper": 51,
        }
        if protocol != "":
            tech_id = tech_ids[technology][protocol][f"{obfuscated}"]
        else:
            tech_id = tech_ids[technology]

    group_ids = {
        "Double_VPN": "1",
        "Onion_Over_VPN": "3",
        "Dedicated_IP": "9",
        "Standard_VPN_Servers": "11",
        "P2P": "15",
        "Obfuscated_Servers": "17",
        "Europe": "19",
        "The_Americas": "21",
        "Asia_Pacific": "23",
        "Africa_The_Middle_East_And_India": "25",
    }

    if group_id != "":
        group_id = group_ids[group_id]

    # api limits
    time.sleep(2)
    url = f"https://api.nordvpn.com/v1/servers?limit=10&filters[servers.status]=online&filters[servers_technologies]={tech_id}&filters[servers_groups]={group_id}"
    logging.debug(url)
    server_info = requests.get(url, timeout=5).json()[0]
    return ServerInfo(server_info=server_info)


def get_server_info(server_name):
    server_name = quote(server_name)
    url = f"https://api.nordvpn.com/v1/servers?filters[servers.name]={server_name}&fields[servers.locations]"
    server_info = requests.get(url, timeout=5).json()

    city = server_info[0]["locations"][0]["country"]["city"]["name"]
    country = server_info[0]["locations"][0]["country"]["name"]

    return city, country

# TODO: LVPN-7744
def get_dedicated_ip() -> None | str:
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

    return str(response.json()[0]['hostname'].split(".")[0])
