import logging
import random
import requests
import time

# get server name and hostname from core API
def get_hostname_by(technology, protocol, obfuscated):
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

    # api limits
    time.sleep(2)
    url = f"https://api.nordvpn.com/v1/servers?limit=10&filters[servers.status]=online&filters[servers_technologies]={tech_id}"
    logging.debug(url)
    server = requests.get(url).json()[random.randint(0, 9)]
    return server["name"], server["hostname"]
