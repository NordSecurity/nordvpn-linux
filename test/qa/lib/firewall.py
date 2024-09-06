import os
import re

import sh

from . import Port, Protocol, daemon, logging

IP_ROUTE_TABLE = 205

# Rules for killswitch
# -A INPUT -i {iface} -m comment --comment nordvpn -j DROP
# -A FORWARD -o {iface} -m comment --comment nordvpn -j DROP
# -A OUTPUT -o {iface} -m comment --comment nordvpn -j DROP

# Rules for firewall
# -A INPUT -i {iface} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A INPUT -i {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m comment --comment nordvpn -j DROP
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m comment --comment nordvpn -j DROP

# Rules for allowlisted subnet
# -A INPUT -s {subnet_ip} -i {iface} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A INPUT -i {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m comment --comment nordvpn -j DROP
# -A FORWARD -d {subnet_ip} -o {iface} -m comment --comment nordvpn -j ACCEPT
# -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -m comment --comment nordvpn -j ACCEPT
# -A FORWARD -o {iface} -m comment --comment nordvpn -j DROP
# -A OUTPUT -d {subnet_ip} -o {iface} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m comment --comment nordvpn -j DROP

# Rules for allowlisted port
# -A INPUT -i {iface} -p udp -m udp --dport {port} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -p udp -m udp --sport {port} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -p tcp -m tcp --dport {port} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -p tcp -m tcp --sport {port} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A INPUT -i {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m comment --comment nordvpn -j DROP
# -A FORWARD -o {iface} -m comment --comment nordvpn -j DROP
# -A OUTPUT -o {iface} -p udp -m udp --dport {port} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -p udp -m udp --sport {port} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -p tcp -m tcp --dport {port} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -p tcp -m tcp --sport {port} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m comment --comment nordvpn -j DROP

# Rules for allowlisted ports range
# -A INPUT -i {iface} -p udp -m udp --dport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -p udp -m udp --sport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -p tcp -m tcp --dport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -p tcp -m tcp --sport {port} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A INPUT -i {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m comment --comment nordvpn -j DROP
# -A FORWARD -o {iface} -m comment --comment nordvpn -j DROP
# -A OUTPUT -o {iface} -p udp -m udp --dport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -p udp -m udp --sport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -p tcp -m tcp --dport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -p tcp -m tcp --sport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m comment --comment nordvpn -j DROP

# Rules for allowlisted port and protocol
# -A INPUT -i {iface} -p {protocol} -m {protocol} --dport {port} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -p {protocol} -m {protocol} --sport {port} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A INPUT -i {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m comment --comment nordvpn -j DROP
# -A FORWARD -o {iface} -m comment --comment nordvpn -j DROP
# -A OUTPUT -o {iface} -p {protocol} -m {protocol} --dport {port} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -p {protocol} -m {protocol} --sport {port} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m comment --comment nordvpn -j DROP

# Rules for allowlisted ports range and protocol
# -A INPUT -i {iface} -p {protocol} -m {protocol} --sport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A INPUT -i {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m comment --comment nordvpn -j DROP
# -A OUTPUT -o {iface} -p {protocol} -m {protocol} --dport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -p {protocol} -m {protocol} --sport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m comment --comment nordvpn -j DROP

INPUT_LAN_DISCOVERY_RULES = [
    "-A INPUT -s 169.254.0.0/16 -i eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A INPUT -s 192.168.0.0/16 -i eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A INPUT -s 172.16.0.0/12 -i eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A INPUT -s 10.0.0.0/8 -i eth0 -m comment --comment nordvpn -j ACCEPT",
]

FORWARD_LAN_DISCOVERY_RULES = [
    "-A FORWARD -d 169.254.0.0/16 -o eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A FORWARD -d 192.168.0.0/16 -o eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A FORWARD -d 172.16.0.0/12 -o eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A FORWARD -d 10.0.0.0/8 -o eth0 -m comment --comment nordvpn -j ACCEPT",
]

OUTPUT_LAN_DISCOVERY_RULES = [
    "-A OUTPUT -d 169.254.0.0/16 -o eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A OUTPUT -d 192.168.0.0/16 -o eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A OUTPUT -d 172.16.0.0/12 -o eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A OUTPUT -d 10.0.0.0/8 -o eth0 -m comment --comment nordvpn -j ACCEPT",
]


def __rules_connmark_chain_input(interface: str):
    return \
        [
            f"-A INPUT -i {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
            f"-A INPUT -i {interface} -m comment --comment nordvpn -j DROP",
        ]


def __rules_connmark_chain_forward(interface: str):
    return \
        [
            f"-A FORWARD -o {interface} -m comment --comment nordvpn -j DROP",
        ]


def __rules_connmark_chain_output(interface: str):
    return \
        [
            "-A OUTPUT -d 169.254.0.0/16 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
            "-A OUTPUT -d 169.254.0.0/16 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
            "-A OUTPUT -d 192.168.0.0/16 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
            "-A OUTPUT -d 192.168.0.0/16 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
            "-A OUTPUT -d 172.16.0.0/12 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
            "-A OUTPUT -d 172.16.0.0/12 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
            "-A OUTPUT -d 10.0.0.0/8 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP",
            "-A OUTPUT -d 10.0.0.0/8 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP",
            f"-A OUTPUT -o {interface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
            f"-A OUTPUT -o {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
            f"-A OUTPUT -o {interface} -m comment --comment nordvpn -j DROP"
        ]


