"""Functions to make it easier to interact with nordvpnd."""

import glob
import os
import socket
import subprocess
import time

import sh

from . import logging, ssh
from lib.shell import sh_no_tty


class Env:
    DEV = "dev"
    PROD = "prod"


def _rewrite_log_path():
    project_root = os.environ["WORKDIR"].replace("/", "\\/")
    pattern = f"s/^LOGFILE=.*/LOGFILE={project_root}\\/dist\\/logs\\/daemon.log/"
    # this fn is executed only in docker (below line would not work under snap)
    sh.sudo.sed("-i", pattern, "/etc/init.d/nordvpn")


# returns True on SystemD distros
def is_init_systemd():
    return "systemd" in sh.ps("--no-headers", "-o", "comm", "1")


def is_under_snap():
    return "snap" in sh.which("nordvpn")


def is_connected() -> bool:
    """Returns True when connected to VPN server."""
    try:
        return "Connected" in sh.nordvpn.status()
    except sh.ErrorReturnCode:
        return False


def is_disconnected() -> bool:
    """Returns True when not connected to VPN."""
    try:
        print(sh_no_tty.nordvpn.status())
        return "Disconnected" in sh.nordvpn.status()
    except sh.ErrorReturnCode as ex:
        logging.log(data=f"is_disconnected: {ex}")
        return True


def is_killswitch_on():
    """Return True when Killswitch is activated."""
    try:
        return "Kill Switch: enabled" in sh.nordvpn.settings()
    except sh.ErrorReturnCode:
        return False


def install_peer(ssh_client: ssh.Ssh):
    """Installs nordvpn in peer."""
    project_root = os.environ["WORKDIR"]
    deb_path = glob.glob(f"{project_root}/dist/app/deb/*amd64.deb")[0]
    ssh_client.send_file(deb_path, "/tmp/nordvpn.deb")
    # TODO: Install required dependencies during qa-peer image build, then replace with 'dpkg -i /tmp/nordvpn.deb'
    ssh_client.exec_command("sudo apt-get update")
    ssh_client.exec_command("sudo apt install -y /tmp/nordvpn.deb")


def uninstall_peer(ssh_client: ssh.Ssh):
    """Uninstalls nordvpn in peer."""
    # TODO: Replace with 'dpkg --purge nordvpn'
    ssh_client.exec_command("sudo apt remove -y nordvpn")


def start():
    """Starts daemon and blocks until it is actually started."""
    if is_under_snap():
        # sh.sudo.snap("start", "nordvpn")
        os.popen("sudo snap start nordvpn").read()
    elif is_init_systemd():
        # sh.sudo.systemctl.start.nordvpnd()
        os.popen("sudo systemctl start nordvpn").read()
    else:
        # call to init.d returns before the daemon is actually started
        _rewrite_log_path()
        if is_running():
            print("Nordvpn is already running")
        try:
            print("Starting NordVPN service...")
            # Use the init.d script as recommended in the error message
            result = subprocess.run(
                ["sudo", "/etc/init.d/nordvpn", "start"], capture_output=True, text=True, check=False
            )

            if result.returncode != 0:
                print(f"Error starting NordVPN service: {result.stderr}")

            # Wait for the socket file to appear
            max_wait = 30  # seconds
            start_time = time.time()
            socket_path = "/run/nordvpn/nordvpnd.sock"

            while not os.path.exists(socket_path) and (time.time() < start_time + max_wait):
                print(f"Waiting for socket file {socket_path}...")
                time.sleep(1)

            if not os.path.exists(socket_path):
                print(f"Socket file {socket_path} not created within timeout")

            print("NordVPN service started successfully")

        except (subprocess.SubprocessError, FileNotFoundError, TimeoutError) as e:
            print(f"Exception while starting NordVPN service: {e}")
    while not is_running():
        time.sleep(1)


def start_peer(ssh_client: ssh.Ssh):
    """Starts daemon in peer and blocks until it is actually started."""
    ssh_client.exec_command("sudo /etc/init.d/nordvpn start")
    time.sleep(1)
    while not is_peer_running(ssh_client):
        time.sleep(1)


def stop():
    """Stops the daemon and blocks until it is actually stopped."""
    if is_under_snap():
        # sh.sudo.snap("stop", "nordvpn")
        os.popen("sudo snap stop nordvpn").read()
    elif is_init_systemd():
        # sh.sudo.systemctl.stop.nordvpnd()
        os.popen("sudo systemctl stop nordvpn").read()
    else:
        # call to init.d returns before the daemon is actually stopped
        try:
            sh.sudo("/etc/init.d/nordvpn", "stop")
        except (subprocess.SubprocessError, FileNotFoundError, TimeoutError):
            if is_running():
                print("Service still running, trying sudo pkill...")
                subprocess.run(["sudo", "pkill", "nordvpnd"], capture_output=True, check=False)
            time.sleep(3)
            if is_running():
                print("Service still running, trying sudo pkill -9 ...")
                subprocess.run(["sudo", "pkill", "-9", "nordvpnd"], capture_output=True, check=False)
            time.sleep(3)
            if is_running():
                print("Failed to stop nordvpn")
        while is_running():
            time.sleep(1)


