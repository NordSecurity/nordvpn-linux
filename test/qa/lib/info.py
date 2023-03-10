import sh


def collect():
    """collect system information and return as multiline string"""
    link_layer_info = sh.sudo.ip.link()
    network_interface_info = sh.sudo.ip.addr()
    rounting_info = sh.sudo.ip.route()
    firewall_info = sh.sudo.iptables("-S")
    nameserver_info = sh.sudo.cat("/etc/resolv.conf")
    processes = sh.ps("-ef")

    return "\n".join(
        [
            "------------------start-of-system-information-------------------",
            "Link Layer:",
            str(link_layer_info),
            "Network Interfaces:",
            str(network_interface_info),
            "Routing:",
            str(rounting_info),
            "Firewall:",
            str(firewall_info),
            "DNS:",
            str(nameserver_info),
            "Processes:",
            str(processes),
            "-------------------end of system-information--------------------",
        ]
    )
