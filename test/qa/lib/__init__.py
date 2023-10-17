from enum import Enum
import os
import sh
import time

# Used for test parametrization, when the tested functionality does not work with obfuscated.
STANDARD_TECHNOLOGIES = [
    # technology, protocol, obfuscation,
    ("openvpn", "udp", "off"),
    ("openvpn", "tcp", "off"),
    ("nordlynx", "", ""),
]

# Used for test parametrization, when the same test has to be run for obfuscated technologies.
OBFUSCATED_TECHNOLOGIES = [
    # technology, protocol, obfuscation,
    ("openvpn", "udp", "on"),
    ("openvpn", "tcp", "on"),
]

# Used for test parametrization, when the tested functionality does not work with obfuscated.
OVPN_STANDARD_TECHNOLOGIES = [
    # technology, protocol, obfuscation,
    ("openvpn", "udp", "off"),
    ("openvpn", "tcp", "off"),
]

# Used for test parametrization, when the same test has to be run for all technologies.
TECHNOLOGIES = OBFUSCATED_TECHNOLOGIES + STANDARD_TECHNOLOGIES

TECHNOLOGIES_BASIC1 = [
    ("nordlynx", "", ""),
]
TECHNOLOGIES_BASIC2 = [
    ("openvpn", "udp", "off"),
]


# no obfuscated servers with ipv6 2021/05/24
TECHNOLOGIES_WITH_IPV6 = STANDARD_TECHNOLOGIES

# Used for test parametrization, when the same test has to be run for different threat protection lite settings.
THREAT_PROTECTION_LITE = [
    "on",
    "off",
]

IPV6 = [
    "on",
    "off",
]

# Used for test parametrization, when the same test has to be run for obfuscated technologies.
STANDARD_GROUPS = [
    "Africa_The_Middle_East_And_India",
    "Asia_Pacific",
    "The_Americas",
    "Europe",
]

# Used for test parametrization, when the tested functionality does not work with obfuscated.
ADDITIONAL_GROUPS = [
    #"Double_VPN", # TODO: LVPN-3699
    "Onion_Over_VPN",
    "Standard_VPN_Servers",
    "P2P",
]

# Used for test parametrization, when the tested functionality only works with non-obfuscated OPENVPN.
OVPN_GROUPS = [
    "Dedicated_IP"
]

# Used for test parametrization, when the tested functionality only works with obfuscated OPENVPN.
OVPN_OBFUSCATED_GROUPS = [
    "Obfuscated_Servers"
]

# Used for test parametrization, when the same test has to be run for different groups.
GROUPS = STANDARD_GROUPS + ADDITIONAL_GROUPS

# Used for test parametrization, when the same test has to be run for different countries.
COUNTRIES = [
    "Germany",
    "Netherlands",
    "United_States",
    "France",
]

# Used for test parametrization, when the same test has to be run for different countries.
COUNTRY_CODES = [
    "de",
    "nl",
    "us",
    "fr",
]

# Used for test parametrization, when the same test has to be run for different cities.
CITIES = [
    "Frankfurt",
    "Amsterdam",
    "New_York",
    "Paris",
]

# Used for testing, when specific server is offine.
#
# curl api.nordvpn.com/v1/servers\?limit=6000 -L | jq '[.[] | select((.ips | length) > 1)] | map(.hostname |= rtrimstr(".nordvpn.com")) | map(.hostname)'
IPV6_SERVERS = [
    "us9591", "us9592"
]

class Protocol(Enum):
    UDP = "UDP"
    TCP = "TCP"

    def __str__(self):
        return self.value


PROTOCOLS = [
    Protocol.UDP,
    Protocol.TCP,
]

# Used for test parametrization, when the same test has to be run for different subnets.
SUBNETS = [
    "192.168.1.1/32",
]

# Used for test parametrization, when the same test has to be run for different ports.
PORTS = [
    "22",
]

# Used for test parametrization, when the same test has to be run for different ports.
PORTS_RANGE = [
    "3000:3100",
]

# Used for integration test coverage
os.environ["GOCOVERDIR"] = os.environ["WORKDIR"] + "/" + os.environ["COVERDIR"]