def stop_peer(ssh_client: ssh.Ssh):
    """Stops the daemon in peer and blocks until it is actually stopped."""
    ssh_client.exec_command("sudo /etc/init.d/nordvpn stop")
    while is_peer_running(ssh_client):
        time.sleep(1)


# restarts the daemon and blocks until it is actually restarted
def restart():
    try:
        stop()
        if is_init_systemd():
            # The default limit is to allow 5 restarts in a 10sec period.
            # There are 100ms intervals between restarts.
            time.sleep(2)
        start()
    except sh.ErrorReturnCode as ex:
        print(f"Got next error: {ex}")


# retrieving links inside this function creates a race condition,
# therefore it is safer to provide them as arguments
def wait_for_reconnect(links: list[tuple[int, str]]):
    logging.log("waiting for reconnect")
    while True:
        got = socket.if_nameindex()
        if len(got) != len(links):  # old tunnel is gone
            continue
        if got != links:  # new tunnel appeared
            logging.log(got)
            if list(filter(lambda x: "nordvpn-wg" in x, got)):
                # not yet connected to actual VPN, this is just a test interface
                continue
            time.sleep(2)
            break


def wait_for_autoconnect():
    while not is_connected():
        time.sleep(1)


# returns True when daemon is running
def is_running():
    try:
        sh.nordvpn.status()
    except sh.ErrorReturnCode_1 as ex:
        # if user is not part of the nordvpn group assert
        assert "Permission needed" not in ex.stdout.decode()
        return False
    else:
        return True


# returns True when daemon is running in peer
def is_peer_running(ssh_client: ssh.Ssh) -> bool:
    try:
        ssh_client.exec_command("nordvpn status")
        return True
    except Exception:  # noqa: BLE001
        return False


def get_unavailable_groups():
    """Returns groups that are not available with current connection settings."""
    all_groups = [
        "Africa_The_Middle_East_And_India",
        "Asia_Pacific",
        "Dedicated_IP",
        "Double_VPN",
        "Europe",
        "Obfuscated_Servers",
        "Onion_Over_VPN",
        "P2P",
        "Standard_VPN_Servers",
        "The_Americas",
    ]

    current_groups = str(sh.nordvpn.groups(_tty_out=False)).strip().split()

    return set(all_groups) - set(current_groups)


def get_status_data() -> dict:
    lines = sh.nordvpn.status(_tty_out=False).strip().split("\n")
    colon_separated_pairs = (element.split(":") for element in lines)
    formatted_pairs = {(key.lower(), value.strip()) for key, value in colon_separated_pairs}
    return dict(formatted_pairs)


def get_env() -> str:
    """
    Detects and returns the active environment (DEV or PROD) based on the NordVPN version output.

    :return: the active environment (DEV or PROD)
    """
    result = sh.nordvpn("--version")
    if Env.DEV in result:
        return Env.DEV
    return Env.PROD

def enable_rc_local_config_usage():
    """
    Modifies the nordvpn service file to enable usage of local remote config.

    This function is typically used to force the application to use locally cached remote config files.
    """
    service_path = "/etc/init.d/nordvpn"

    # Print original service file for reference
    print(f"Service original:\n {sh.cat(service_path)}")

    # Insert environment variable into systemd service file
    sh.sudo("sed", "-i", r"1a export RC_USE_LOCAL_CONFIG=1", service_path)

    sh.sudo.systemctl("daemon-reload", _ok_code=(0, 1))

    # Print service file after modification
    print(f"Service after:\n {sh.cat(service_path)}")

def disable_rc_local_config_usage():
    """
    Restores the original nordvpn service file by removing the forced local config env variable.

    This function is typically used to clean up the service file after testing,
    ensuring that the local remote config behavior is disabled.
    """
    service_path = "/etc/init.d/nordvpn"
    # Cleanup: remove the injected environment variable line
    sh.sudo("sed", "-i", r"/^export RC_USE_LOCAL_CONFIG=1$/d", service_path)

    sh.sudo.systemctl("daemon-reload", _ok_code=(0, 1))

    # Print restored service file for verification
    print(f"Service restored:\n {sh.cat(service_path)}")