def __rules_allowlist_subnet_chain_input(interface: str, subnets: list[str]):
    result = []

    for subnet in subnets:
        result += (f"-A INPUT -s {subnet} -i {interface} -m comment --comment nordvpn -j ACCEPT", )

    current_subnet_rules_input_chain = []

    fw_lines = os.popen("sudo iptables -S").read()

    for line in fw_lines.splitlines():
        if "INPUT" in line and "-s" in line:
            current_subnet_rules_input_chain.append(line)

    if current_subnet_rules_input_chain:
        return sort_list_by_other_list(result, current_subnet_rules_input_chain)
    return result


def __rules_allowlist_subnet_chain_forward(interface: str, subnets: list[str]):
    result = []

    for subnet in subnets:
        result += (f"-A FORWARD -d {subnet} -o {interface} -m comment --comment nordvpn -j ACCEPT", )

    result += (f"-A FORWARD -o {interface} -m comment --comment nordvpn -j DROP", )

    current_subnet_rules_forward_chain = []

    fw_lines = os.popen("sudo iptables -S").read()

    for line in fw_lines.splitlines():
        if "FORWARD" in line and ("-d" in line or "DROP" in line):
            current_subnet_rules_forward_chain.append(line)

    if len(current_subnet_rules_forward_chain) > len(result):
        return sort_list_by_other_list(result, current_subnet_rules_forward_chain)
    return result


def __rules_allowlist_subnet_chain_output(interface: str, subnets: list[str]):
    result = []

    for subnet in subnets:
        result += (f"-A OUTPUT -d {subnet} -o {interface} -m comment --comment nordvpn -j ACCEPT", )

    result += ("-A OUTPUT -d 169.254.0.0/16 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP", )
    result += ("-A OUTPUT -d 169.254.0.0/16 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP", )
    result += ("-A OUTPUT -d 192.168.0.0/16 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP", )
    result += ("-A OUTPUT -d 192.168.0.0/16 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP", )
    result += ("-A OUTPUT -d 172.16.0.0/12 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP", )
    result += ("-A OUTPUT -d 172.16.0.0/12 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP", )
    result += ("-A OUTPUT -d 10.0.0.0/8 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -j DROP", )
    result += ("-A OUTPUT -d 10.0.0.0/8 -p udp -m udp --dport 53 -m comment --comment nordvpn -j DROP", )

    current_subnet_rules_input_chain = []

    fw_lines = os.popen("sudo iptables -S").read()

    for line in fw_lines.splitlines():
        if "OUTPUT" in line and "-d" in line:
            current_subnet_rules_input_chain.append(line)

    if len(current_subnet_rules_input_chain) > len(result):
        return sort_list_by_other_list(result, current_subnet_rules_input_chain)
    return result


def __rules_allowlist_port_chain_input(interface: str, ports_udp: list[Port], ports_tcp: list[Port]):
    result = []

    for port in ports_udp:
        result.extend([
            f"-A INPUT -i {interface} -p udp -m udp --dport {port.value} -m comment --comment nordvpn -j ACCEPT",
            f"-A INPUT -i {interface} -p udp -m udp --sport {port.value} -m comment --comment nordvpn -j ACCEPT"
        ])
    for port in ports_tcp:
        result.extend([
            f"-A INPUT -i {interface} -p tcp -m tcp --dport {port.value} -m comment --comment nordvpn -j ACCEPT",
            f"-A INPUT -i {interface} -p tcp -m tcp --sport {port.value} -m comment --comment nordvpn -j ACCEPT"
        ])

    return result


def __rules_allowlist_port_chain_output(interface: str, ports_udp: list[Port], ports_tcp: list[Port]):
    result = []

    for port in ports_udp:
        result.extend([
            f"-A OUTPUT -o {interface} -p udp -m udp --dport {port.value} -m comment --comment nordvpn -j ACCEPT",
            f"-A OUTPUT -o {interface} -p udp -m udp --sport {port.value} -m comment --comment nordvpn -j ACCEPT",
        ])
    for port in ports_tcp:
        result.extend([
            f"-A OUTPUT -o {interface} -p tcp -m tcp --dport {port.value} -m comment --comment nordvpn -j ACCEPT",
            f"-A OUTPUT -o {interface} -p tcp -m tcp --sport {port.value} -m comment --comment nordvpn -j ACCEPT",
        ])

    return result


def _get_rules_killswitch_on(interface: str):
    result = []

    result.extend(__rules_connmark_chain_input(interface))

    result.extend(__rules_connmark_chain_forward(interface))

    result.extend(__rules_connmark_chain_output(interface))

    return result


def _get_rules_connected_to_vpn_server(interface: str):
    return _get_rules_killswitch_on(interface)


