import datetime
import io
import re
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
from constants import NORDVPND, NORDVPND_FILE

pytest_plugins = "lib.pytest_timeouts.pytest_timeouts"

_CHECK_FREQUENCY = 5
RC_TIMEOUT = 1

sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "lib/protobuf/daemon")))


def print_to_string(*args, **kwargs):
    output = io.StringIO()
    _original_print(*args, file=output, **kwargs)
    contents = output.getvalue()
    output.close()
    return contents


_original_print = print


def _print_with_timestamp(*args, **kwargs):
    # Get the current time and format it
    timestamp = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    # Prepend the timestamp to the original print arguments
    _original_print(timestamp, *args, **kwargs)
    logging.log(data=print_to_string(timestamp, *args, **kwargs))


# Replace the built-in print with our custom version
print = _print_with_timestamp  # noqa: A001


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
    threads.append(
        threading.Thread(target=_check_connection_to_ip_outside_vpn, args=["1.1.1.1", stop_event], daemon=True)
    )
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
        except Exception as e:  # noqa: BLE001
            print(f"_check_connection_to_ip: FAILURE for {ip_address}: {e}.")
        stop_event.wait(_CHECK_FREQUENCY)


def _check_connection_to_ip_outside_vpn(ip_address, stop_event):
    print("Start _check_connection_to_ip_outside_vpn")
    while not stop_event.is_set():
        try:
            network.is_internet_reachable_outside_vpn(ip_address=ip_address, retry=1)
        except Exception as e:  # noqa: BLE001
            print(f"~~~_check_connection_to_ip_outside_vpn: {ip_address} FAILURE: {e}.")
        stop_event.wait(_CHECK_FREQUENCY)


def _check_dns_resolution(domain, stop_event):
    print("Start _check_dns_resolution")
    while not stop_event.is_set():
        try:
            resolver = dns.resolver.Resolver()
            resolver.nameservers = ["8.8.8.8"]
            resolver.resolve(domain, "A")  # 'A' for IPv4
        except Exception as e:  # noqa: BLE001
            print(f"~~~_check_dns_resolution: DNS {domain} FAILURE. Error: {e}")
        stop_event.wait(_CHECK_FREQUENCY)


def _capture_traffic(stop_event):
    print("Start _capture_traffic")
    # use circular log files, keep only 2 latest each 10MB size
    command = [
        "tshark",
        "-a",
        "filesize:10240",
        "-b",
        "files:2",
        "-i",
        "any",
        "-w",
        os.environ["WORKDIR"] + "/dist/logs/tshark_capture.pcap",
    ]
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


@pytest.fixture(scope="module")
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


@pytest.fixture(scope="module")
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
def rc_config_manager(default_config: dict) -> RemoteConfigManager:
    """
    Fixture to create and provide an instance of "RemoteConfigManager".

    :param default_config: A dictionary containing the default config values.

    :return: An instance of the "RemoteConfigManager".
    """
    yield RemoteConfigManager(env="dev", cache_dir=LOCAL_CACHE_DIR, default_config=default_config)


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
    print("Clearing local cache file")
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
    except Exception as e:  # noqa: BLE001
        print(f"Got an error during restoring {hosts_original} from {hosts_backup}: {e}")
        raise


@pytest.fixture
def set_custom_timeout_for_rc_retry_scheme(daemon_log_reader):
    """Fixture for setting a custom timeout for the NordVPN daemon's rc retry scheme."""
    print("Setting custom timeout for NordVPN daemon's rc retry scheme")
    daemon_path = NORDVPND_FILE.get(os.environ.get("NORDVPN_TYPE"))
    parameters = ["RC_USE_LOCAL_CONFIG=1", "RC_LOAD_TIME_MIN=4", "IGNORE_HEADER_VALIDATION=1"]

    os.makedirs(f"{os.getcwd()}/tmp/", exist_ok=True)
    subprocess.run(f"sudo cp {daemon_path} {os.getcwd()}/tmp", shell=True, check=True)

    for parameter in parameters:
        if os.environ.get("NORDVPN_TYPE") == "deb":
            # For container
            sed_command = ["sudo", "sed", "-i", f"1a export {parameter}", daemon_path]
        else:
            # For VM
            sed_command = ["sudo", "sed", "-i", f'/$$Service$$/a Environment="{parameter}"']

        subprocess.run(sed_command)

    time_mark = daemon_log_reader.get_cursor()
    if os.environ.get("NORDVPN_TYPE") == "snap":
        subprocess.run(f"sudo sudo systemctl daemon-reload", shell=True, check=True)
        subprocess.run(f'sudo snap restart {NORDVPND.get(os.environ.get("NORDVPN_TYPE"))}', shell=True, check=True)
    else:
        daemon.restart()

    if not daemon_log_reader.wait_for_messages("[Info] remote config download job time period: 4m0s", cursor=time_mark):
        print("Service doesn't applied new time period.")

    yield

    subprocess.run(f'sudo cp {os.getcwd()}/tmp/{daemon_path.split("/")[-1]} {daemon_path}')

    if os.environ.get("NORDVPN_TYPE") == "snap":
        subprocess.run(f"sudo systemctl daemon-reload", shell=True, check=True)
        subprocess.run(f'sudo snap restart {NORDVPND.get(os.environ.get("NORDVPN_TYPE"))}', shell=True, check=True)
    else:
        daemon.restart()


@pytest.fixture
def stop_nordvpnd():
    """Fixture to stop nordvpnd before tests and start it after"""
    daemon.stop()

    yield

    daemon.start()


@pytest.fixture(scope="session", autouse=True)
def get_package_system():
    """Fixture to set in env variable system version of package"""
    try:
        dpkg_result = subprocess.run(["dpkg", "-l", "nordvpn"], capture_output=True, text=True)
        if dpkg_result.returncode == 0 and "nordvpn" in dpkg_result.stdout:
            os.environ["NORDVPN_TYPE"] = "deb"
    except FileNotFoundError:
        # dpkg command not found - might not be a Debian-based system
        pass

    try:
        snap_result = subprocess.run(["snap", "list", "nordvpn"], capture_output=True, text=True)
        if snap_result.returncode == 0 and "nordvpn" in snap_result.stdout:
            os.environ["NORDVPN_TYPE"] = "snap"
    except FileNotFoundError:
        # snap command not found
        pass


@pytest.fixture
def backup_restore_rc_config_files():
    """Fixture to backup original config for remote config, and restore it after tests."""
    os.makedirs(f"{os.getcwd()}/tmp", exist_ok=True)

    subprocess.run(f"sudo cp -r {LOCAL_CACHE_DIR} {os.getcwd()}/tmp", check=True, shell=True)
    yield
    subprocess.run(f"sudo cp -r {os.getcwd()}/tmp {LOCAL_CACHE_DIR} ", check=True, shell=True)
