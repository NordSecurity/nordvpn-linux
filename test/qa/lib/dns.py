import random

import dns.resolver
import sh

# Used for test parametrization.
DNS_NORD = ["103.86.96.100", "103.86.99.100"]
DNS_NORD_IPV6 = ["2400:bb40:4444::100", "2400:bb40:8888::100"]

# Used for test parametrization.
DNS_TPL = ["103.86.96.96", "103.86.99.99"]
DNS_TPL_IPV6 = ["2400:bb40:4444::103", "2400:bb40:8888::103"]

# Used for test parametrization, when the same test has to be run with different Threat Protection Lite alias.
TPL_ALIAS = [
    "threatprotectionlite",
    "tplite",
    "tpl",
    "cybersec"
]

# Used for test parametrization, when the same test has to be run for different values of custom dns.
DNS_CASE_CUSTOM_SINGLE = "2.0.0.0"
DNS_CASE_CUSTOM_DOUBLE = "2.0.0.1 2.0.0.2"
DNS_CASE_CUSTOM_TRIPLE = "2.0.0.3 2.0.0.4 2.0.0.5"
DNS_CASES_CUSTOM = [DNS_CASE_CUSTOM_SINGLE, DNS_CASE_CUSTOM_DOUBLE, DNS_CASE_CUSTOM_TRIPLE]

ALL_TEST_DNS_ADDRESSES = \
    DNS_NORD + \
    DNS_NORD_IPV6 + \
    DNS_TPL + \
    DNS_TPL_IPV6 + \
    DNS_CASE_CUSTOM_SINGLE.split(" ") + \
    DNS_CASE_CUSTOM_DOUBLE.split(" ") + \
    DNS_CASE_CUSTOM_TRIPLE.split(" ")

# Used for DNS test parametrization
DNS_CASES_ERROR = [
    ("a", "The provided IP address is invalid."),
    (["1.1.1.1", "1.1.1.1", "1.1.1.1", "1.1.1.1"], "More than 3 DNS addresses provided.")
]

# Used to check if error messages are correct
DNS_MSG_ERROR_ALREADY_SET = "DNS is already set to %s."
DNS_MSG_ERROR_ALREADY_DISABLED = "DNS is already set to disabled."

TPL_MSG_WARNING_DISABLING = "Disabling Threat Protection Lite."


def is_unset() -> bool:
    """Returns True when NordVPN app has not modified the DNS."""
    return all(os_address != address
               for os_address in dns.resolver.Resolver().nameservers
               for address in ALL_TEST_DNS_ADDRESSES)


def is_set_for(dns_set_in_app: list) -> bool:
    """Returns True, if NordVPN application has successfully set and overriden DNS servers in Resolver."""

    dns_set_in_os_addresses = get_dns_servers()

    return all(item in dns_set_in_os_addresses for item in dns_set_in_app)


# get list of dns servers for all/any interfaces
def get_dns_servers():
    dns_status = ""
    try:
        dns_status = sh.resolvectl("status")
    except sh.ErrorReturnCode_1:
        dns_status = ""

    if dns_status != "":
        servers = []
        for line in dns_status:
            if "DNS Servers" in line:
                for item in line.strip().split(":")[1].strip().split(" "):
                    servers.append(item)
        return servers
    return dns.resolver.Resolver().nameservers


def get_tpl_alias() -> str:
    """
    This function randomly picks an alias from the predefined list 'TPL_ALIAS' and returns it.

    Returns:
        str: A randomly selected alias from TPL_ALIAS.
    """
    return random.choice(TPL_ALIAS)
