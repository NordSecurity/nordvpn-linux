import random

import sh

from . import UserConsentMode


MSG_AUTOCONNECT_ENABLE_SUCCESS = "Auto-connect is set to 'enabled' successfully."
MSG_AUTOCONNECT_DISABLE_SUCCESS = "Auto-connect is set to 'disabled' successfully."
MSG_AUTOCONNECT_DISABLE_FAIL = "Auto-connect is already set to 'disabled'."

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

MSG_SET_DEFAULTS = "Settings were successfully restored to defaults."

# Used for test parametrization, when the same test has to be run with different Post-quantum VPN alias.
PQ_ALIAS = [
    "post-quantum",
    "pq"
]

def get_pq_alias() -> str:
    """
    This function randomly picks an alias from the predefined list 'PQ_ALIAS' and returns it.

    Returns:
        str: A randomly selected alias from PQ_ALIAS.
    """
    return random.choice(PQ_ALIAS)


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


def is_user_consent_granted():
    """
    Returns True, if User Consent is enabled, False if it's disabled.

    If the consent was not declared. It raises an exception.
    """
    user_consent = Settings().get("user consent")
    if user_consent == UserConsentMode.ENABLED:
        return True

    if user_consent == UserConsentMode.DISABLED:
        return False

    raise Exception("user consent is undefined")


def is_user_consent_declared():
    """Returns True, if User Consent is enabled or disabled, False if it is undefined in application settings."""
    return Settings().get("user consent") != UserConsentMode.UNDEFINED


def is_virtual_location_enabled():
    """Returns True, if Virtual Location is enabled in application settings."""
    return Settings().get("Virtual Location") == "enabled"


def is_post_quantum_disabled():
    """Returns True, if Post-quantum VPN is disabled in application settings."""
    return Settings().get("Post-quantum VPN") == "disabled"


def app_has_defaults_settings():
    """Returns True, if application settings match the default settings."""
    settings = sh.nordvpn.settings()
    return (
        "Technology: NORDLYNX" in settings and
        "Firewall: enabled" in settings and
        "Firewall Mark: 0xe1f1" in settings and
        "Routing: enabled" in settings and
        # User Consent is not restored to default on reset
        ("User Consent: enabled" in settings or "User Consent: disabled" in settings) and
        "Kill Switch: disabled" in settings and
        "Threat Protection Lite: disabled" in settings and
        "Notify: enabled" in settings and
        "Tray: enabled" in settings and
        "Auto-connect: disabled" in settings and
        "Meshnet: disabled" in settings and
        "DNS: disabled" in settings and
        "LAN Discovery: disabled" in settings and
        "Virtual Location: enabled" in settings and
        "Post-quantum VPN: disabled" in settings

    )
