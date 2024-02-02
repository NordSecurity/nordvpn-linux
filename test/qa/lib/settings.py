import sh


def get_server_ip() -> str:
    """ returns str with IP Address of the server from `nordvpn status`, that NordVPN client is currently connected to """
    return sh.nordvpn.status().split('\n')[2].replace('IP: ', '')


def get_current_connection_protocol():
    """ returns str current connection protocol from `nordvpn settings` """
    current_protocol = sh.nordvpn("settings").split('\n')[1]

    if "UDP" in current_protocol:
        return "udp"
    elif "TCP" in current_protocol:
        return "tcp"
    else:
        return "nordlynx"


def get_is_obfuscated():
    """ returns True, if Obfuscate is enabled in application settings """
    return "Obfuscate: enabled" in sh.nordvpn.settings()


def is_meshnet_on():
    """ return True when Meshnet is enabled """
    try:
        return "Meshnet: enabled" in sh.nordvpn.settings()
    except sh.ErrorReturnCode:
        return False


def dns_visible_in_settings(dns: list) -> bool:
    """ return True, if DNS that were passed as parameter are visible in app settings """
    current_dns_settings = sh.nordvpn("settings").split('\n')[-3]

    return all(entry in current_dns_settings for entry in dns)


def get_is_tpl_enabled():
    """ returns True, if Threat Protection Lite is enabled in application settings """
    return "Threat Protection Lite: enabled" in sh.nordvpn.settings()


def get_is_notify_enabled():
    """ returns True, if Threat Protection Lite is enabled in application settings """
    return "Notify: enabled" in sh.nordvpn.settings()


def get_is_routing_enabled():
    """ returns True, if Routing is enabled in application settings """
    return "Routing: enabled" in sh.nordvpn.settings()
