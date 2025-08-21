import os

import sh
import requests


def fetch(url: str) -> str:
    try:
        resp = requests.get(url, timeout=2)
        resp.raise_for_status()  # raise if not 2xx
        return resp.text
    except requests.RequestException as e:
        return f"{e}"


def collect():
    """Collect system information and return as multiline string."""
    link_layer_info = os.popen("sudo ip link").read() #sh.sudo.ip.link()
    network_interface_info = os.popen("sudo ip addr").read() #sh.sudo.ip.addr()
    routing_info = os.popen("sudo ip route").read() #sh.sudo.ip.route()
    firewall_info = os.popen("sudo iptables -S").read() #sh.sudo.iptables("-S")
    nameserver_info = os.popen("sudo cat /etc/resolv.conf").read() #sh.sudo.cat("/etc/resolv.conf")

    # without `ww` we cannot see full process lines, as it is cut off early
    processes = sh.ps("-ef", "ww")
    goroutines = fetch("http://localhost:6960/debug/pprof/goroutine?debug=1")
    ls = os.popen("sudo ls -lah /var/lib/nordvpn/data/").read()

    return "\n".join(
        [
            "------------------start-of-system-information-------------------",
            "Link Layer:",
            str(link_layer_info),
            "Network Interfaces:",
            str(network_interface_info),
            "Routing:",
            str(routing_info),
            "Firewall:",
            str(firewall_info),
            "DNS:",
            str(nameserver_info),
            "Processes:",
            str(processes),
            "Goroutines",
            goroutines,
            ls,
            "-------------------end of system-information--------------------",
        ]
    )
