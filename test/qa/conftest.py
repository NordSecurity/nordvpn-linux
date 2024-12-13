import datetime
import io
import threading
import time

import dns.resolver
import pytest
import sh

import sys
import os

from lib import logging, network

_CHECK_FREQUENCY=5

sys.path.append(os.path.abspath(os.path.join(
    os.path.dirname(__file__), 'lib/protobuf/daemon')))

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
    _original_print(timestamp, *args, **kwargs)
    logging.log(data=print_to_string(timestamp, *args, **kwargs))


# Replace the built-in print with our custom version
print = _print_with_timestamp # noqa: A001

@pytest.fixture(scope="function", autouse=True)
def setup_check_internet_connection():
    print("~~~setup_check_internet_connection: Check internet connection before starting tests")
    if network.is_available():
        print("~~~setup_check_internet_connection: BEFORE TEST network.is_available SUCCESS")
    else:
        print("~~~setup_check_internet_connection: BEFORE TEST network.is_available FAILURE")


@pytest.fixture(scope="session", autouse=True)
def start_system_monitoring():
    print("~~~start_system_monitoring: Start system monitoring")

    # control running threads execution
    stop_event = threading.Event()

    connection_check_thread = threading.Thread(target=_check_connection_to_ip, args=("1.1.1.1", stop_event), daemon=True)
    connection_out_vpn_check_thread = threading.Thread(target=_check_connection_to_ip_outside_vpn, args=("1.1.1.1", stop_event), daemon=True)
    dns_resolver_thread = threading.Thread(target=_check_dns_resolution, args=("nordvpn.com", stop_event), daemon=True)
    connection_check_thread.start()
    connection_out_vpn_check_thread.start()
    dns_resolver_thread.start()

    # execute tests
    yield

    # stop monitoring after execution
    stop_event.set()
    connection_check_thread.join()
    connection_out_vpn_check_thread.join()
    dns_resolver_thread.join()

def _check_connection_to_ip(ip_address, stop_event):
    while not stop_event.is_set():
        try:
            "icmp_seq=" in sh.ping("-c", "3", "-W", "3", ip_address) # noqa: B015
            print(f"~~~_check_connection_to_ip: IN-PING {ip_address} SUCCESS")
        except sh.ErrorReturnCode as e:
            print(f"~~~_check_connection_to_ip: IN-PING {ip_address} FAILURE: {e}.")
            data = "\n".join(["_check_connection_to_ip: Default route:",
                        str(os.popen("sudo ip route get 1.1.1.1").read()),
                        "iptables stats",str(os.popen("sudo iptables -L -v -n").read())])
            logging.log(data=data)
        time.sleep(_CHECK_FREQUENCY)


def _check_connection_to_ip_outside_vpn(ip_address, stop_event):
    while not stop_event.is_set():
        try:
            "icmp_seq=" in sh.sudo.ping("-c", "3", "-W", "3", "-m", "57841", ip_address) # noqa: B015
            print(f"~~~_check_connection_to_ip_outside_vpn: OUT-PING {ip_address} SUCCESS")
        except sh.ErrorReturnCode as e:
            print(f"~~~_check_connection_to_ip_outside_vpn: OUT-PING {ip_address} FAILURE: {e}.")
        time.sleep(_CHECK_FREQUENCY)


def _check_dns_resolution(domain, stop_event):
    while not stop_event.is_set():
        try:
            resolver = dns.resolver.Resolver()
            resolver.nameservers = ['8.8.8.8']
            resolver.resolve(domain, 'A')  # 'A' for IPv4
            print(f"~~~_check_dns_resolution: DNS {domain} SUCCESS")
        except Exception as e:  # noqa: BLE001
            print(f"~~~_check_dns_resolution: DNS {domain} FAILURE. Error: {e}")
        time.sleep(_CHECK_FREQUENCY)
