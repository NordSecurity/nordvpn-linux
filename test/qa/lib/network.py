import socket
import time
from itertools import cycle
from threading import Thread

import dns.resolver
import pytest
import requests
import sh

from . import daemon, firewall, info, logging, settings

# private variable for storing routes
_blackholes = []

API_EXTERNAL_IP = "https://api.nordvpn.com/v1/helpers/ips/insights"

TSHARK_FILTER_NORDLYNX = "(udp port 51820) and (ip dst %s)"
TSHARK_FILTER_UDP = "(udp port 1194) and (ip dst %s)"
TSHARK_FILTER_TCP = "(tcp port 443) and (ip dst %s)"
TSHARK_FILTER_UDP_OBFUSCATED = "udp and (port not 1194) and (ip dst %s)"
TSHARK_FILTER_TCP_OBFUSCATED = "tcp and (port not 443) and (ip dst %s)"

FWMARK = 57841

class PacketCaptureThread(Thread):
    def __init__(self, connection_settings):
        Thread.__init__(self)
        self.packets_captured: int = -1
        self.connection_settings = connection_settings

    def run(self):
        self.packets_captured = _capture_packets(self.connection_settings)


def _capture_packets(connection_settings: (str, str, str)) -> int:
    technology = connection_settings[0]
    protocol = connection_settings[1]
    obfuscated = connection_settings[2]

    # Collect information needed for tshark filter
    server_ip = settings.get_server_ip()

    # Choose traffic filter according to information collected above
    if technology == "nordlynx" and protocol == "" and obfuscated == "":
        traffic_filter = TSHARK_FILTER_NORDLYNX % server_ip
    elif technology == "openvpn" and protocol == "udp" and obfuscated == "off":
        traffic_filter = TSHARK_FILTER_UDP % server_ip
    elif technology == "openvpn" and protocol == "tcp" and obfuscated == "off":
        traffic_filter = TSHARK_FILTER_TCP % server_ip
    elif technology == "openvpn" and protocol == "udp" and obfuscated == "on":
        traffic_filter = TSHARK_FILTER_UDP_OBFUSCATED % server_ip
    elif technology == "openvpn" and protocol == "tcp" and obfuscated == "on":
        traffic_filter = TSHARK_FILTER_TCP_OBFUSCATED % server_ip

    # If enough packets are captured, do not wait the duration time, exit early, show compact output
    tshark_result: str = sh.tshark("-i", "any", "-T", "fields", "-e", "ip.src", "-e", "ip.dst", "-a", "duration:3", "-a", "packets:1", "-f", traffic_filter)
    #tshark_result: str = os.popen(f"sudo tshark -i any -T fields -e ip.src -e ip.dst -a duration:3 -a packets:1 -f {traffic_filter}").read()

    packets = tshark_result.strip().splitlines()

    return len(packets)


def measure_time(func):
    """A decorator that measures the execution time of a function and prints it."""
    def wrapper(*args, **kwargs):
        start = time.perf_counter()  # more precise for short durations
        result = func(*args, **kwargs)
        end = time.perf_counter()
        print(f"~~~Function {func.__name__} took {end - start:.6f} seconds to complete.")
        return result
    return wrapper


def capture_traffic(connection_settings) -> int:
    """Returns count of captured packets."""

    # We try to capture packets using other thread
    t_connect = PacketCaptureThread(connection_settings)
    t_connect.start()

    #sh.ping("-c", "2", "-w", "2", "1.1.1.1")
    _check_connectivity("1.1.1.1", 53)

    t_connect.join()

    return t_connect.packets_captured


@measure_time
def _check_connectivity(host, port=80, timeout=5) -> bool:
    resolver = dns.resolver.get_default_resolver()
    print("~~~CURRENT DNS: ", resolver.nameservers)
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(timeout)
        result = sock.connect_ex((host, port))
        sock.close()
        return result == 0
    finally:
        sock.close()

    return False


def _is_internet_reachable(retry=5) -> bool:
    """Returns True when remote host is reachable by its public IP."""
    i = 0
    while i < retry:
        try:
            #return "icmp_seq=" in sh.ping("-c", "1", "-w", "1", "1.1.1.1")
            return _check_connectivity("1.1.1.1", 53)
        except sh.ErrorReturnCode:
            time.sleep(1)
            i += 1
    return False


def _is_internet_reachable_outside_vpn(retry=5) -> bool:
    """Returns True when remote host is reachable by its public IP outside VPN tunnel."""
    i = 0
    while i < retry:
        try:
            return "icmp_seq=" in sh.sudo.ping("-c", "1", "-m", f"{FWMARK}", "-w", "1", "1.1.1.1")
        except sh.ErrorReturnCode:
            time.sleep(1)
            i += 1
    return False


