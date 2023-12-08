from lib import daemon, logging
import lib
import re
import sh

IP_ROUTE_TABLE = 205

# Rules for killswitch
# -A INPUT -i {iface} -m comment --comment nordvpn -j DROP
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

inputLanDiscoveryRules = [
    "-A INPUT -s 169.254.0.0/16 -i eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A INPUT -s 192.168.0.0/16 -i eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A INPUT -s 172.16.0.0/12 -i eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A INPUT -s 10.0.0.0/8 -i eth0 -m comment --comment nordvpn -j ACCEPT",
]

outputLanDiscoveryRules = [
    "-A OUTPUT -d 169.254.0.0/16 -o eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A OUTPUT -d 192.168.0.0/16 -o eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A OUTPUT -d 172.16.0.0/12 -o eth0 -m comment --comment nordvpn -j ACCEPT",
    "-A OUTPUT -d 10.0.0.0/8 -o eth0 -m comment --comment nordvpn -j ACCEPT",
]


def _get_rules_killswitch_on(interface:str):
    return \
    [
        f"-A INPUT -i {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
        f"-A INPUT -i {interface} -m comment --comment nordvpn -j DROP",
        f"-A OUTPUT -o {interface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
        f"-A OUTPUT -o {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
        f"-A OUTPUT -o {interface} -m comment --comment nordvpn -j DROP"
    ]


def _get_rules_connected_to_vpn_server(interface:str):
    return _get_rules_killswitch_on(interface)


def _get_rules_allowlist_subnet_on(interface:str, subnets:list[str]):
    # Subnet allowlist rules
    result = []

    for subnet in subnets:
        result += f"-A INPUT -s {subnet} -i {interface} -m comment --comment nordvpn -j ACCEPT",

    result += \
    [
        f"-A INPUT -i {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
        f"-A INPUT -i {interface} -m comment --comment nordvpn -j DROP",
    ]

    for subnet in subnets:
        result += f"-A OUTPUT -d {subnet} -o {interface} -m comment --comment nordvpn -j ACCEPT",

    result += \
    [
        f"-A OUTPUT -o {interface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
        f"-A OUTPUT -o {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
        f"-A OUTPUT -o {interface} -m comment --comment nordvpn -j DROP"
    ]
    return result


def _get_rules_allowlist_port_on(interface:str, ports:list[lib.Port]):
    # Port(range) whitelist rules for SPECIFIC protocols

    ports_udp: list[lib.Port]
    ports_tcp: list[lib.Port]
    ports_udp, ports_tcp = _sort_ports_by_protocol(ports)

    result = []

    for port in ports_udp:
        result.extend([
            f"-A INPUT -i {interface} -p udp -m udp --dport {port.value} -m comment --comment nordvpn -j ACCEPT",
            f"-A INPUT -i {interface} -p udp -m udp --sport {port.value} -m comment --comment nordvpn -j ACCEPT",
        ])
    for port in ports_tcp:
        result.extend([
            f"-A INPUT -i {interface} -p tcp -m tcp --dport {port.value} -m comment --comment nordvpn -j ACCEPT",
            f"-A INPUT -i {interface} -p tcp -m tcp --sport {port.value} -m comment --comment nordvpn -j ACCEPT"
        ])

    result += \
    [
        f"-A INPUT -i {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
        f"-A INPUT -i {interface} -m comment --comment nordvpn -j DROP",
    ]

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
    result += \
    [
        f"-A OUTPUT -o {interface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
        f"-A OUTPUT -o {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
        f"-A OUTPUT -o {interface} -m comment --comment nordvpn -j DROP"
    ]
    return result


def _get_rules_allowlist_subnet_and_port_on(interface:str, subnets:list[str], ports:list[lib.Port]):
    # Port(range) && subnet whitelist rules for ALL protocols

    ports_udp, ports_tcp = _sort_ports_by_protocol(ports)

    result = []

    for port in ports_udp:
        result.extend([
            f"-A INPUT -i {interface} -p udp -m udp --dport {port.value} -m comment --comment nordvpn -j ACCEPT",
            f"-A INPUT -i {interface} -p udp -m udp --sport {port.value} -m comment --comment nordvpn -j ACCEPT",
        ])
    for port in ports_tcp:
        result.extend([
            f"-A INPUT -i {interface} -p tcp -m tcp --dport {port.value} -m comment --comment nordvpn -j ACCEPT",
            f"-A INPUT -i {interface} -p tcp -m tcp --sport {port.value} -m comment --comment nordvpn -j ACCEPT"
        ])
    for subnet in subnets:
        result += f"-A INPUT -s {subnet} -i {interface} -m comment --comment nordvpn -j ACCEPT",

    result += \
    [
        f"-A INPUT -i {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
        f"-A INPUT -i {interface} -m comment --comment nordvpn -j DROP",
    ]

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
    for subnet in subnets:
        result += f"-A OUTPUT -d {subnet} -o {interface} -m comment --comment nordvpn -j ACCEPT",

    result += \
    [
        f"-A OUTPUT -o {interface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
        f"-A OUTPUT -o {interface} -m connmark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT",
        f"-A OUTPUT -o {interface} -m comment --comment nordvpn -j DROP"
    ]
    return result


# ToDo: Add missing IPv6 rules (icmp6 & dhcp6)
def _get_firewall_rules(ports: list[lib.Port]=None, subnets: list[str]=None) -> list[str]:
    if subnets:
        subnets.sort(reverse=True)

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


def is_active(ports: list[lib.Port]=None, subnets: list[str]=None) -> bool:
    """ returns True when all expected rules are found in iptables, in matching order """
    print(sh.ip.route())

    expected_rules = _get_firewall_rules(ports, subnets)
    print("\nExpected rules:")
    for rule in expected_rules:
        print(rule)

    current_rules = _get_iptables_rules()
    print("\nCurrent rules:")
    for rule in current_rules:
        print(rule)

    print()
    print(sh.nordvpn.settings())

    return current_rules == expected_rules


def is_empty() -> bool:
    """ returns True when firewall does not have DROP rules """
    return "DROP" not in sh.sudo.iptables("-S")


def _get_iptables_rules() -> list[str]:
    # TODO: add full ipv6 support, separate task #LVPN-3684
    print("Using iptables")
    return sh.sudo.iptables("-S").split('\n')[3:-1]


def _sort_ports_by_protocol(ports: list[lib.Port]) -> tuple[list[lib.Port], list[lib.Port]]:
    """ Sorts a list of ports and their corresponding protocols into UDP and TCP, both in descending order. """

    ports_udp: list[lib.Port] = []
    ports_tcp: list[lib.Port] = []

    for port in ports:
        if port.protocol == lib.Protocol.UDP:
            ports_udp.append(port)
        elif port.protocol == lib.Protocol.TCP:
            ports_tcp.append(port)
        else:
            ports_udp.append(port)
            ports_tcp.append(port)

    # Sort lists in descending order
    ports_udp.sort(key=lambda x: [int(i) if i.isdigit() else i for i in re.split('(\\d+)', x.value)], reverse=True)
    ports_tcp.sort(key=lambda x: [int(i) if i.isdigit() else i for i in re.split('(\\d+)', x.value)], reverse=True)

    return ports_udp, ports_tcp
