import pytest

import threading
import sh
import socket
import time
import datetime
from lib import network

_CHECK_FREQUENCY=5

_original_print = print
def _print_with_timestamp(*args, **kwargs):
    # Get the current time and format it
    timestamp = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    # Prepend the timestamp to the original print arguments
    _original_print(timestamp, *args, **kwargs)

# Replace the built-in print with our custom version
print = _print_with_timestamp

@pytest.fixture(scope="function", autouse=True)
def setup_check_internet_connection():
    print("Check internet connection before starting tests")
    assert network.is_available()
    

@pytest.fixture(scope="session", autouse=True)
def start_system_monitoring():
    print("Start system monitoring")
    
    connection_check_thread = threading.Thread(target=_check_connection_to_ip, args=("1.1.1.1",), daemon=True)
    dns_resolver_thread = threading.Thread(target=_check_dns_resolution, args=("nordvpn.com",), daemon=True)
    connection_check_thread.start()
    dns_resolver_thread.start()
    
    yield


def _check_connection_to_ip(ip_address):
    while True:
        try:
            "icmp_seq=" in sh.ping("-c", "1", "-w", "1", ip_address)
        except sh.ErrorReturnCode as e:
            print(f"Failed to connect to {ip_address}: {e}.")
        time.sleep(_CHECK_FREQUENCY)      
        
def _check_dns_resolution(domain):
    while True:
        try:
            socket.gethostbyname(domain)
        except socket.gaierror:
            print(f"DNS resolution for {domain} failed.")
        time.sleep(_CHECK_FREQUENCY)