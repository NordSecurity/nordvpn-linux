from lib import daemon
import sh

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

# Rules for whitelisted subnet
# -A INPUT -s {subnet_ip} -i {iface} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A INPUT -i {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m comment --comment nordvpn -j DROP
# -A OUTPUT -d {subnet_ip} -o {iface} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m comment --comment nordvpn -j DROP

# Rules for whitelisted port
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

# Rules for whitelisted ports range
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

# Rules for whitelisted port and protocol
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

# Rules for whitelisted ports range and protocol
# -A INPUT -i {iface} -p {protocol} -m {protocol} --sport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A INPUT -i {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A INPUT -i {iface} -m comment --comment nordvpn -j DROP
# -A OUTPUT -o {iface} -p {protocol} -m {protocol} --dport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -p {protocol} -m {protocol} --sport {port_start}:{port_end} -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
# -A OUTPUT -o {iface} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
# -A OUTPUT -o {iface} -m comment --comment nordvpn -j DROP


# ToDo: Add missing IPv6 rules (icmp6 & dhcp6)
def _get_firewall_rules(killswitch, server_ip, iface, port="", protocol="", subnet=""):
    if killswitch == True and server_ip == "":
        return """-A INPUT -i {face} -m comment comment nordvpn -j DROP
-A OUTPUT -o {face} -m comment comment nordvpn -j DROP""".format(
            face=iface
        )

    if port == "" and protocol == "" and subnet == "":
        return """-A INPUT -i {face} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
-A INPUT -i {face} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -m comment --comment nordvpn -j DROP
-A OUTPUT -o {face} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
-A OUTPUT -o {face} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -m comment --comment nordvpn -j DROP""".format(
            ip=server_ip, face=iface
        )

    if port == "" and protocol == "":
        return """-A INPUT -s {subnet_addr} -i {face} -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
-A INPUT -i {face} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -m comment --comment nordvpn -j DROP
-A OUTPUT -d {subnet_addr} -o {face} -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
-A OUTPUT -o {face} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -m comment --comment nordvpn -j DROP""".format(
            ip=server_ip, face=iface, subnet_addr=subnet
        )

    if protocol == "":
        return """-A INPUT -i {face} -p udp -m udp --dport {p} -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -p udp -m udp --sport {p} -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -p tcp -m tcp --dport {p} -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -p tcp -m tcp --sport {p} -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
-A INPUT -i {face} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -m comment --comment nordvpn -j DROP
-A OUTPUT -o {face} -p udp -m udp --dport {p} -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -p udp -m udp --sport {p} -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -p tcp -m tcp --dport {p} -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -p tcp -m tcp --sport {p} -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
-A OUTPUT -o {face} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -m comment --comment nordvpn -j DROP""".format(
            ip=server_ip, face=iface, p=port
        )

    return """-A INPUT -i {face} -p {proto} -m {proto} --dport {p} -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -p {proto} -m {proto} --sport {p} -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -m connmark --mark 0xe1f1 -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff
-A INPUT -i {face} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
-A INPUT -i {face} -m comment --comment nordvpn -j DROP
-A OUTPUT -o {face} -p {proto} -m {proto} --dport {p} -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -p {proto} -m {proto} --sport {p} -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -m mark --mark 0xe1f1 -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff
-A OUTPUT -o {face} -m mark --mark 0xe1f1 -m comment --comment nordvpn -j ACCEPT
-A OUTPUT -o {face} -m comment --comment nordvpn -j DROP""".format(
        ip=server_ip, face=iface, p=port, proto=protocol.lower()
    )


def is_active(port="", protocol="", subnet=""):
    # Get interface name of your default gateway
    output = sh.ip.route.show("default")
    print(sh.ip.route.show())
    _, _, _, _, iface = output.split(None, 5)

    try:
        # Get VPN server's IP address
        status = sh.grep(sh.nordvpn.status(), "IP")
        _, _, server_ip = status.split(None, 3)
    except sh.ErrorReturnCode:
        server_ip = ""
    print("Default gateway:", iface, "Server's IP:  ", server_ip)

    rules = _get_firewall_rules(
        daemon.is_killswitch_on(), server_ip, iface, port, protocol, subnet
    )
    print("Expected rules:\n", rules)

    current_rules = _get_iptables_rules()
    print("Current rules:\n", current_rules)

    print(sh.nordvpn.settings())

    return rules in current_rules


def is_empty() -> bool:
    """returns True when firewall does not have DROP rules"""
    return "DROP" not in sh.sudo.iptables("-S")


def _get_iptables_rules():
    if daemon.is_ipv6_on():
        print("Using ip6tables")
        return sh.sudo.ip6tables("-S")

    print("Using iptables")
    return sh.sudo.iptables("-S")
