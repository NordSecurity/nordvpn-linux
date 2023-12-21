import dns.resolver

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
    """returns True when NordVPN app has not modified the DNS"""
    return all(os_address != address
               for os_address in dns.resolver.Resolver().nameservers
               for address in ALL_TEST_DNS_ADDRESSES)


def is_set_for(dns_set_in_app: list) -> bool:
    """returns True, if NordVPN application has successfully set and overriden DNS servers in Resolver"""

    # DNS Addresses set in Resolver:
    dns_set_in_os_addresses = dns.resolver.Resolver().nameservers

    # Make sure, that:
    # 1. All DNS from NordVPN app were successfully set in Resolver
    # 2. All DNS Addresses in Resolver were overriden with DNS from NordVPN app
    return sorted(dns_set_in_app) == sorted(dns_set_in_os_addresses)
