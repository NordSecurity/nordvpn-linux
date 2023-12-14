from itertools import cycle
from lib import daemon, firewall, info, logging, network
import pytest
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


def _is_dns_not_resolvable(retry=5) -> bool:
    """returns True when domain resolution is not working"""
    for i in range(retry):
        try:
            with pytest.raises(sh.ErrorReturnCode_2) as ex:
                sh.ping("-4", "-c", "1", "nordvpn.com")

            return "Network is unreachable" in str(ex) or \
                   "Name or service not known" in str(ex) or \
                   "Temporary failure in name resolution" in str(ex)
        except Exception as ex:
            time.sleep(1)
    return False


def is_not_available(retry=5) -> bool:
    """ returns True when network access is not available """
    # If assert below fails, and you are running Kill Switch tests on your machine, inside of Docker,
    # set DNS in resolv.conf of your system to anything else but 127.0.0.53
    return not _is_internet_reachable(retry) and _is_dns_not_resolvable(retry)


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
def start(default_gateway: str):
    '''Must pass default_gateway returned from stop()'''
    if daemon.is_init_systemd():
        sh.sudo.nmcli.networking.on()
    else:
        sh.sudo.ip.link.set.dev.eth0.up()
        cmd = sh.sudo.ip.route.add.default.via.bake(default_gateway)
        cmd.dev.eth0()

    logging.log("starting network")
    while not daemon.is_running():
        time.sleep(1)
    logging.log(info.collect())


# stop the networking and wait for completion
def stop() -> str:
    '''Returns default_gateway to be used when starting network again'''
    for line in sh.ip.route().split('\n'):
        if line.startswith('default'):
            default_gateway = line.split()[2]
    assert default_gateway is not None

    if daemon.is_init_systemd():
        sh.sudo.nmcli.networking.off()
    else:
        sh.sudo.ip.link.set.dev.eth0.down()

    logging.log("stopping network")
    assert network.is_not_available()
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
