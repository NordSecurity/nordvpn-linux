from itertools import cycle
from lib import daemon, firewall, info, logging, network
import sh
import socket
import time

# private variable for storing routes
_blackholes = []


def _is_internet_reachable(retry=5) -> bool:
    """returns True when remote host is reachable by it's public IP"""
    i = 0
    while i < retry:
        try:
            return "icmp_seq=" in sh.ping("-c", "1", "-w", "1", "1.1.1.1")
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
    """returns True when domain resolution is working"""
    i = 0
    while i < retry:
        try:
            # @TODO gitlab docker runner has public ipv6, but no connectivity. remove -4 once fixed
            return "icmp_seq=" in sh.ping("-4", "-c", "1", "nordvpn.com")
        except sh.ErrorReturnCode:
            time.sleep(1)
            i += 1
    return False


def is_not_available(retry=5) -> bool:
    try:
        is_available(retry)
        return False
    except AssertionError:
        return True


def is_available(retry=5) -> bool:
    """returns True when network access is available or throws AssertionError otherwise"""
    assert _is_internet_reachable(retry)
    assert _is_dns_resolvable(retry)
    return True


def is_connected() -> bool:
    """returns True when connected to VPN server or throws AssertionError otherwise"""
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
    """returns True when not connected to VPN server or throws AssertionError otherwise"""
    assert firewall.is_empty()
    assert daemon.is_disconnected()
    assert is_available(retry)
    return True


# start the networking and wait for completion
def start():
    if daemon.is_init_systemd():
        sh.sudo.nmcli.networking.on()
    else:
        sh.sudo.ip.link.set.dev.eth0.up()
        cmd = sh.sudo.ip.route.add.default.via.bake("172.17.0.1")
        cmd.dev.eth0()

    logging.log("starting network")
    while not daemon.is_running():
        time.sleep(1)
    logging.log(info.collect())


# stop the networking and wait for completion
def stop():
    if daemon.is_init_systemd():
        sh.sudo.nmcli.networking.off()
    else:
        sh.sudo.ip.link.set.dev.eth0.down()

    logging.log("stopping network")
    assert network.is_not_available()
    logging.log(info.collect())


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
