import os
import re

import sh

from . import Port, Protocol, daemon, logging, dns

IP_ROUTE_TABLE = 205
ENDPOINTS = "endpoints"

# Rules for killswitch
# mangle
# -A PREROUTING -i {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP
# -A POSTROUTING -o {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP

# Rules for firewall
# mangle
# -A PREROUTING -i {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT
# -A PREROUTING -i {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP
# -A POSTROUTING -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A POSTROUTING -o {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -m comment --comment nordvpn -j ACCEPT
# -A POSTROUTING -o {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP

# Rules for allowlisted subnet
# mangle
# -A PREROUTING -s {subnet_ip} -i {iface} -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT
# -A PREROUTING -i {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT
# -A PREROUTING -i {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP
# -A POSTROUTING -d {subnet_ip} -o {iface} -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT
# -A POSTROUTING -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A POSTROUTING -o {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT
# -A POSTROUTING -o {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP

# Rules for allowlisted port
# mangle
# -A PREROUTING -i {iface} -p udp -m udp --dport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_udp -j ACCEPT
# -A PREROUTING -i {iface} -p udp -m udp --sport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_udp -j ACCEPT
# -A PREROUTING -i {iface} -p tcp -m tcp --dport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_tcp -j ACCEPT
# -A PREROUTING -i {iface} -p tcp -m tcp --sport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_tcp -j ACCEPT
# -A PREROUTING -i {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT
# -A PREROUTING -i {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP
# -A OUTPUT -p udp -m udp --sport {port} -m comment --comment nordvpn_allowlist -j MARK --set-xmark 0xe1f1/0xffffffff
# -A OUTPUT -p tcp -m tcp --sport {port} -m comment --comment nordvpn_allowlist -j MARK --set-xmark 0xe1f1/0xffffffff
# -A POSTROUTING -o {iface} -m mark --mark 0xe1f1 m comment --comment nordvpn -m comment --comment api_allowlist -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A POSTROUTING -o {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT     p
# -A POSTROUTING -o {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP

# Rules for allowlisted ports range
# mangle
# -A PREROUTING -i {iface} -p udp -m udp --dport {port_start}:{port_end} -m comment --comment nordvpn -m comment --comment allowlist_ports_udp -j ACCEPT
# -A PREROUTING -i {iface} -p udp -m udp --sport {port_start}:{port_end} -m comment --comment nordvpn -m comment --comment allowlist_ports_udp -j ACCEPT
# -A PREROUTING -i {iface} -p tcp -m tcp --dport {port_start}:{port_end} -m comment --comment nordvpn -m comment --comment allowlist_ports_tcp -j ACCEPT
# -A PREROUTING -i {iface} -p tcp -m tcp --sport {port_start}:{port_end} -m comment --comment nordvpn -m comment --comment allowlist_ports_tcp -j ACCEPT
# -A PREROUTING -i {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT
# -A PREROUTING -i {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP
# -A OUTPUT -p udp -m udp --sport {port_start}:{port_end} -m comment --comment nordvpn_allowlist -j MARK --set-xmark 0xe1f1/0xffffffff
# -A OUTPUT -p tcp -m tcp --sport {port_start}:{port_end} -m comment --comment nordvpn_allowlist -j MARK --set-xmark 0xe1f1/0xffffffff
# -A POSTROUTING -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A POSTROUTING -o {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT
# -A POSTROUTING -o {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP

# Rules for allowlisted port and protocol
# mangle
# -A PREROUTING -i {iface} -p {protocol} -m {protocol} --dport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_{protocol} -j ACCEPT
# -A PREROUTING -i {iface} -p {protocol} -m {protocol} --sport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_{protocol} -j ACCEPT
# -A PREROUTING -i {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT
# -A PREROUTING -i {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP
# -A OUTPUT -p {protocol} -m {protocol} --sport {port} -m comment --comment nordvpn_allowlist -j MARK --set-xmark 0xe1f1/0xffffffff
# -A POSTROUTING -o {iface} -p {protocol} -m {protocol} --dport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_{protocol} -j ACCEPT
# -A POSTROUTING -o {iface} -p {protocol} -m {protocol} --sport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_{protocol} -j ACCEPT
# -A POSTROUTING -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A POSTROUTING -o {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT
# -A POSTROUTING -o {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP

