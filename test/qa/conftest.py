import datetime
import io
import socket
import threading
import time

import pytest
import sh

from lib import logging, network

_CHECK_FREQUENCY=5

#TODO: check connectivity outside VPN tunnel!


def print_to_string(*args, **kwargs):
    output = io.StringIO()
    _original_print(*args, file=output, **kwargs)
    contents = output.getvalue()
    output.close()
    return contents


_original_print = print
def _print_with_timestamp(*args, **kwargs):
    # Get the current time and format it
    timestamp = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    # Prepend the timestamp to the original print arguments
    #_original_print(timestamp, *args, **kwargs)
    logging.log(data=print_to_string(timestamp, *args, **kwargs))


# Replace the built-in print with our custom version
print = _print_with_timestamp # noqa: A001

@pytest.fixture(scope="function", autouse=True)
def setup_check_internet_connection():
    print("~~~setup_check_internet_connection: Check internet connection before starting tests")
    assert network.is_available()
    print("~~~setup_check_internet_connection: Check internet connection before starting tests SUCCESS")


@pytest.fixture(scope="session", autouse=True)
def start_system_monitoring():
    print("~~~start_system_monitoring: Start system monitoring")

    connection_check_thread = threading.Thread(target=_check_connection_to_ip, args=("1.1.1.1",), daemon=True)
    connection_out_vpn_check_thread = threading.Thread(target=_check_connection_to_ip_outside_vpn, args=("1.1.1.1",), daemon=True)
    dns_resolver_thread = threading.Thread(target=_check_dns_resolution, args=("nordvpn.com",), daemon=True)
    connection_check_thread.start()
    connection_out_vpn_check_thread.start()
    dns_resolver_thread.start()

    yield


def _check_connection_to_ip(ip_address):
    while True:
        try:
            print(f"~~~_check_connection_to_ip: {ip_address}")
            "icmp_seq=" in sh.ping("-c", "1", "-w", "1", ip_address) # noqa: B015
            print(f"~~~_check_connection_to_ip: {ip_address} SUCCESS")
        except sh.ErrorReturnCode as e:
            print(f"~~~_check_connection_to_ip: Failed to connect to {ip_address}: {e}.")
        time.sleep(_CHECK_FREQUENCY)


def _check_connection_to_ip_outside_vpn(ip_address):
    while True:
        try:
            print(f"~~~_check_connection_to_ip_outside_vpn: {ip_address}")
            "icmp_seq=" in sh.ping("-c", "1", "-w", "1", "-I", "eth0", ip_address) # noqa: B015
            print(f"~~~_check_connection_to_ip_outside_vpn: {ip_address} SUCCESS")
        except sh.ErrorReturnCode as e:
            print(f"~~~_check_connection_to_ip_outside_vpn: Failed to connect to {ip_address}: {e}.")
        time.sleep(_CHECK_FREQUENCY)


def _check_dns_resolution(domain):
    while True:
        try:
            print(f"~~~_check_dns_resolution: {domain}")
            socket.gethostbyname(domain)
            print(f"~~~_check_dns_resolution: {domain} SUCCESS")
        except socket.gaierror:
            print(f"~~~_check_dns_resolution: DNS resolution for {domain} failed.")
        time.sleep(_CHECK_FREQUENCY)
