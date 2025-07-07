import datetime
import io
import lib
import subprocess
import signal
import threading
import time
import sh

import dns.resolver
import pytest

import sys
import os

from lib import logging, network, daemon, login, info, firewall

pytest_plugins = ("lib.pytest_timeouts.pytest_timeouts")

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
    if not network.is_internet_reachable(retry=1) or not network.is_internet_reachable_outside_vpn(retry=1):
        print("setup_check_internet_connection: no internet available before running the tests")


@pytest.fixture(scope="session", autouse=True)
def start_system_monitoring():
    print("Run start_system_monitoring")

    # control running threads execution
    stop_event = threading.Event()

    threads = []

    threads.append(threading.Thread(target=_check_connection_to_ip, args=["1.1.1.1", stop_event], daemon=True))
    threads.append(threading.Thread(target=_check_connection_to_ip_outside_vpn, args=["1.1.1.1", stop_event], daemon=True))
    threads.append(threading.Thread(target=_check_dns_resolution, args=["nordvpn.com", stop_event], daemon=True))
    threads.append(threading.Thread(target=_capture_traffic, args=[stop_event], daemon=True))
    print(threads)

    for thread in threads:
        thread.start()

    # execute tests
    yield

    # stop monitoring after execution
    stop_event.set()
    for thread in threads:
        thread.join()

def _check_connection_to_ip(ip_address, stop_event):
    print("Start _check_connection_to_ip")
    while not stop_event.is_set():
        try:
            network.is_internet_reachable(ip_address=ip_address, retry=1)
        except Exception as e: # noqa: BLE001
            print(f"_check_connection_to_ip: FAILURE for {ip_address}: {e}.")
        stop_event.wait(_CHECK_FREQUENCY)


def _check_connection_to_ip_outside_vpn(ip_address, stop_event):
    print("Start _check_connection_to_ip_outside_vpn")
    while not stop_event.is_set():
        try:
            network.is_internet_reachable_outside_vpn(ip_address=ip_address, retry=1)
        except Exception as e: # noqa: BLE001
            print(f"~~~_check_connection_to_ip_outside_vpn: {ip_address} FAILURE: {e}.")
        stop_event.wait(_CHECK_FREQUENCY)


def _check_dns_resolution(domain, stop_event):
    print("Start _check_dns_resolution")
    while not stop_event.is_set():
        try:
            resolver = dns.resolver.Resolver()
            resolver.nameservers = ['8.8.8.8']
            resolver.resolve(domain, 'A')  # 'A' for IPv4
        except Exception as e:  # noqa: BLE001
            print(f"~~~_check_dns_resolution: DNS {domain} FAILURE. Error: {e}")
        stop_event.wait(_CHECK_FREQUENCY)


def _capture_traffic(stop_event):
    print("Start _capture_traffic")
    # use circular log files, keep only 2 latest each 10MB size
    command = ["tshark", "-a", "filesize:10240", "-b", "files:2", "-i", "any", "-w", os.environ["WORKDIR"] + "/dist/logs/tshark_capture.pcap"]
    print("Starting tshark")
    process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    stop_event.wait()
    print("Stopping tshark with Ctrl+C")
    process.send_signal(signal.SIGINT)
    try:
        process.wait(timeout=2)
    except Exception as e:  # noqa: BLE001
        print(f"failed to stop tshark. Error: {e}")
        process.kill()
    print(f"tshark out {process.stdout.read().strip()[-10:]} - {process.stderr.read().strip()[-10:]}")
    time.sleep(1)


@pytest.fixture
def collect_logs():
    """Collect logs."""
    logging.log()

    yield

    logging.log(data=info.collect())
    logging.log()


@pytest.fixture
def nordvpnd_scope_function(collect_logs):  # noqa: ARG001
    """Manage the NordVPN daemon start/stop and login/logout states in a function scope."""
    daemon.start()
    login.login_as("default")

    yield

    sh.nordvpn.set.defaults("--logout", "--off-killswitch")
    daemon.stop()


@pytest.fixture(scope='module')
def nordvpnd_scope_module():
    """Manage the NordVPN daemon start/stop and login/logout states in a module scope."""
    daemon.start()
    login.login_as("default")

    yield

    sh.nordvpn.set.defaults("--logout", "--off-killswitch")
    daemon.stop()


@pytest.fixture(scope="module")
def unblock_network(nordvpnd_scope_module):  # noqa: ARG001
    """Unblocks the network after tests run."""

    yield

    network.unblock()


@pytest.fixture(scope='module')
def add_and_delete_random_route():
    """Add and delete a random network route."""
    firewall.add_and_delete_random_route()


@pytest.fixture
def disable_dns_and_threat_protection():
    """Disable DNS and threat protection settings."""
    lib.set_dns("off")
    lib.set_threat_protection_lite("off")


@pytest.fixture
def disable_notifications():
    """Disable notifications."""
    lib.set_notify("off")