# Rules for allowlisted ports range and protocol
# mangle
# -A PREROUTING -i {iface} -p {protocol} -m {protocol} --sport {port_start}:{port_end} -m comment --comment nordvpn -m comment --comment allowlist_ports_{protocol} -j ACCEPT
# -A PREROUTING -i {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT
# -A PREROUTING -i {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP
# -A OUTPUT -p {protocol} -m {protocol} --sport {port} -m comment --comment nordvpn_allowlist -j MARK --set-xmark 0xe1f1/0xffffffff
# -A POSTROUTING -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A POSTROUTING -o {iface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT
# -A POSTROUTING -o {iface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP

PREROUTING_LAN_DISCOVERY_RULES = [
    "-A PREROUTING -s 169.254.0.0/16 -i eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A PREROUTING -s 192.168.0.0/16 -i eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A PREROUTING -s 172.16.0.0/12 -i eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A PREROUTING -s 10.0.0.0/8 -i eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
]

POSTROUTING_LAN_DISCOVERY_RULES = [
    "-A POSTROUTING -d 169.254.0.0/16 -o eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A POSTROUTING -d 192.168.0.0/16 -o eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A POSTROUTING -d 172.16.0.0/12 -o eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
    "-A POSTROUTING -d 10.0.0.0/8 -o eth0 -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
]


def __rules_connmark_chain_prerouting(interfaces: list) -> list[str]:
    result = []
    rules = [
            "-A PREROUTING -i {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT",
            "-A PREROUTING -i {interface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP",
    ]

    for rule in rules:
        for interface in interfaces:
            result.append(rule.format(interface=interface))

    return result


def __rules_block_dns_port() -> list[str]:
    return [
        "-A OUTPUT -d 169.254.0.0/16 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -m comment --comment deny-private-dns -j DROP",
        "-A OUTPUT -d 169.254.0.0/16 -p udp -m udp --dport 53 -m comment --comment nordvpn -m comment --comment deny-private-dns -j DROP",
        "-A OUTPUT -d 192.168.0.0/16 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -m comment --comment deny-private-dns -j DROP",
        "-A OUTPUT -d 192.168.0.0/16 -p udp -m udp --dport 53 -m comment --comment nordvpn -m comment --comment deny-private-dns -j DROP",
        "-A OUTPUT -d 172.16.0.0/12 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -m comment --comment deny-private-dns -j DROP",
        "-A OUTPUT -d 172.16.0.0/12 -p udp -m udp --dport 53 -m comment --comment nordvpn -m comment --comment deny-private-dns -j DROP",
        "-A OUTPUT -d 10.0.0.0/8 -p tcp -m tcp --dport 53 -m comment --comment nordvpn -m comment --comment deny-private-dns -j DROP",
        "-A OUTPUT -d 10.0.0.0/8 -p udp -m udp --dport 53 -m comment --comment nordvpn -m comment --comment deny-private-dns -j DROP"
    ]


def __rules_connmark_chain_output(interfaces: list) -> list[str]:
    result = []
    rules = [
        "-A POSTROUTING -o {interface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
        "-A POSTROUTING -o {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -m comment --comment api_allowlist -j ACCEPT",
    ]
    """
    Need to separate preparing expected rules ordering due to that:
    -A PREROUTING -i eth1 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A PREROUTING -i eth0 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A PREROUTING -i eth1 -m comment --comment nordvpn -j DROP
    -A PREROUTING -i eth0 -m comment --comment nordvpn -j DROP
    -A POSTROUTING -o eth1 -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
    -A POSTROUTING -o eth1 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A POSTROUTING -o eth0 -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
    -A POSTROUTING -o eth0 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A POSTROUTING -o eth1 -m comment --comment nordvpn -j DROP
    -A POSTROUTING -o eth0 -m comment --comment nordvpn -j DROP
    """
    for interface in interfaces:
        for rule in rules:
            result.append(rule.format(interface=interface))

    for interface in interfaces:
        result.append(f"-A POSTROUTING -o {interface} -m comment --comment nordvpn -m comment --comment drop-IPv4 -j DROP")

    return result


def __rules_allowlist_subnet_chain_input(interfaces: list, subnets: list[str]) -> list[str]:
    result = []

    """
    Reverse is needed due to that:
    Expected rules:
    -A PREROUTING -s 192.168.1.1/32 -i eth0 -m comment --comment nordvpn -j ACCEPT
    -A PREROUTING -s 192.168.1.1/32 -i eth1 -m comment --comment nordvpn -j ACCEPT
    Current rules:
    -A PREROUTING -s 192.168.1.1/32 -i eth1 -m comment --comment nordvpn -j ACCEPT
    -A PREROUTING -s 192.168.1.1/32 -i eth0 -m comment --comment nordvpn -j ACCEPT
    """
    interfaces.reverse()

    for subnet in subnets:
        for interface in interfaces:
            result.append(
                f"-A PREROUTING -s {subnet} -i {interface} -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
            )
    result.reverse()  # reverse() is needed because we always insert our rules, so newest one is always on top
    return result