def _get_rules_allowlist_subnet_on(interface: str, subnets: list[str]):
    result = []

    result.extend(__rules_allowlist_subnet_chain_input(interface, subnets))
    result.extend(__rules_connmark_chain_input(interface))

    result.extend(__rules_allowlist_subnet_chain_forward(interface, subnets))

    result.extend(__rules_allowlist_subnet_chain_output(interface, subnets))
    result.extend(__rules_connmark_chain_output(interface))

    return result


def _get_rules_allowlist_port_on(interface: str, ports: list[Port]):
    ports_udp: list[Port]
    ports_tcp: list[Port]
    ports_udp, ports_tcp = _sort_ports_by_protocol(ports)

    result = []

    result.extend(__rules_allowlist_port_chain_input(interface, ports_udp, ports_tcp))
    result.extend(__rules_connmark_chain_input(interface))

    result.extend(__rules_allowlist_port_chain_output(interface, ports_udp, ports_tcp))
    result.extend(__rules_connmark_chain_output(interface))

    return result


def _get_rules_allowlist_subnet_and_port_on(interface: str, subnets: list[str], ports: list[Port]):
    ports_udp, ports_tcp = _sort_ports_by_protocol(ports)

    result = []

    result.extend(__rules_allowlist_port_chain_input(interface, ports_udp, ports_tcp))
    result.extend(__rules_allowlist_subnet_chain_input(interface, subnets))
    result.extend(__rules_connmark_chain_input(interface))

    result.extend(__rules_allowlist_subnet_chain_forward(interface, subnets))

    result.extend(__rules_allowlist_port_chain_output(interface, ports_udp, ports_tcp))
    result.extend(__rules_allowlist_subnet_chain_output(interface, subnets))
    result.extend(__rules_connmark_chain_output(interface))

    return result


# TODO: Add missing IPv6 rules (icmp6 & dhcp6)
def _get_firewall_rules(ports: list[Port] | None = None, subnets: list[str] | None = None) -> list[str]:
    # Default route interface
    interface = sh.ip.route.show("default").split(None)[4]

    print("Default gateway:", interface)

    # Disconnected & Kill Switch ON
    if not daemon.is_connected() and daemon.is_killswitch_on():
        return _get_rules_killswitch_on(interface)

    # Connected
    if not ports and not subnets:
        return _get_rules_connected_to_vpn_server(interface)

    # Connected & Subnet(s) and Port(s) allowlisted
    if subnets and ports:
        return _get_rules_allowlist_subnet_and_port_on(interface, subnets, ports)

    # Connected & Subnet(s) allowlisted
    if subnets and not ports:
        return _get_rules_allowlist_subnet_on(interface, subnets)

    # Connected & Port(s) allowlisted
    if ports:
        return _get_rules_allowlist_port_on(interface, ports)
    return []


def is_active(ports: list[Port] | None = None, subnets: list[str] | None = None) -> bool:
    """Returns True when all expected rules are found in iptables, in matching order."""
    print(sh.ip.route())

    expected_rules = _get_firewall_rules(ports, subnets)
    print("\nExpected rules:")
    logging.log("\nExpected rules:")
    for rule in expected_rules:
        print(rule)
        logging.log(rule)

    current_rules = _get_iptables_rules()
    print("\nCurrent rules:")
    logging.log("\nCurrent rules:")
    for rule in current_rules:
        print(rule)
        logging.log(rule)

    print()
    print(sh.nordvpn.settings())

    return all(ln in current_rules for ln in expected_rules)


def is_empty() -> bool:
    """Returns True when firewall does not have DROP rules."""
    # under snap, also on host, ignore docker rules
    return "DROP" not in os.popen("sudo iptables -S | grep -v DOCKER").read()


def _get_iptables_rules() -> list[str]:
    # TODO: add full ipv6 support, separate task #LVPN-3684
    print("Using iptables")
    fw_lines = os.popen("sudo iptables -S").read()
    return fw_lines.split('\n')[3:-1]


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
    ports_udp.sort(key=lambda x: [int(i) if i.isdigit() else i for i in re.split('(\\d+)', x.value)], reverse=True)
    ports_tcp.sort(key=lambda x: [int(i) if i.isdigit() else i for i in re.split('(\\d+)', x.value)], reverse=True)

    return ports_udp, ports_tcp


def sort_list_by_other_list(to_sort: list[str], sort_by: list[str]) -> list[str]:
    # Create a dictionary to store the order of rules in `sort_by`
    order_dict = {rule: index for index, rule in enumerate(sort_by)}

    # Sort `to_sort` based on the order in `sort_by`
    return sorted(to_sort, key=lambda rule: order_dict[rule])


def add_and_delete_random_route():
    """Adds a random route, and deletes it. If this is not used, exceptions happen in allowlist tests."""
    # cmd = sh.sudo.ip.route.add.default.via.bake("127.0.0.1")
    # cmd.table(IP_ROUTE_TABLE)
    os.popen(f"sudo ip route add default via 127.0.0.1 table {IP_ROUTE_TABLE}").read()
    # sh.sudo.ip.route.delete.default.table(IP_ROUTE_TABLE)
    os.popen(f"sudo ip route delete default table {IP_ROUTE_TABLE}").read()
