import ctypes
import datetime
import io
import os
import threading
import time

import dns.resolver
import pytest
import sh
from pyroute2 import IPRoute

from lib import logging, network

_CHECK_FREQUENCY=5

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


def stop_thread(thread):
    if not thread.is_alive():
        #print("Thread is not alive; nothing to stop.")
        return
    tid = thread.ident
    res = ctypes.pythonapi.PyThreadState_SetAsyncExc(
        ctypes.c_long(tid),
        ctypes.py_object(SystemExit)  # Raise SystemExit in the thread
    )
    if res == 0:
        #raise ValueError("Invalid thread ID")
        return

    if res > 1:
        ctypes.pythonapi.PyThreadState_SetAsyncExc(tid, None)
        #raise SystemError("PyThreadState_SetAsyncExc failed")


# Replace the built-in print with our custom version
print = _print_with_timestamp # noqa: A001

@pytest.fixture(scope="function", autouse=True)
def setup_check_internet_connection(request):

    # Get test name and other test information
    test_name = f"{request.module.__name__}-{request.node.name}"
    print(f"~~~TEST_NAME: {test_name}")

    print("~~~setup_check_internet_connection: Check internet connection before starting tests")
    if network.is_available():
        print("~~~setup_check_internet_connection: BEFORE TEST network.is_available SUCCESS")
    else:
        print("~~~setup_check_internet_connection: BEFORE TEST network.is_available FAILURE")

    # we want to capture network traffic during test, outside vpn tunnel and inside tunnel;
    # we can use tshark to capture on multiple interfaces - only need to know what is
    # the tun interface name: nordlynx or nordtun? just capture on `any` interface...
    # but network traffic captured with `any` does not preserve interface info and later
    # when analyzing captured data is not possible to filter by interface :-/
    # also, start capturing for non-existing yet interface - not possible;
    # solution: start netlink event monitor and handle RTM_NEWLINK event.
    #TODO: cleanup this code after test and proof

    # start capture thread
    th1 = threading.Thread(target=monitor_with_netlink, args=(test_name,), daemon=True)
    th1.start()
    th2 = threading.Thread(target=_capture_packets, args=(test_name,), daemon=True)
    th2.start()
    yield # execute test
    time.sleep(2)
    stop_thread(th2)
    stop_thread(th1)


# @pytest.fixture(scope="session", autouse=True)
# def start_system_monitoring():
#     print("~~~start_system_monitoring: Start system monitoring")

#     connection_check_thread = threading.Thread(target=_check_connection_to_ip, args=("1.1.1.1",), daemon=True)
#     connection_out_vpn_check_thread = threading.Thread(target=_check_connection_to_ip_outside_vpn, args=("1.1.1.1",), daemon=True)
#     dns_resolver_thread = threading.Thread(target=_check_dns_resolution, args=("nordvpn.com",), daemon=True)
#     connection_check_thread.start()
#     connection_out_vpn_check_thread.start()
#     dns_resolver_thread.start()

#     yield


def time_str():
    return datetime.datetime.now().strftime("%Y%m%d-%H%M%S")


def _capture_packets(test_name):
    path = f"{os.environ['WORKDIR']}/dist/logs/{test_name}-eth0-{time_str()}"
    # capture traffic on all interfaces and save to file
    sh.dumpcap("-i", "any", "-w", f"{path}.pcap")
    #_capture_packets_on_interface(test_name, "eth0")


def _capture_packets_on_interface(test_name, interface):
    path = f"{os.environ['WORKDIR']}/dist/logs/{test_name}-{interface}-{time_str()}.pcap"
    print(f"~~~_capture_packets_on_interface to file: {path}")
    # capture traffic on given interface and save to file
    #sh.sudo.touch(path)
    sh.dumpcap("-i", interface, "-w", path)



def monitor_with_netlink(test_name):
    ip = IPRoute()
    ip.bind()

    print("~~~Monitoring network events...")
    for msg in ip.get():
        print(f"~~~~~~received msg: {msg}")
        if msg['event'] == 'RTM_NEWLINK':
            interface = msg['attrs'][0][1]
            print(f"~~~~~~~~~New interface detected: {interface}")
            threading.Thread(target=_capture_packets_on_interface, args=(test_name,interface,), daemon=True).start()
            #_capture_packets_on_interface(test_name, interface)

    print("~~~Exit monitoring")



def _check_connection_to_ip(ip_address):
    while True:
        try:
            print(f"~~~_check_connection_to_ip: {ip_address}")
            "icmp_seq=" in sh.ping("-c", "3", "-W", "3", ip_address) # noqa: B015
            print(f"~~~_check_connection_to_ip: IN-PING {ip_address} SUCCESS")
        except sh.ErrorReturnCode as e:
            print(f"~~~_check_connection_to_ip: IN-PING {ip_address} FAILURE: {e}.")
        time.sleep(_CHECK_FREQUENCY)


def _check_connection_to_ip_outside_vpn(ip_address):
    while True:
        try:
            print(f"~~~_check_connection_to_ip_outside_vpn: {ip_address}")
            "icmp_seq=" in sh.sudo.ping("-c", "3", "-W", "3", "-m", "57841", ip_address) # noqa: B015
            print(f"~~~_check_connection_to_ip_outside_vpn: OUT-PING {ip_address} SUCCESS")
        except sh.ErrorReturnCode as e:
            print(f"~~~_check_connection_to_ip_outside_vpn: OUT-PING {ip_address} FAILURE: {e}.")
        time.sleep(_CHECK_FREQUENCY)


def _check_dns_resolution(domain):
    while True:
        try:
            print(f"~~~_check_dns_resolution: {domain}")
            resolver = dns.resolver.Resolver()
            resolver.nameservers = ['8.8.8.8']
            resolver.resolve(domain, 'A')  # 'A' for IPv4
            print(f"~~~_check_dns_resolution: DNS {domain} SUCCESS")
        except Exception as e:  # noqa: BLE001
            print(f"~~~_check_dns_resolution: DNS {domain} FAILURE. Error: {e}")
        time.sleep(_CHECK_FREQUENCY)