def __rules_allowlist_subnet_chain_output(interfaces: list, subnets: list[str]) -> list[str]:
    result = []

    for subnet in subnets:
        for interface in interfaces:
            result.append(
                f"-A POSTROUTING -d {subnet} -o {interface} -m comment --comment nordvpn -m comment --comment allowlist_subnets -j ACCEPT",
            )
    result.reverse()
    return result


def __rules_allowlist_port_chain_input(interfaces: list, ports_udp: list[Port], ports_tcp: list[Port]) -> list[str]:
    result = []

    udp_rules = [
        "-A PREROUTING -i {interface} -p udp -m udp --dport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_udp -j ACCEPT",
        "-A PREROUTING -i {interface} -p udp -m udp --sport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_udp -j ACCEPT",
    ]
    tcp_rules = [
        "-A PREROUTING -i {interface} -p tcp -m tcp --dport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_tcp -j ACCEPT",
        "-A PREROUTING -i {interface} -p tcp -m tcp --sport {port} -m comment --comment nordvpn -m comment --comment allowlist_ports_tcp -j ACCEPT",
    ]

    for interface in interfaces:
        for udp_rule in udp_rules:
            for port in ports_udp:
                result.append(
                    udp_rule.format(interface=interface, port=port.value),
                )

    for interface in interfaces:
        for tcp_rule in tcp_rules:
            for port in ports_tcp:
                result.append(
                    tcp_rule.format(interface=interface, port=port.value),
                )

    return result


def __rules_allowlist_port_chain_output(ports_udp: list[Port], ports_tcp: list[Port]) -> list[str]:
    result = []

    for port in ports_udp:
        result.extend(
            [
                f"-A OUTPUT -p udp -m udp --sport {port.value} -m comment --comment nordvpn_allowlist -j MARK --set-xmark 0xe1f1/0xffffffff",
            ]
        )
    for port in ports_tcp:
        result.extend(
            [
                f"-A OUTPUT -p tcp -m tcp --sport {port.value} -m comment --comment nordvpn_allowlist -j MARK --set-xmark 0xe1f1/0xffffffff",
            ]
        )

    return result


def _get_rules_killswitch_on(interfaces: list) -> list[str]:
    result = []
    # mangle table rules
    result.extend(__rules_connmark_chain_prerouting(interfaces))

    result.extend(__rules_connmark_chain_output(interfaces))
    # filter table rules
    result.extend(__rules_block_dns_port())

    return result


def _get_rules_connected_to_vpn_server(interfaces: list) -> list[str]:
    return _get_rules_killswitch_on(interfaces)


def _get_rules_allowlist_subnet_on(interfaces: list, subnets: list[str]) -> list[str]:
    result = []
    # mangle table rules
    result.extend(__rules_allowlist_subnet_chain_input(interfaces, subnets))
    result.extend(__rules_connmark_chain_prerouting(interfaces))

    result.extend(__rules_allowlist_subnet_chain_output(interfaces, subnets))
    result.extend(__rules_connmark_chain_output(interfaces))
    # filter table rules
    result.extend(__rules_block_dns_port())

    return result


def _get_rules_allowlist_port_on(interfaces: list, ports: list[Port]) -> list[str]:
    ports_udp: list[Port]
    ports_tcp: list[Port]
    ports_udp, ports_tcp = _sort_ports_by_protocol(ports)

    result = []
    # mangle table rules
    result.extend(__rules_allowlist_port_chain_input(interfaces, ports_udp, ports_tcp))
    result.extend(__rules_connmark_chain_prerouting(interfaces))

    result.extend(__rules_allowlist_port_chain_output(ports_udp, ports_tcp))

    result.extend(__rules_connmark_chain_output(interfaces))
    # filter table rules
    result.extend(__rules_block_dns_port())

    return result


