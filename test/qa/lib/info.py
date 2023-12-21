import sh


def collect():
    """collect system information and return as multiline string"""
    link_layer_info = sh.sudo.ip.link()
    network_interface_info = sh.sudo.ip.addr()
    routing_info = sh.sudo.ip.route()
    firewall_info = sh.sudo.iptables("-S")
    nameserver_info = sh.sudo.cat("/etc/resolv.conf")

    # without `ww` we cannot see full process lines, as it is cut off early
    processes = sh.ps("-ef", "ww")

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
            "-------------------end of system-information--------------------",
        ]
    )
