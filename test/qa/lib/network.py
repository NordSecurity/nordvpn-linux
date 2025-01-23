import socket
import time
from itertools import cycle
from threading import Thread

import dns.resolver
import requests
import sh

from . import daemon, firewall, info, logging, settings

# private variable for storing routes
_blackholes = []

API_EXTERNAL_IP = "https://api.nordvpn.com/v1/helpers/ips/insights"

FWMARK = 57841

class PacketCaptureThread(Thread):
    def __init__(self, connection_settings, duration: int):
        Thread.__init__(self)
        self.connection_settings = connection_settings
        self.packets = ""
        self.duration=duration

    def run(self):
        self.packets = self._capture_packets()

    def _add_filters(self) -> list[str]:
        technology = self.connection_settings[0]
        protocol = self.connection_settings[1]
        obfuscated = self.connection_settings[2]
        server_ip = settings.get_server_ip()

        if technology == "nordlynx":
            return ["-f", f"(udp port 51820) and (host {server_ip})", "-Y", "wg"]

        if technology == "openvpn":
            if obfuscated == "off":
                return ["-f", f"{protocol} and (host {server_ip})", "-Y", "openvpn"]
            if obfuscated =="on":
                return ["-f", f"{protocol} and (host {server_ip})", "-Y", "not openvpn"]

        print("_add_filters: no filters were added")
        return []

    def _capture_packets(self) -> str:
        technology = self.connection_settings[0]

        command = ["-i", "any", "-a", f"duration:{self.duration}"]
        if technology != "" :
            command += self._add_filters()
        logging.log(f"start capturing {command}")
        tshark_result: str = sh.tshark(command)
        # in some cases there will be no output from tshark. This might be a python or tshark problem
        logging.log(f"captured traffic: {tshark_result}")

        return tshark_result.strip()


def capture_traffic(connection_settings, duration: int=5) -> str:
    """Returns count of captured packets."""

    # We try to capture packets using other thread
    t_connect = PacketCaptureThread(connection_settings, duration)
    t_connect.start()

    try:
        # generate some traffic
        generate_traffic(retry=5)
    except Exception as e: # noqa: BLE001
        logging.log(f"capture_traffic exception: {e}")
        logging.log(t_connect.packets)

    t_connect.join()

    return t_connect.packets


def is_internet_reachable(ip_address="1.1.1.1", port=443, retry=5) -> bool:
    """Returns True when remote host is reachable by its public IP."""
    i = 0
    while i < retry:
        try:
            sock = socket.create_connection((ip_address, port), timeout=2)
            sock.close()
            return True
        except Exception as e: # noqa: BLE001
            print(f"is_internet_reachable failed {ip_address}: {e}")
            time.sleep(1)
            i += 1
    return False


def is_internet_reachable_outside_vpn(retry=5, ip_address="1.0.0.1") -> bool:
    """Returns True when remote host is reachable by its public IP outside VPN tunnel."""
    i = 0
    response = ""
    while i < retry:
        try:
            # ping can remain since it is executed with FWMARK
            response = ""
            response = sh.sudo.ping("-c", "1", "-m", f"{FWMARK}", "-w", "1", ip_address)
            return "icmp_seq=" in response
        except sh.ErrorReturnCode as ex:
            print(f"is_internet_reachable_outside_vpn({i}) - failed {ex}: {response}")
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


def _is_dns_resolvable(domain = "nordvpn.com", retry=5) -> bool:
    """Returns True when domain resolution is working."""
    i = 0
    while i < retry:
        try:
            resolver = dns.resolver.Resolver()
            resolver.nameservers = ["103.86.96.100"] # specify server so it will not get the result from the docker host
            resolver.resolve(domain, 'A', lifetime=5)
            return True
        except Exception as e:  # noqa: BLE001
            print(f"_is_dns_resolvable: DNS {domain} FAILURE. Error: {e}")
            time.sleep(1)
            i += 1
    return False


def is_not_available(retry=5) -> bool:
    """Returns True when network access is not available."""

    if daemon.is_init_systemd():
        sh.sudo("resolvectl", "flush-caches")

    # If assert below fails, and you are running Kill Switch tests on your machine, inside of Docker,
    # set DNS in resolv.conf of your system to anything else but 127.0.0.53
    return not is_internet_reachable(retry=retry, ip_address="8.8.8.8") and not _is_dns_resolvable(retry=retry)


def is_available(retry=5) -> bool:
    """Returns True when network access is available or throws AssertionError otherwise."""
    assert is_internet_reachable_outside_vpn(retry=retry)
    assert is_internet_reachable(retry=retry)
    assert _is_dns_resolvable(retry=retry)
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

    logging.log(f"stopping network {default_gateway}")
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


def generate_traffic(retry=1):
    # use an invalid server name to be sure that there will be DNS requests and that the result will not be from OS cache
    _is_dns_resolvable(domain="invalid-server-name.com", retry=retry)