def _get_rules_allowlist_subnet_and_port_on(interfaces: list, subnets: list[str], ports: list[Port]) -> list[str]:
    ports_udp, ports_tcp = _sort_ports_by_protocol(ports)

    result = []
    # mangle table rules
    result.extend(__rules_allowlist_port_chain_input(interfaces, ports_udp, ports_tcp))
    result.extend(__rules_allowlist_subnet_chain_input(interfaces, subnets))
    result.extend(__rules_connmark_chain_prerouting(interfaces))

    result.extend(__rules_allowlist_port_chain_output(ports_udp, ports_tcp))

    result.extend(__rules_allowlist_subnet_chain_output(interfaces, subnets))
    result.extend(__rules_connmark_chain_output(interfaces))
    # filter table rules
    result.extend(__rules_block_dns_port())

    return result


def _get_all_interfaces() -> list[str]:
    """
    Extract interface names from 'ip a show dynamic' output

    :returns: list of interface names that have dynamic configuration
    """
    # Run the command
    output = sh.ip("a", "show", "dynamic", "up")

    interfaces = []

    # Parse the output line by line
    for line in output:
        # Look for lines that contain interface definitions
        # These start with a number followed by a colon and the interface name
        if line and line[0].isdigit() and ": " in line:
            # Extract the interface name (between the colon and the colon or end of line)
            # Example -> 1: enp0s31f6: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc fq_codel state DOWN group 78760 qlen 1000
            parts = line.split(": ", 2)
            interface_name = parts[1].split(":", 1)[0].strip()
            interfaces.append(interface_name)

    return interfaces if interfaces else [sh.ip.route.show("default").split(None)[4]]


def _get_firewall_rules(ports: list[Port] | None = None, subnets: list[str] | None = None) -> list[str]:
    # Default route interface
    rules = []
    interfaces = _get_all_interfaces()

    print("Default gateway:", interfaces)

    interfaces.reverse()
    """
    Need reverse interfaces due to that
    Expected rules:
    -A PREROUTING -i eth0 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A PREROUTING -i eth1 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A PREROUTING -i eth0 -m comment --comment nordvpn -j DROP
    -A PREROUTING -i eth1 -m comment --comment nordvpn -j DROP
    -A POSTROUTING -o eth0 -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
    -A POSTROUTING -o eth1 -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
    -A POSTROUTING -o eth0 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A POSTROUTING -o eth1 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A POSTROUTING -o eth0 -m comment --comment nordvpn -j DROP
    -A POSTROUTING -o eth1 -m comment --comment nordvpn -j DROP

    Current rules:
    -A PREROUTING -i eth1 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A PREROUTING -i eth0 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A PREROUTING -i eth1 -m comment --comment nordvpn -j DROP
    -A PREROUTING -i eth0 -m comment --comment nordvpn -j DROP
    -A POSTROUTING -o eth1 -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
    -A POSTROUTING -o eth1 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A POSTROUTING -o eth0 -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
    -A POSTROUTING -o eth0 -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
    -A POSTROUTING -o eth1 -m comment --comment nordvpn -j DROP
    -A POSTROUTING -o eth0 -m comment --comment nordvpn -j DROP
    """
    # Disconnected & Kill Switch ON
    if not daemon.is_connected() and daemon.is_killswitch_on():
        rules = _get_rules_killswitch_on(interfaces)

    # Connected
    if not ports and not subnets:
        rules = _get_rules_connected_to_vpn_server(interfaces)

    # Connected & Subnet(s) and Port(s) allowlisted
    if subnets and ports:
        rules = _get_rules_allowlist_subnet_and_port_on(interfaces, subnets, ports)

    # Connected & Subnet(s) allowlisted
    if subnets and not ports:
        rules = _get_rules_allowlist_subnet_on(interfaces, subnets)

    # Connected & Port(s) allowlisted
    if ports:
        rules = _get_rules_allowlist_port_on(interfaces, ports)

    return rules


def is_active(ports: list[Port] | None = None, subnets: list[str] | None = None) -> bool:
    """Returns True when all expected rules are found in iptables, in matching order."""
    print(sh.ip.route())

    expected_rules = _get_firewall_rules(ports, subnets)
    print("\nExpected rules:")
    logging.log("\nExpected rules:")
    for rule in expected_rules:
        print(f"{rule}")
        logging.log(rule)

    current_rules = _get_iptables_rules()
    print("\nCurrent rules:")
    logging.log("\nCurrent rules:")
    for rule in current_rules:
        print(f"{rule}")
        logging.log(rule)

    print()
    print(sh.nordvpn.settings())
    return current_rules == expected_rules


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