# Implements context manager a.k.a. with block and executes command on exit if exception was thrown.
class ErrorDefer:
    def __init__(self, command: sh.Command):
        self.command = command

    def __enter__(self):
        pass

    def __exit__(self, exc_type, exc_value, traceback):
        if exc_type and exc_value and traceback:
            print(self.command())


# Implements context manager a.k.a. with block and executes command on exit.
class Defer:
    def __init__(self, command: sh.Command):
        self.command = command

    def __enter__(self):
        pass

    def __exit__(self, exc_type, exc_value, traceback):
        print(self.command())


def set_technology_and_protocol(tech, proto, obfuscation):
    """Allows setting technology, protocol and obfuscation regardless
    of whether it is already set or not.

    Tests do not break on reordering when using this."""
    if tech:
        try:
            print(sh.nordvpn.set.technology(tech))
        except sh.ErrorReturnCode_1 as ex:
            print("WARNING:", ex)

    if proto:
        try:
            print(sh.nordvpn.set.protocol(proto))
        except sh.ErrorReturnCode_1 as ex:
            print("WARNING:", ex)

    if obfuscation:
        try:
            print(sh.nordvpn.set.obfuscate(obfuscation))
        except sh.ErrorReturnCode_1 as ex:
            print("WARNING:", ex)


# Allows setting threat protection lite regardless of whether it is already set or not.
#
# Tests do not break on reordering when using this.
def set_threat_protection_lite(dns):
    try:
        print(sh.nordvpn.set.cybersec(dns))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def set_dns(dns):
    try:
        print(sh.nordvpn.set.dns(dns))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def set_ipv6(ipv6):
    try:
        print(sh.nordvpn.set.ipv6(ipv6))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def set_firewall(firewall):
    try:
        print(sh.nordvpn.set.firewall(firewall))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def set_routing(routing):
    try:
        print(sh.nordvpn.set.routing(routing))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def set_killswitch(killswitch):
    try:
        print(sh.nordvpn.set.killswitch(killswitch))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def set_notify(dns):
    try:
        print(sh.nordvpn.set.notify(dns))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def add_port_to_allowlist(port):
    try:
        print(sh.nordvpn.allowlist.add.port(port))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def add_ports_range_to_allowlist(ports):
    port_range = ports.split(":")
    try:
        print(sh.nordvpn.allowlist.add.ports(port_range[0], port_range[1]))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def add_port_and_protocol_to_allowlist(port, protocol):
    try:
        print(sh.nordvpn.allowlist.add.port(port, "protocol", protocol))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def add_subnet_to_allowlist(subnet_addr):
    try:
        print(sh.nordvpn.allowlist.add.subnet(subnet_addr))
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


def flush_allowlist():
    try:
        print(sh.nordvpn.allowlist.remove.all())
    except sh.ErrorReturnCode_1 as ex:
        print("WARNING:", ex)


# returns True when successfully connected
def is_connect_successful(output, name="", hostname=""):
    if name and hostname:
        return (
            f"Connecting to {name} ({hostname})"
            and f"You are connected to {name} ({hostname})!" in output
        )
    return "Connecting to" and "You are connected to" in output


# returns True when failed to connect
def is_connect_unsuccessful(exception):
    return (
        "The specified server does not exist." in str(exception.value)
        or "The specified server is not available at the moment or does not support your connection settings."
        in str(exception.value)
        or "You cannot connect to a group and set the group option at the same time."
        in str(exception.value)
        or "Something went wrong. Please try again. If the problem persists, contact our customer support."
        in str(exception.value)
        or "The specified group does not exist."
        in str(exception.value)
    )


# returns True when successfully disconnected
def is_disconnect_successful(output):
    return "You are disconnected from NordVPN" in output


# returns True when command is invalid
def is_invalid_command(command, exception):
    return f"Command '{command}' doesn't exist." in str(exception.value)


def poll(func, attempts: int = 3, sleep: float = 1.0):
    for _ in range(attempts):
        yield func()
        time.sleep(sleep)
