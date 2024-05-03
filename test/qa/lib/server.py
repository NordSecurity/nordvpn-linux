import logging
import time
from urllib.parse import quote

import requests

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
        "Africa_The_Middle_East_And_India": "25"
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
