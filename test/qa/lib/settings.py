import sh


class Settings:
    def __init__(self):
        output = sh.nordvpn("settings").strip(" \r-\n")

        self.settings={}
        previous_key=""
        for line in output.split("\n"):
            values = line.split(":")
            if len(values) == 2:
                previous_key = values[0].lower().strip()
                self.settings[previous_key] = values[1].strip()
            elif len(previous_key) > 0:
                # for allow list the values are on a different line
                self.settings[previous_key] += line.strip() + " "

    def get(self, key: str) -> str:
        key = key.lower()
        return self.settings.get(key, "")


def get_server_ip() -> str:
    """Returns str with IP Address of the server from `nordvpn status`, that NordVPN client is currently connected to."""
    return sh.nordvpn.status().split('\n')[3].replace('IP: ', '')


def get_current_connection_protocol():
    """Returns str current connection protocol from `nordvpn settings`."""
    settings = Settings()
    if settings.get("Technology") == "NORDLYNX":
        return "nordlynx"

    return settings.get("Protocol").lower()


def is_obfuscated_enabled():
    """Returns True, if Obfuscate is enabled in application settings."""
    return Settings().get("Obfuscate") == "enabled"


def is_meshnet_enabled():
    """Return True when Meshnet is enabled."""
    return Settings().get("Meshnet") == "enabled"


def dns_visible_in_settings(dns: list) -> bool:
    """Return True, if DNS that were passed as parameter are visible in app settings."""
    current_dns_settings = Settings().get("DNS")
    return all(entry in current_dns_settings for entry in dns)


def is_tpl_enabled():
    """Returns True, if Threat Protection Lite is enabled in application settings."""
    return Settings().get("Threat Protection Lite") == "enabled"


def is_notify_enabled():
    """Returns True, if Threat Protection Lite is enabled in application settings."""
    return Settings().get("Notify") == "enabled"


def is_routing_enabled():
    """Returns True, if Routing is enabled in application settings."""
    return Settings().get("Routing") == "enabled"


def is_autoconnect_enabled():
    """Returns True, if Auto-connect is enabled in application settings."""
    return Settings().get("Auto-connect") == "enabled"


def is_lan_discovery_enabled():
    """Returns True, if LAN Discovery is enabled in application settings."""
    return Settings().get("LAN Discovery") == "enabled"


def is_firewall_enabled():
    """Returns True, if Firewall is enabled in application settings."""
    return Settings().get("Firewall") == "enabled"


def is_dns_disabled():
    """Returns True, if DNS is disabled in application settings."""
    return Settings().get("DNS") == "disabled"


def are_analytics_enabled():
    """Returns True, if Analytics are enabled in application settings."""
    return Settings().get("Analytics") == "enabled"


def is_ipv6_enabled():
    """Returns True, if IPv6 is enabled in application settings."""
    return Settings().get("IPv6") == "enabled"


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
        "Notify: enabled" in settings and
        "Tray: enabled" in settings and
        "Auto-connect: disabled" in settings and
        "IPv6: disabled" in settings and
        "Meshnet: disabled" in settings and
        "DNS: disabled" in settings and
        "LAN Discovery: disabled" in settings and
        "Virtual Location: enabled" in settings
    )