def _is_ipv6_internet_reachable(retry=5) -> bool:
    i = 0
    last = Exception("_is_ipv6_internet_reachable", "error")
    anycast_ips = cycle(["2400:bb40:4444::100", "2400:bb40:8888::100"])
    while i < retry:
        try:
            return "icmp_seq=" in sh.ping("-c", "1", next(anycast_ips))
        except sh.ErrorReturnCode as e:
            time.sleep(1)
            i += 1
            last = e
    raise last


def _is_dns_resolvable(retry=5) -> bool:
    """Returns True when domain resolution is working."""
    i = 0
    while i < retry:
        try:
            # @TODO gitlab docker runner has public ipv6, but no connectivity. remove -4 once fixed
            #return "icmp_seq=" in sh.ping("-4", "-c", "1", "-w", "1", "nordvpn.com")
            return _check_connectivity("nordvpn.com", 80)
        except sh.ErrorReturnCode:
            time.sleep(1)
            i += 1
    return False


def _is_dns_resolvable_outside_vpn(retry: int = 5) -> bool:
    """Returns True when domain resolution outside vpn is not working."""
    i = 0
    while i < retry:
        try:
            # @TODO gitlab docker runner has public ipv6, but no connectivity. remove -4 once fixed
            # @TODO need dns query to go arround vpn tunnel, here with regular ping it does not
            return "icmp_seq=" in sh.ping("-4", "-c", "1", "-m", f"{FWMARK}", "-w", "1", "nordvpn.com")
        except sh.ErrorReturnCode:
            time.sleep(1)
            i += 1
    return False


def _is_dns_not_resolvable(retry: int = 5) -> bool:
    """Returns True when domain resolution is not working."""
    i = 0
    while i < retry:
        try:
            with pytest.raises((dns.resolver.NoNameservers, dns.resolver.LifetimeTimeout)):
                resolver = dns.resolver.Resolver()
                resolver.lifetime = 1
                resolver.resolve("nordvpn.com")
            return True
        except:  # noqa: E722
            time.sleep(1)
            i += 1
    return False


def is_not_available(retry=5) -> bool:
    """Returns True when network access is not available."""

    if daemon.is_init_systemd():
        sh.sudo("resolvectl", "flush-caches")

    # If assert below fails, and you are running Kill Switch tests on your machine, inside of Docker,
    # set DNS in resolv.conf of your system to anything else but 127.0.0.53
    return not _is_internet_reachable(retry) and _is_dns_not_resolvable(retry)


def is_available(retry=5) -> bool:
    """Returns True when network access is available or throws AssertionError otherwise."""
    assert _is_internet_reachable_outside_vpn(retry)
    assert _is_internet_reachable(retry)
    #assert _is_dns_resolvable_outside_vpn(retry)
    assert _is_dns_resolvable(retry)
    return True


def is_connected() -> bool:
    """Returns True when connected to VPN server or throws AssertionError otherwise."""
    assert daemon.is_connected()
    assert is_available()
    return True


def is_ipv4_and_ipv6_connected(retry=5) -> bool:
    return (
            daemon.is_connected()
            and is_available(retry)
            and _is_ipv6_internet_reachable(retry)
    )


def is_ipv6_connected(retry=5) -> bool:
    return (
            daemon.is_connected()
            and _is_ipv6_internet_reachable(retry)
            and _is_dns_resolvable()
    )


def is_disconnected(retry=5) -> bool:
    """Returns True when not connected to VPN server or throws AssertionError otherwise."""
    assert firewall.is_empty()
    assert daemon.is_disconnected()
    assert is_available(retry)
    return True


# start the networking and wait for completion
def start(default_gateway: str):
    """Must pass default_gateway returned from stop()."""
    if daemon.is_init_systemd():
        sh.sudo.nmcli.networking.on()
    else:
        sh.sudo.ip.link.set.dev.eth0.up()
        cmd = sh.sudo.ip.route.add.default.via.bake(default_gateway)
        cmd.dev.eth0()

    logging.log("starting network")
    while is_not_available():
        time.sleep(1)

    logging.log(info.collect())


# stop the networking and wait for completion
def stop() -> str:
    """Returns default_gateway to be used when starting network again."""
    default_gateway = None
    for line in sh.ip.route().split('\n'):
        if line.startswith('default'):
            default_gateway = line.split()[2]
    assert default_gateway is not None

    if daemon.is_init_systemd():
        sh.sudo.nmcli.networking.off()
    else:
        sh.sudo.ip.link.set.dev.eth0.down()

    logging.log("stopping network")
    assert is_not_available()
    logging.log(info.collect())
    return default_gateway


# block url by domain
def block(url):
    _, _, ips = socket.gethostbyname_ex(url)
    for ip in ips:
        destination = f"{ip}/32"
        if destination not in _blackholes:
            sh.sudo.ip.route.add.blackhole(destination)
            _blackholes.append(destination)


# unblock all blocker domains
def unblock():
    for destination in reversed(_blackholes):
        sh.sudo.ip.route.delete(destination)
        _blackholes.pop()


def get_external_device_ip() -> str:
    """Returns external device IP."""
    return requests.get(API_EXTERNAL_IP, timeout=5).json().get("ip")
