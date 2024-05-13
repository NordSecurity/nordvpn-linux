import sh


def get_server_ip() -> str:
    """Returns str with IP Address of the server from `nordvpn status`, that NordVPN client is currently connected to."""
    return sh.nordvpn.status().split('\n')[3].replace('IP: ', '')


def get_current_connection_protocol():
    """Returns str current connection protocol from `nordvpn settings`."""
    current_protocol = sh.nordvpn("settings").split('\n')[1]

    if "UDP" in current_protocol:
        return "udp"
    elif "TCP" in current_protocol:
        return "tcp"
    else:
        return "nordlynx"


def is_obfuscated_enabled():
    """Returns True, if Obfuscate is enabled in application settings."""
    return "Obfuscate: enabled" in sh.nordvpn.settings()


def is_meshnet_enabled():
    """Return True when Meshnet is enabled."""
    try:
        return "Meshnet: enabled" in sh.nordvpn.settings()
    except sh.ErrorReturnCode:
        return False


def dns_visible_in_settings(dns: list) -> bool:
    """Return True, if DNS that were passed as parameter are visible in app settings."""
    current_dns_settings = sh.nordvpn("settings").split('\n')[-3]

    return all(entry in current_dns_settings for entry in dns)


def is_tpl_enabled():
    """Returns True, if Threat Protection Lite is enabled in application settings."""
    return "Threat Protection Lite: enabled" in sh.nordvpn.settings()


def is_notify_enabled():
    """Returns True, if Threat Protection Lite is enabled in application settings."""
    return "Notify: enabled" in sh.nordvpn.settings()


def is_routing_enabled():
    """Returns True, if Routing is enabled in application settings."""
    return "Routing: enabled" in sh.nordvpn.settings()


def is_autoconnect_enabled():
    """Returns True, if Auto-connect is enabled in application settings."""
    return "Auto-connect: enabled" in sh.nordvpn.settings()


def is_lan_discovery_enabled():
    """Returns True, if LAN Discovery is enabled in application settings."""
    return "LAN Discovery: enabled" in sh.nordvpn.settings()


def is_firewall_enabled():
    """Returns True, if Firewall is enabled in application settings."""
    return "Firewall: enabled" in sh.nordvpn.settings()


def is_dns_disabled():
    """Returns True, if DNS is disabled in application settings."""
    return "DNS: disabled" in sh.nordvpn.settings()


def are_analytics_enabled():
    """Returns True, if Analytics are enabled in application settings."""
    return "Analytics: enabled" in sh.nordvpn.settings()


def is_ipv6_enabled():
    """Returns True, if IPv6 is enabled in application settings."""
    return "IPv6: enabled" in sh.nordvpn.settings()


def app_has_defaults_settings():
    """Returns True, if application settings match the default settings."""
    settings = sh.nordvpn.settings()
    return (
        "Technology: NORDLYNX" in settings and
        "Firewall: enabled" in settings and
        "Firewall Mark: 0xe1f1" in settings and
        "Routing: enabled" in settings and
        "Analytics: enabled" in settings and
        "Kill Switch: disabled" in settings and
        "Threat Protection Lite: disabled" in settings and
        "Notify: disabled" in settings and
        "Auto-connect: disabled" in settings and
        "IPv6: disabled" in settings and
        "Meshnet: disabled" in settings and
        "DNS: disabled" in settings and
        "LAN Discovery: disabled" in settings
    )
