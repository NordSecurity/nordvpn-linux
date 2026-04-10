import os
import re
import socket

import sh

from . import Port, Protocol, logging, dns

IP_ROUTE_TABLE = 205
ENDPOINTS = "endpoints"
SOCK_TIMEOUT = 5
TCP_DST_PORT = 1234
UDP_DST_PORT = 1235

PREROUTING_LAN_DISCOVERY_RULES = [
    "-A PREROUTING -s 169.254.0.0/16 -i eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A PREROUTING -s 192.168.0.0/16 -i eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A PREROUTING -s 172.16.0.0/12 -i eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A PREROUTING -s 10.0.0.0/8 -i eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
]

LAN_DISCOVERY_SUBNETS = ["169.254.0.0/16", "192.168.0.0/16", "172.16.0.0/12", "10.0.0.0/8"]

POSTROUTING_LAN_DISCOVERY_RULES = [
    "-A POSTROUTING -d 169.254.0.0/16 -o eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A POSTROUTING -d 192.168.0.0/16 -o eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A POSTROUTING -d 172.16.0.0/12 -o eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A POSTROUTING -d 10.0.0.0/8 -o eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
]


def set_trace():
    # sh.sudo.nft("insert", "chain", "ip", "nat", "PREROUTING")
    # sh.sudo.nft("insert", "rule", "ip", "nat","DOCKER_OUTPUT", "meta", "nftrace", "set", "1")
    # sh.sudo.nft("monitor", "trace&", ">", "/opt/tracelog.log")
    # import time
    # time.sleep(30)
    pass


def is_active() -> bool:
    # change comment
    """Returns True when all expected rules are found in iptables, in matching order."""
    print(sh.ip.route())
    try:
        out = sh.sudo.nft("-a", "list", "ruleset")
    except:
        return False

    print(out)
    print(sh.nordvpn.settings())
    return "nordvpn" in out


tun_interface_names = ["nordtun", "qtun", "nordlynx"]


def is_active_subnet(subnets: list[str]) -> bool:
    for subnet in subnets:
        print(sh.ip.route.get(subnet))
        # Allowlisted subnet should not return tunnel interface name when using ip route get
        return not any(iface_name in sh.ip.route.get(subnet) for iface_name in tun_interface_names)


# rename better
def is_source_port_reachable(ports: list[Port]) -> bool:
    for port in ports:
        # Given port is a range `3000:3100`, in such a case we wish to test both ends of the range
        if ":" in port.value:
            port_range_start, port_range_end = port.value.split(":")
            return process_port(Port(port_range_start, port.protocol)) and process_port(Port(port_range_end, port.protocol))
        return process_port(port)


def process_port(port: Port) -> bool:
    if port.protocol == Protocol.TCP:
        return is_port_accessible_TCP(int(port.value))
    return is_port_accessible_UDP(int(port.value))


def is_port_accessible_TCP(src_port: int) -> bool:
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind(("0.0.0.0", src_port))
    s.settimeout(SOCK_TIMEOUT)
    try:
        s.connect(("172.19.0.1", TCP_DST_PORT))
        s.send(b"ping")
        print("TCP data sent")
        data = s.recv(4096)
        print("data received: ", data)
        return True
    except socket.timeout:
        print(
            f"timeout of {SOCK_TIMEOUT} hit with TCP, source port {src_port}, dst port : {TCP_DST_PORT}",
        )
        return False
    # `OSError: [Errno 99] Cannot assign requested address` thrown when unable to connect
    except OSError as e:
        print(e)
        return False
    finally:
        s.close()


def is_port_accessible_UDP(src_port):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind(("0.0.0.0", src_port))
    s.settimeout(SOCK_TIMEOUT)
    try:
        s.sendto(b"ping", ("172.19.0.1", UDP_DST_PORT))
        print("UDP data sent", UDP_DST_PORT, src_port)
        data, addr = s.recvfrom(4096)
        print("data received back from server: ", UDP_DST_PORT, data, addr)
        return True
    except PermissionError:
        print("unable to send packet to address")
        return False
    except socket.timeout:
        print(
            f"timeout of {SOCK_TIMEOUT} hit with UDP, source port {src_port}, dst port : {UDP_DST_PORT}",
        )
        return False
    finally:
        s.close()


def is_empty() -> bool:
    """Returns True when firewall does not have DROP rules."""
    # under snap, also on host, ignore docker rules
    rules = os.popen("sudo iptables -S | grep -v DOCKER").read()
    result = "DROP" not in rules
    if not result:
        logging.log(data=f"firewall.is_empty rules: {rules}")
    return result


def _get_iptables_rules() -> list[str]:
    print("Using iptables")

    mangle_fw_lines = os.popen("sudo iptables -S -t mangle").read()
    mangle_fw_list = mangle_fw_lines.split("\n")[5:-1]

    filter_fw_lines = os.popen("sudo iptables -S -t filter").read()
    filter_fw_list = filter_fw_lines.split("\n")[3:-1]
    fw_list = mangle_fw_list + filter_fw_list

    dns_full = dns.DNS_NORD + dns.DNS_TPL

    return [rule for rule in fw_list if not any(dns in rule for dns in dns_full)]


def _sort_ports_by_protocol(ports: list[Port]) -> tuple[list[Port], list[Port]]:
    """Sorts a list of ports and their corresponding protocols into UDP and TCP, both in descending order."""

    ports_udp: list[Port] = []
    ports_tcp: list[Port] = []

    for port in ports:
        if port.protocol == Protocol.UDP:
            ports_udp.append(port)
        elif port.protocol == Protocol.TCP:
            ports_tcp.append(port)
        else:
            ports_udp.append(port)
            ports_tcp.append(port)

    # Sort lists in descending order, since app sort rules like this in iptables
    ports_udp.sort(key=lambda x: [int(i) if i.isdigit() else i for i in re.split("(\\d+)", x.value)], reverse=True)
    ports_tcp.sort(key=lambda x: [int(i) if i.isdigit() else i for i in re.split("(\\d+)", x.value)], reverse=True)

    return ports_udp, ports_tcp


def add_and_delete_random_route():
    """Adds a random route, and deletes it. If this is not used, exceptions happen in allowlist tests."""
    # cmd = sh.sudo.ip.route.add.default.via.bake("127.0.0.1")
    # cmd.table(IP_ROUTE_TABLE)
    os.popen(f"sudo ip route add default via 127.0.0.1 table {IP_ROUTE_TABLE}").read()
    # sh.sudo.ip.route.delete.default.table(IP_ROUTE_TABLE)
    os.popen(f"sudo ip route delete default table {IP_ROUTE_TABLE}").read()
