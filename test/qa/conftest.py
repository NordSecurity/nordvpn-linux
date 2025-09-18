import datetime
import io
from urllib.parse import urlparse

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

from lib.remote_config_manager import RemoteConfigManager, LOCAL_CACHE_DIR, REMOTE_DIR
from lib.logging import FILE
from lib.log_reader import LogReader

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


@pytest.fixture
def default_config() -> dict:
    """Fixture to provide a default config."""
    return dict()


@pytest.fixture
def rc_config_manager(default_config: dict
) -> RemoteConfigManager:
    """
    Fixture to create and provide an instance of "RemoteConfigManager".

    :param default_config: A dictionary containing the default config values.

    :return: An instance of the "RemoteConfigManager".
    """
    return RemoteConfigManager(
        env="dev",
        cache_dir=LOCAL_CACHE_DIR,
        default_config=default_config
    )


@pytest.fixture
def daemon_log_reader() -> LogReader:
    """
    Fixture to provide an instance of the "LogReader" class.

    :return: an instance of the "LogReader" class with the default log file path.
    """
    return LogReader(FILE)


@pytest.fixture
def daemon_log_cursor(daemon_log_reader) -> int:
    """
    Fixture to get the current cursor for the log file.

    Uses the "LogReader" class to determine the current EOF cursor.
    If the log file is not found or empty, the cursor is set to 0.

    :param daemon_log_reader: An instance of the "LogReader" class.

    :return: An integer cursor representing the current end of the log file..
    """
    cursor = daemon_log_reader.get_cursor()

    print(f"Cursor: {cursor}")
    return cursor


@pytest.fixture
def clean_cache_files(rc_config_manager):
    """
    Fixture to clean up the files in "LOCAL_CACHE_DIR" directory before running tests.

    :param rc_config_manager: An instance of the "RemoteConfigManager" class.
    """
    if rc_config_manager.set_permissions_cache_dir():
        subprocess.run(["sudo", "rm", "-rf", LOCAL_CACHE_DIR], check=True)
    else:
        print(f"Unable to clean {LOCAL_CACHE_DIR}, as it does not exist.")

    yield


@pytest.fixture
def disable_remote_endpoint():
    """
    Fixture to temporarily disable connections to "remote endpoint" by modifying "/etc/hosts".

    This fixture performs the following steps:
    1. Backs up the current "/etc/hosts" file to "/etc/hosts.bak".
    2. Appends the following entry "0.0.0.0 %remote endpoint%" to "/etc/hosts" to block "remote endpoint".
        This ensures the application cannot reach this endpoint (e.g., to download remote config files).
    3. Restores the original "/etc/hosts" file after the test completes.

    :raises ValueError: If REMOTE_DIR is not set.
    """
    if not REMOTE_DIR:
        raise ValueError("REMOTE_DIR environment variable is not set.")
    domain = urlparse(REMOTE_DIR).netloc

    hosts_original = "/etc/hosts"
    hosts_backup = "/etc/hosts.bak"

    subprocess.run(["sudo", "cp", hosts_original, hosts_backup], check=True)
    print(f"Backup created at {hosts_backup}")

    print(f"Adding '0.0.0.0 {domain}' to {hosts_original}")
    cmd = ["sudo", "bash", "-c", f'echo "0.0.0.0 {domain}" >> /etc/hosts']
    subprocess.run(cmd, check=True)

    yield

    try:
        subprocess.run(["sudo", "cp", hosts_backup, hosts_original], check=True)
        subprocess.run(["sudo", "chmod", "644", hosts_original], check=True)
        print(f"{hosts_original} restored from {hosts_backup}.")
    except Exception as e: # noqa: BLE001
        print(f"Got an error during restoring {hosts_original} from {hosts_backup}: {e}")
        raise


@pytest.fixture(scope="session")
def env():
    """Detects and returns the active environment (DEV or PROD) based on the NordVPN version output."""
    env = daemon.get_env()
    print(f"Current env: '{env}'")
    return env
